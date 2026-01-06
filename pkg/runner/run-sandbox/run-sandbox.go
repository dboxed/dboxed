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
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/network"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/gofrs/flock"
	"github.com/opencontainers/runc/libcontainer"
	"go4.org/netipx"
)

type RunSandbox struct {
	Debug bool

	Client    *baseclient.Client
	BoxId     string
	SandboxId string

	InfraImage      string
	WorkDir         string
	VethNetworkCidr string

	acquiredVethNetworkCidr string

	sandbox *sandbox.Sandbox
}

func DetermineSandboxId(ctx context.Context, c *baseclient.Client, box *models.Box, workDir string) (string, error) {
	c2 := clients.BoxClient{Client: c}

	machineIdBytes, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	machineId := strings.TrimSpace(string(machineIdBytes))
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	var sandboxId string
	if box.Sandbox != nil {
		sandboxDir := GetSandboxDir(workDir, box.Sandbox.ID)
		si, err := sandbox.ReadSandboxInfo(sandboxDir)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if si == nil {
			if machineId == box.Sandbox.MachineID {
				return "", fmt.Errorf("box has a sandbox (%s) which is not on this host anymore. It looks like it got deleted", box.Sandbox.ID)
			} else {
				return "", fmt.Errorf("box has a sandbox (%s) which is not on this host", box.Sandbox.ID)
			}
		}
		sandboxId = box.Sandbox.ID
	} else {
		slog.InfoContext(ctx, "creating sandbox", "machineId", machineId, "hostname", hostname)
		sb, err := c2.CreateSandbox(ctx, box.ID, models.CreateBoxSandbox{
			MachineId: machineId,
			Hostname:  hostname,
		})
		if err != nil {
			return "", err
		}
		sandboxId = sb.ID
	}

	slog.InfoContext(ctx, "determined sandbox id", "sandboxId", sandboxId)

	return sandboxId, nil
}

func (rn *RunSandbox) GetSandboxDir() string {
	return rn.getSandboxDir2(rn.SandboxId)
}

func (rn *RunSandbox) getSandboxDir2(sandboxId string) string {
	return GetSandboxDir(rn.WorkDir, sandboxId)
}

func GetSandboxDir(workDir string, sandboxId string) string {
	return filepath.Join(workDir, "sandboxes", sandboxId)
}

func (rn *RunSandbox) Run(ctx context.Context) error {
	if rn.Client.GetApiToken() == nil {
		return fmt.Errorf("can only run box with static token")
	}

	sandboxDir := rn.GetSandboxDir()

	err := os.MkdirAll(sandboxDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(sandboxDir, "logs"), 0700)
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

	err = rn.reserveVethCIDR(ctx)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "using veth cidr", slog.Any("cidr", rn.acquiredVethNetworkCidr))

	namesAndIps, err := network.NewNamesAndIPs(rn.SandboxId, rn.acquiredVethNetworkCidr)
	if err != nil {
		return err
	}

	rn.sandbox = &sandbox.Sandbox{
		Debug:                rn.Debug,
		InfraImage:           rn.InfraImage,
		HostWorkDir:          rn.WorkDir,
		SandboxId:            rn.SandboxId,
		SandboxDir:           sandboxDir,
		NetworkNamespaceName: namesAndIps.SandboxNamespaceName,
	}

	needDestroy := false

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

	if needDestroy {
		err = rn.sandbox.StopOrKillSandboxContainer(ctx, time.Second*30, time.Second*10)
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
		err = network.Destroy(ctx, nil, namesAndIps, rn.sandbox.GetSandboxRoot())
		if err != nil {
			return err
		}
	} else {
		err = rn.sandbox.StopRunInSandboxService(ctx, false)
		if err != nil {
			return err
		}
	}

	err = rn.sandbox.CopyBinaries(ctx)
	if err != nil {
		return err
	}

	newSandboxInfo := &sandbox.SandboxInfo{
		SandboxId:               rn.SandboxId,
		Box:                     box,
		Workspace:               workspace,
		GlobalVethNetworkCidr:   rn.VethNetworkCidr,
		AcquiredVethNetworkCidr: rn.acquiredVethNetworkCidr,
	}
	err = sandbox.WriteSandboxInfo(sandboxDir, newSandboxInfo)
	if err != nil {
		return err
	}
	err = sandbox.WriteSandboxInfo(filepath.Join(rn.sandbox.GetSandboxRoot(), consts.DboxedDataDir), newSandboxInfo)
	if err != nil {
		return err
	}

	err = network.SetupSandboxNamespace(ctx, namesAndIps)
	if err != nil {
		return err
	}

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
		err = rn.sandbox.RunDockerCli(ctx, "restart", "dboxed-dns-proxy")
		if err != nil {
			return err
		}
		err = rn.sandbox.RunDockerCli(ctx, "start", "dboxed-run-in-sandbox")
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "sandbox is up and running")

	return nil
}

func (rn *RunSandbox) reserveVethCIDR(ctx context.Context) error {
	slog.InfoContext(ctx, "reserving CIDR for veth pair")

	fl := flock.New(filepath.Join(rn.getSandboxDir2(""), "veth-cidrs.lock"))
	err := fl.Lock()
	if err != nil {
		return err
	}
	defer fl.Unlock()

	p, err := sandbox.ReadVethCidr(rn.GetSandboxDir())
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

	err = sandbox.WriteVethCidr(rn.GetSandboxDir(), &newPrefix)
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
dboxed volume-mount cleanup-loop-devs
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
		filepath.Join(rn.sandbox.GetSandboxRoot(), consts.SandboxClientAuthFile),
		rn.Client.GetClientAuth(true),
		0600,
	)
	if err != nil {
		return err
	}

	hostResolvConf, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(
		filepath.Join(rn.sandbox.GetSandboxRoot(), consts.HostResolvConfFile),
		hostResolvConf,
		0644,
	)
	if err != nil {
		return err
	}
	return nil
}
