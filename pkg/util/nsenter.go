package util

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os/exec"
	"runtime"
)

type nsInfo struct {
	flag int
	name string
}

var nss = []nsInfo{
	{unix.CLONE_NEWNS, "mount"},
	{unix.CLONE_NEWUTS, "uts"},
	{unix.CLONE_NEWIPC, "ipc"},
	{unix.CLONE_NEWNET, "net"},
	{unix.CLONE_NEWCGROUP, "cgroup"},
	{unix.CLONE_NEWPID, "pid"},
}

func ExecInLinuxNamespaces(pid int, namespaces int, cmd *exec.Cmd) error {
	if namespaces == 0 {
		for _, ns := range nss {
			namespaces |= ns.flag
		}
	}

	if namespaces&unix.CLONE_NEWNS != 0 {
		p, err := exec.LookPath("nsenter")
		if err != nil {
			return err
		}
		cmd.Path = p

		// we only rely on nsenter for the mount namespace (this can't be easily done in Go)
		var newArgs []string
		newArgs = append(newArgs, p, "-m", "-t", fmt.Sprintf("%d", pid), "--")
		newArgs = append(newArgs, cmd.Args...) // cmd.Path is already part of cmd.Args

		cmd.Args = newArgs
	}

	doSetNS := func(pidFd int) error {
		for _, ns := range nss {
			if namespaces&ns.flag != 0 && ns.flag != unix.CLONE_NEWNS {
				err := unix.Setns(pidFd, ns.flag)
				if err != nil {
					return fmt.Errorf("failed to enter %s namespace: %w", ns.name, err)
				}
			}
		}
		return nil
	}

	errCh := make(chan error)
	go func() {
		runtime.LockOSThread()
		// we do not unlock the OSThread so that it gets killed when finished

		pidFd, err := unix.PidfdOpen(pid, 0)
		if err != nil {
			errCh <- err
			return
		}
		defer unix.Close(pidFd)

		err = doSetNS(pidFd)
		if err != nil {
			errCh <- err
			return
		}
		errCh <- cmd.Run()
	}()

	err := <-errCh
	return err
}
