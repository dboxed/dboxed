package start_box

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/selfupdate"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
)

type StartBox struct {
	BoxUrl          string
	BoxName         string
	WorkDir         string
	VethNetworkCidr *net.IPNet

	boxSpec *types.BoxSpec

	sandbox *sandbox.Sandbox
}

func (rn *StartBox) Start(ctx context.Context) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if os.Getpid() == 1 {
		slog.InfoContext(ctx, "evacuating cgroup2")
		err := cgrouputil.EvacuateCgroup2("init")
		if err != nil {
			return fmt.Errorf("failed to evacuate root cgroup: %w", err)
		}
	}

	var err error
	rn.boxSpec, err = rn.retrieveBoxSpec(ctx)
	if err != nil {
		return err
	}

	err = os.MkdirAll(rn.WorkDir, 0700)
	if err != nil {
		return err
	}

	err = selfupdate.SelfUpdateIfNeeded(ctx, rn.boxSpec.UnboxedBinaryUrl, rn.boxSpec.UnboxedBinaryHash, rn.WorkDir)
	if err != nil {
		return err
	}

	rn.loadModules(ctx)

	rn.sandbox = &sandbox.Sandbox{
		HostWorkDir:     rn.WorkDir,
		SandboxName:     rn.BoxName,
		SandboxDir:      filepath.Join(rn.WorkDir, "boxes", rn.BoxName),
		BoxSpec:         rn.boxSpec,
		VethNetworkCidr: rn.VethNetworkCidr,
	}
	err = rn.sandbox.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
