package start_box

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/sys/unix"
)

func (rn *StartBox) loadModules(ctx context.Context) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rn.loadModule(ctx, "overlay")
	rn.loadModule(ctx, "rbd")
	rn.loadModule(ctx, "nbd")
}

func (rn *StartBox) loadModule(ctx context.Context, name string) {
	// first try it while assuming that /lib/modules is usable
	// (e.g. because we're bare-metal or inside a container with /lib/modules host mounted)
	err := rn.runModprobeCmd(ctx, name, false)
	if err == nil {
		return
	}
	// now retry by entering PID 1 mount namespace (only works with privileged containers)
	slog.WarnContext(ctx, "modprobe failed, retrying with nsenter", slog.Any("error", err))
	err = rn.runModprobeCmd(ctx, name, true)
	if err == nil {
		return
	}
	slog.WarnContext(ctx, "modprobe with nsenter failed as well", slog.Any("error", err))
}

func (rn *StartBox) runModprobeCmd(ctx context.Context, name string, pid1Namespace bool) error {
	cmd := exec.CommandContext(ctx, "modprobe", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if pid1Namespace {
		return util.ExecInLinuxNamespaces(1, unix.CLONE_NEWNS, cmd)
	} else {
		return cmd.Run()
	}
}
