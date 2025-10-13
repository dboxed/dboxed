//go:build linux

package run_sandbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/gofrs/flock"
	"github.com/opencontainers/runc/libcontainer"
	"go4.org/netipx"
	"sigs.k8s.io/yaml"
)

type RunSandbox struct {
	Debug bool

	Client *baseclient.Client
	BoxId  int64

	InfraImage      string
	SandboxName     string
	WorkDir         string
	VethNetworkCidr string

	acquiredVethNetworkCidr string

	logsPublisher logs.LogsPublisher
	sandbox       *sandbox.Sandbox
}

func (rn *RunSandbox) getSandboxDir() string {
	return rn.getSandboxDir2(rn.SandboxName)
}

func (rn *RunSandbox) getSandboxDir2(sandboxName string) string {
	return filepath.Join(rn.WorkDir, "sandboxes", sandboxName)
}

func (rn *RunSandbox) Run(ctx context.Context, logHandler *logs.MultiLogHandler) error {
	if rn.Client.GetApiToken() == nil {
		return fmt.Errorf("can only run box with static token")
	}

	sandboxDir := rn.getSandboxDir()

	err := os.MkdirAll(sandboxDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(sandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}

	err = rn.initFileLogging(ctx, sandboxDir, logHandler)
	if err != nil {
		return err
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	workspacesClient := clients.WorkspacesClient{Client: rn.Client}
	box, err := boxesClient.GetBoxById(ctx, rn.BoxId)
	if err != nil {
		return err
	}
	workspace, err := workspacesClient.GetWorkspaceById(ctx, box.Workspace)
	if err != nil {
		return err
	}

	// we should start publishing logs asap, but the earliest point at the moment is after box spec retrieval
	err = rn.initLogsPublishing(ctx, sandboxDir)
	if err != nil {
		return err
	}
	defer rn.logsPublisher.Stop()

	err = rn.reserveVethCIDR(ctx)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "using veth cidr", slog.Any("cidr", rn.acquiredVethNetworkCidr))

	rn.sandbox = &sandbox.Sandbox{
		Debug:           rn.Debug,
		InfraImage:      rn.InfraImage,
		HostWorkDir:     rn.WorkDir,
		SandboxName:     rn.SandboxName,
		SandboxDir:      sandboxDir,
		VethNetworkCidr: rn.acquiredVethNetworkCidr,
	}

	needDestroy := false

	localSandboxInfo, err := sandbox.ReadSandboxInfo(sandboxDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if localSandboxInfo != nil && localSandboxInfo.Box.Uuid != box.Uuid {
		return fmt.Errorf("sandbox %s already exists and serves a different box (id=%d, uuid=%s)", rn.SandboxName, rn.BoxId, box.Uuid)
	}

	container, err := rn.sandbox.GetSandboxContainer()
	if err != nil {
		if !errors.Is(err, libcontainer.ErrNotExist) {
			return err
		}
	}

	if container == nil {
		needDestroy = true
	} else {
		s, err := container.Status()
		if err != nil {
			return err
		}
		if s != libcontainer.Running {
			slog.InfoContext(ctx, fmt.Sprintf("old sandbox container is in state '%s', re-creating it", s))
			needDestroy = true
		}
	}

	err = rn.sandbox.PrepareNetworkingConfig()
	if err != nil {
		return err
	}

	if needDestroy {
		err = rn.sandbox.StopOrKillSandboxContainer(ctx)
		if err != nil {
			return err
		}

		err = rn.sandbox.Destroy(ctx)
		if err != nil {
			return err
		}
		container = nil

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
		err = rn.sandbox.S6SvcDown(ctx, "run-in-sandbox")
		if err != nil {
			return err
		}
	}

	err = rn.sandbox.CopyBinaries(ctx)
	if err != nil {
		return err
	}

	newSandboxInfo := &sandbox.SandboxInfo{
		SandboxName:     rn.SandboxName,
		Box:             box,
		Workspace:       workspace,
		VethNetworkCidr: rn.VethNetworkCidr,
	}
	err = sandbox.WriteSandboxInfo(sandboxDir, newSandboxInfo)
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

	err = rn.writeDboxedConfFiles(ctx)
	if err != nil {
		return err
	}

	if needDestroy {
		slog.InfoContext(ctx, "starting sandbox")
		err = rn.sandbox.Start(ctx)
		if err != nil {
			return err
		}
		container, err = rn.sandbox.GetSandboxContainer()
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, "starting dboxed service inside sandbox")
		err = rn.sandbox.S6SvcUp(ctx, "run-in-sandbox")
		if err != nil {
			return err
		}
	}

	lastBoxSpecHash := ""
	for {
		s, err := container.Status()
		if err != nil {
			return err
		}
		if s != libcontainer.Running {
			slog.ErrorContext(ctx, "sandbox container is not in running state, exiting", slog.Any("error", err), slog.Any("status", s))
			if err != nil {
				return err
			}
			return fmt.Errorf("sandbox container is not in running state anymore")
		}

		boxFile, err := boxesClient.GetBoxSpecById(ctx, rn.BoxId)
		if err != nil {
			if baseclient.IsNotFound(err) {
				slog.InfoContext(ctx, "box spec was deleted, exiting")
				err = rn.sandbox.StopOrKillSandboxContainer(ctx)
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

func (rn *RunSandbox) reserveVethCIDR(ctx context.Context) error {
	slog.InfoContext(ctx, "reserving CIDR for veth pair")

	fl := flock.New(filepath.Join(rn.WorkDir, "veth-cidrs.lock"))
	err := fl.Lock()
	if err != nil {
		return err
	}
	defer fl.Unlock()

	p, err := sandbox.ReadVethCidr(rn.getSandboxDir())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		rn.acquiredVethNetworkCidr = p.String()
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

	err = sandbox.WriteVethCidr(rn.getSandboxDir(), &newPrefix)
	if err != nil {
		return err
	}

	rn.acquiredVethNetworkCidr = newPrefix.String()

	return nil
}

func (rn *RunSandbox) readReservedIPs() ([]netip.Prefix, error) {
	sandboxesDir := rn.getSandboxDir2("")
	des, err := os.ReadDir(sandboxesDir)
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
		p, err := sandbox.ReadVethCidr(rn.getSandboxDir2(de.Name()))
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

func (rn *RunSandbox) runDboxedVolumeCleanup(ctx context.Context) error {
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

func (rn *RunSandbox) writeDboxedConfFiles(ctx context.Context) error {
	envFile := ""
	envFile += fmt.Sprintf("export DBOXED_SANDBOX=1\n")
	if rn.Debug {
		envFile += fmt.Sprintf("export DBOXED_DEBUG=1\n")
	}

	err := util.AtomicWriteFile(
		filepath.Join(rn.sandbox.GetSandboxRoot(), consts.SandboxEnvironmentFile),
		[]byte(envFile),
		0600,
	)
	if err != nil {
		return err
	}

	err = util.AtomicWriteFileYaml(
		filepath.Join(rn.sandbox.GetSandboxRoot(), consts.BoxClientAuthFile),
		rn.Client.GetClientAuth(true),
		0600,
	)
	if err != nil {
		return err
	}
	return nil
}
