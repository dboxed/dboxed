//go:build linux

package run_box

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/runner/selfupdate"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/gofrs/flock"
	"go4.org/netipx"
	"sigs.k8s.io/yaml"
)

type RunBox struct {
	Debug bool

	Client *baseclient.Client
	BoxId  int64

	InfraImage      string
	BoxName         string
	WorkDir         string
	VethNetworkCidr string

	acquiredVethNetworkCidr *net.IPNet

	logsPublisher logs.LogsPublisher
	sandbox       *sandbox.Sandbox
}

func (rn *RunBox) Run(ctx context.Context) error {
	if rn.Client.GetClientAuth().StaticToken == nil {
		return fmt.Errorf("can only run box with static token")
	}

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

	boxesClient := clients.BoxClient{Client: rn.Client}
	initialBoxFile, err := boxesClient.GetBoxSpecById(ctx, rn.BoxId)
	if err != nil {
		return err
	}
	initialBoxSpec := &initialBoxFile.Spec

	// we should start publishing logs asap, but the earliest point at the moment is after box spec retrieval
	err = rn.initLogsPublishing(ctx, sandboxDir)
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

	err = rn.sandbox.PrepareNetworkingConfig()
	if err != nil {
		return err
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
		// we must run this after Prepare as it will need networking tools in rootfs
		err = rn.sandbox.DestroyNetworking(ctx)
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, "stopping dboxed service inside sandbox")
		err = rn.sandbox.S6SvcDown(ctx, "dboxed")
		if err != nil {
			return err
		}
	}

	err = rn.sandbox.CopyBinaries(ctx)
	if err != nil {
		return err
	}

	err = rn.writeBoxUuid(initialBoxSpec.Uuid)
	if err != nil {
		return err
	}

	err = rn.sandbox.SetupNetworking(ctx)
	if err != nil {
		return err
	}

	// now that we ensured that the potentially running sandbox does not belong to another box, we can start publishing
	// the sandbox internal logs
	err = rn.initLogsPublishingSandbox(ctx, sandboxDir)
	if err != nil {
		return err
	}

	specFile := filepath.Join(rn.sandbox.GetSandboxRoot(), consts.BoxSpecFile)

	err = rn.runDboxedVolumeCleanup(ctx)
	if err != nil {
		return err
	}

	err = rn.writeDboxedAuthFile(ctx)
	if err != nil {
		return err
	}

	if needDestroy {
		slog.InfoContext(ctx, "starting sandbox")
		err = rn.sandbox.Start(ctx)
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, "starting dboxed service inside sandbox")
		err = rn.sandbox.S6SvcUp(ctx, "dboxed")
		if err != nil {
			return err
		}
	}

	lastBoxSpecHash := ""
	for {
		boxFile, err := boxesClient.GetBoxSpecById(ctx, rn.BoxId)
		if err != nil {
			if baseclient.IsNotFound(err) {
				slog.InfoContext(ctx, "box spec was deleted, exiting")
				err = rn.sandbox.Stop(ctx)
				if err != nil {
					return err
				}
				return nil
			}
			slog.ErrorContext(ctx, "error in GetBoxSpecById", slog.Any("error", err))
		} else {
			newHash, err := util.Sha256SumJson(boxFile.Spec)
			if err != nil {
				return err
			}
			if newHash != lastBoxSpecHash {
				slog.InfoContext(ctx, "a new box spec was received")
				b, err := yaml.Marshal(boxFile)
				if err != nil {
					return err
				}
				err = util.AtomicWriteFile(specFile, b, 0600)
				if err != nil {
					return err
				}
				lastBoxSpecHash = newHash
			}
		}

		if !util.SleepWithContext(ctx, time.Second*5) {
			return ctx.Err()
		}
	}
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
	pth := filepath.Join(rn.WorkDir, "boxes", boxName, consts.VethIPStoreFile)
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
	pth := filepath.Join(rn.WorkDir, "boxes", rn.BoxName, consts.VethIPStoreFile)
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
	pth := filepath.Join(rn.WorkDir, "boxes", boxName, consts.BoxSpecUuidFile)
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
	pth := filepath.Join(rn.WorkDir, "boxes", rn.BoxName, consts.BoxSpecUuidFile)
	return util.AtomicWriteFile(pth, []byte(uuid), 0644)
}

func (rn *RunBox) runDboxedVolumeCleanup(ctx context.Context) error {
	// this must run in the host mount namespace, but at the same time we want to run
	// inside the sandbox root, so we must setup a minimal chroot without switching namespaces

	env := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin",
	}

	script := `
set -e
mount -t proc none /proc
mount -t devtmpfs none /dev
mount -t sysfs none /sys
dboxed volume cleanup-loop-devs
`

	cleanupScript := `
umount /proc
umount /dev
umount /sys
`

	defer func() {
		cmd := exec.CommandContext(ctx,
			"chroot",
			rn.sandbox.GetSandboxRoot(),
			"sh", "-c", cleanupScript,
		)

		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}()

	cmd := exec.CommandContext(ctx,
		"chroot",
		rn.sandbox.GetSandboxRoot(),
		"sh", "-c", script,
	)

	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunBox) writeDboxedAuthFile(ctx context.Context) error {
	err := util.AtomicWriteFileYaml(filepath.Join(rn.sandbox.GetSandboxRoot(), consts.BoxClientAuthFile), rn.Client.GetClientAuth(), 0600)
	if err != nil {
		return err
	}
	return nil
}
