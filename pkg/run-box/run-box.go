package run_box

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed-common/util"
	box_spec_runner "github.com/dboxed/dboxed/pkg/box-spec-runner"
	"github.com/dboxed/dboxed/pkg/box-spec-runner/source"
	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/sandbox"
	"github.com/dboxed/dboxed/pkg/selfupdate"
	"github.com/dboxed/dboxed/pkg/types"
	util2 "github.com/dboxed/dboxed/pkg/util"
	"github.com/gofrs/flock"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"go4.org/netipx"
)

type RunBox struct {
	Debug bool

	InfraImage      string
	BoxUrl          *url.URL
	Nkey            string
	BoxName         string
	WorkDir         string
	VethNetworkCidr string

	natsConn *nats.Conn

	acquiredVethNetworkCidr *net.IPNet

	boxSpecSource *source.BoxSpecSource

	logsPublisher logs.LogsPublisher
	sandbox       *sandbox.Sandbox
}

func (rn *RunBox) Run(ctx context.Context) error {
	sandboxDir := filepath.Join(rn.WorkDir, "boxes", rn.BoxName)

	err := os.MkdirAll(rn.WorkDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(sandboxDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(sandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}

	err = rn.initFileLogging(ctx, sandboxDir)
	if err != nil {
		return err
	}

	err = rn.connectNats(ctx)
	if err != nil {
		return err
	}

	err = rn.startBoxSpecSource(ctx)
	if err != nil {
		return err
	}
	initialBoxSpec := rn.boxSpecSource.GetCurSpec().Spec

	// we should start publishing logs asap, but the earliest point at the moment is after box spec retrieval
	err = rn.initLogsPublishing(ctx, sandboxDir, &initialBoxSpec)
	if err != nil {
		return err
	}
	defer rn.logsPublisher.Stop()

	err = selfupdate.SelfUpdateIfNeeded(ctx, initialBoxSpec.DboxedBinaryUrl, initialBoxSpec.DboxedBinaryHash, rn.WorkDir)
	if err != nil {
		return err
	}

	err = rn.reserveVethCIDR(ctx)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "using veth cidr", slog.Any("cidr", rn.acquiredVethNetworkCidr.String()))

	rn.sandbox = &sandbox.Sandbox{
		Debug:           rn.Debug,
		InfraImage:      rn.InfraImage,
		HostWorkDir:     rn.WorkDir,
		SandboxName:     rn.BoxName,
		SandboxDir:      sandboxDir,
		VethNetworkCidr: rn.acquiredVethNetworkCidr,
	}

	needDestroy := false

	localUuid, err := rn.readBoxUuid(rn.BoxName)
	if err != nil {
		return err
	}
	if localUuid != initialBoxSpec.Uuid {
		if localUuid != "" {
			slog.InfoContext(ctx, fmt.Sprintf("serving a new box (new uuid %s), destroying old one (uuid %s)", initialBoxSpec.Uuid, localUuid))
		}
		needDestroy = true
	}

	if !needDestroy {
		runcState, err := rn.sandbox.RuncState(ctx)
		if err != nil {
			return err
		}

		if runcState == nil {
			needDestroy = true
		} else {
			if runcState.Status != "running" {
				slog.InfoContext(ctx, fmt.Sprintf("old sandbox container is in state '%s', re-creating it", runcState.Status))
				needDestroy = true
			}
		}
	}

	if needDestroy {
		err = rn.sandbox.Destroy(ctx)
		if err != nil {
			return err
		}
		err = rn.sandbox.Prepare(ctx)
		if err != nil {
			return err
		}
	}

	err = rn.writeBoxUuid(initialBoxSpec.Uuid)
	if err != nil {
		return err
	}

	err = rn.sandbox.SetupNetworking(ctx)
	if err != nil {
		return err
	}

	util2.LoadMod(ctx, "dm-mod")
	util2.LoadMod(ctx, "dm-thin-pool")
	util2.LoadMod(ctx, "dm-snapshot")
	util2.LoadMod(ctx, "dm-zero")

	// now that we ensured that the potentially running sandbox does not belong to another box, we can start publishing
	// the sandbox internal logs
	err = rn.initLogsPublishingSandbox(ctx, sandboxDir, &initialBoxSpec)
	if err != nil {
		return err
	}

	if needDestroy {
		err = rn.sandbox.Start(ctx)
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "waiting for docker to become available in the sandbox")
	for {
		_, err = rn.sandbox.RunDockerCli(ctx, slog.Default(), true, "", "info")
		if err == nil {
			break
		}
		if !util.SleepWithContext(ctx, 2*time.Second) {
			return ctx.Err()
		}
	}
	slog.InfoContext(ctx, "docker is up and running")

	for {
		boxSpec := rn.boxSpecSource.GetCurSpec().Spec

		boxSpecRunner := &box_spec_runner.BoxSpecRunner{
			Sandbox: rn.sandbox,
			BoxSpec: &boxSpec,
		}
		err := boxSpecRunner.Reconcile(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error while reconciling box spec", slog.Any("error", err))
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-rn.boxSpecSource.Chan:
			if !ok {
				slog.InfoContext(ctx, "box spec was deleted, downing and exiting")
				err = boxSpecRunner.Down(ctx)
				if err != nil {
					return err
				}
				err = rn.sandbox.Stop(ctx)
				if err != nil {
					return err
				}
				return nil
			}
			slog.InfoContext(ctx, "a new box spec was received")
			continue
		}
	}
}

func (rn *RunBox) startBoxSpecSource(ctx context.Context) error {
	var err error
	if rn.natsConn != nil {
		bucket := rn.BoxUrl.Query().Get("bucket")
		if bucket == "" {
			return fmt.Errorf("missing bucket in nats url")
		}
		key := rn.BoxUrl.Query().Get("key")
		if key == "" {
			return fmt.Errorf("missing key in nats url")
		}
		rn.boxSpecSource, err = source.NewNatsSource(ctx, rn.natsConn, bucket, key)
		if err != nil {
			return err
		}
	} else {
		rn.boxSpecSource, err = source.NewUrlSource(ctx, *rn.BoxUrl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *RunBox) connectNats(ctx context.Context) error {
	if rn.BoxUrl.Scheme != "nats" {
		return nil
	}

	token := rn.BoxUrl.Query().Get("token")

	var opts []nats.Option
	if rn.Nkey != "" {
		nkeySeed, err := os.ReadFile(rn.Nkey)
		if err != nil {
			return err
		}

		kp, err := nkeys.FromSeed(nkeySeed)
		if err != nil {
			return err
		}
		nkey, err := kp.PublicKey()
		if err != nil {
			return err
		}
		opts = append(opts, nats.Nkey(nkey, kp.Sign))
	} else if token != "" {
		opts = append(opts, nats.Token(token))
	}

	slog.InfoContext(ctx, "connecting to nats",
		slog.Any("url", rn.BoxUrl.String()),
	)
	nc, err := nats.Connect(rn.BoxUrl.String(), opts...)
	if err != nil {
		return err
	}

	rn.natsConn = nc
	return err
}

func (rn *RunBox) reserveVethCIDR(ctx context.Context) error {
	slog.InfoContext(ctx, "reserving CIDR for veth pair")

	fl := flock.New(filepath.Join(rn.WorkDir, "veth-cidrs.lock"))
	err := fl.Lock()
	if err != nil {
		return err
	}
	defer fl.Unlock()

	p, err := rn.readVethCidr(rn.BoxName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		_, rn.acquiredVethNetworkCidr, _ = net.ParseCIDR(p.String())
		return nil
	}

	otherIpsNets, err := rn.readReservedIPs()
	if err != nil {
		return err
	}

	pc, err := netip.ParsePrefix(rn.VethNetworkCidr)
	if err != nil {
		return err
	}
	pr := netipx.RangeOfPrefix(pc)
	if !pr.IsValid() {
		return fmt.Errorf("invalid cidr")
	}

	var b netipx.IPSetBuilder
	b.AddRange(pr)
	for _, op := range otherIpsNets {
		b.RemovePrefix(op)
	}

	ips, err := b.IPSet()
	if err != nil {
		return err
	}
	newPrefix, _, ok := ips.RemoveFreePrefix(30)
	if !ok {
		return fmt.Errorf("failed to reserve veth pair CIDR")
	}

	err = rn.writeVethCidr(&newPrefix)
	if err != nil {
		return err
	}

	_, rn.acquiredVethNetworkCidr, _ = net.ParseCIDR(newPrefix.String())

	return nil
}

func (rn *RunBox) readVethCidr(boxName string) (*netip.Prefix, error) {
	pth := filepath.Join(rn.WorkDir, "boxes", boxName, types.VethIPStoreFile)
	ipB, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}
	p, err := netip.ParsePrefix(string(ipB))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (rn *RunBox) writeVethCidr(p *netip.Prefix) error {
	pth := filepath.Join(rn.WorkDir, "boxes", rn.BoxName, types.VethIPStoreFile)
	return util.AtomicWriteFile(pth, []byte(p.String()), 0644)
}

func (rn *RunBox) readReservedIPs() ([]netip.Prefix, error) {
	boxesDir := filepath.Join(rn.WorkDir, "boxes")
	des, err := os.ReadDir(boxesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	var ret []netip.Prefix
	for _, de := range des {
		if !de.IsDir() {
			continue
		}
		p, err := rn.readVethCidr(de.Name())
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			ret = append(ret, *p)
		}
	}
	return ret, nil
}

func (rn *RunBox) readBoxUuid(boxName string) (string, error) {
	pth := filepath.Join(rn.WorkDir, "boxes", boxName, types.BoxSpecUuidFile)
	b, err := os.ReadFile(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func (rn *RunBox) writeBoxUuid(uuid string) error {
	pth := filepath.Join(rn.WorkDir, "boxes", rn.BoxName, types.BoxSpecUuidFile)
	return util.AtomicWriteFile(pth, []byte(uuid), 0644)
}
