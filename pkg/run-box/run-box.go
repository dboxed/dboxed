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

	box_spec "github.com/dboxed/dboxed/pkg/box-spec"
	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/sandbox"
	"github.com/dboxed/dboxed/pkg/selfupdate"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/gofrs/flock"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"
	"go4.org/netipx"
)

type RunBox struct {
	Debug bool

	BoxUrl          *url.URL
	Nkey            string
	BoxName         string
	WorkDir         string
	VethNetworkCidr string

	natsConn *nats.Conn

	acquiredVethNetworkCidr *net.IPNet

	boxSpec *types.BoxSpec

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
	err = os.MkdirAll(filepath.Join(sandboxDir, "docker"), 0700)
	if err != nil {
		return err
	}

	err = rn.initFileLogging(ctx, sandboxDir)
	if err != nil {
		return err
	}

	if os.Getpid() == 1 {
		slog.InfoContext(ctx, "evacuating cgroup2")
		err := cgrouputil.EvacuateCgroup2("init")
		if err != nil {
			return fmt.Errorf("failed to evacuate root cgroup: %w", err)
		}
	}

	err = rn.connectNats(ctx)
	if err != nil {
		return err
	}

	var boxSpecSource *box_spec.BoxSpecSource
	if rn.natsConn != nil {
		bucket := rn.BoxUrl.Query().Get("bucket")
		if bucket == "" {
			return fmt.Errorf("missing bucket in nats url")
		}
		key := rn.BoxUrl.Query().Get("key")
		if key == "" {
			return fmt.Errorf("missing key in nats url")
		}
		boxSpecSource, err = box_spec.NewNatsSource(ctx, rn.natsConn, bucket, key)
		if err != nil {
			return err
		}
	} else {
		boxSpecSource, err = box_spec.NewUrlSource(ctx, *rn.BoxUrl)
		if err != nil {
			return err
		}
	}

	rn.boxSpec = &boxSpecSource.GetCurSpec().Spec

	// we should start publishing logs asap, but the earliest point at the moment is after box spec retrieval
	// (we need nats credentials)
	err = rn.initLogsPublishing(ctx, sandboxDir)
	if err != nil {
		return err
	}
	defer rn.logsPublisher.Stop()

	rn.sandbox = &sandbox.Sandbox{
		Debug:       rn.Debug,
		HostWorkDir: rn.WorkDir,
		SandboxName: rn.BoxName,
		SandboxDir:  sandboxDir,
		BoxSpec:     rn.boxSpec,
	}

	err = rn.sandbox.Destroy(ctx)
	if err != nil {
		return err
	}

	err = selfupdate.SelfUpdateIfNeeded(ctx, rn.boxSpec.DboxedBinaryUrl, rn.boxSpec.DboxedBinaryHash, rn.WorkDir)
	if err != nil {
		return err
	}

	err = rn.reserveVethCIDR(ctx)
	if err != nil {
		return err
	}
	rn.sandbox.VethNetworkCidr = rn.acquiredVethNetworkCidr
	slog.InfoContext(ctx, "using veth cidr", slog.Any("cidr", rn.acquiredVethNetworkCidr.String()))

	err = rn.sandbox.Prepare(ctx)
	if err != nil {
		return err
	}
	err = rn.sandbox.SetupNetworking(ctx)
	if err != nil {
		return err
	}

	err = rn.sandbox.Start(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "up and running")

	<-ctx.Done()

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
