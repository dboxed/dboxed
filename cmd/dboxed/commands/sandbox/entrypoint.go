//go:build linux

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/sendnshandle"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"golang.org/x/sys/unix"
)

const debugReaping = true

type EntrypointCmd struct {
	Child bool

	childProcess *os.Process
}

func (cmd *EntrypointCmd) Run(logHandler *logs.MultiLogHandler) error {
	if !cmd.Child {
		return cmd.runReaperPid1()
	}

	ctx := context.Background()

	logFile := filepath.Join(consts.LogsDir, "sandbox-entrypoint.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)

	exitCh := make(chan struct{})
	var doneWg sync.WaitGroup

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)

		slog.Info("closing exitCh")
		close(exitCh)
		slog.Info("waiting for goroutines to finish")
		doneWg.Wait()
		slog.Info("goroutines exited")
	}()

	err := cmd.startNetnsHolder(ctx, exitCh, &doneWg)
	if err != nil {
		return err
	}

	err = cmd.startDockerd(ctx, exitCh, &doneWg)
	if err != nil {
		return err
	}

	err = cmd.loadBusyboxImage(ctx)
	if err != nil {
		return err
	}

	err = cmd.startDockerContainers(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "running")
loop:
	for {
		select {
		case sig := <-sigs:
			slog.InfoContext(ctx, "received signal", "signal", sig.String())
			break loop
		case <-time.After(200 * time.Millisecond):
			if _, err := os.Stat("/exit-marker"); err == nil {
				slog.InfoContext(ctx, "exit-marker detected")
				break loop
			}
		}
	}

	err = cmd.stopDockerContainers(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *EntrypointCmd) runReaperPid1() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGCHLD, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	c := exec.Command(os.Args[0], "sandbox", "entrypoint", "--child")
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin

	err := c.Start()
	if err != nil {
		return err
	}
	cmd.childProcess = c.Process

	if debugReaping {
		fmt.Printf("reaper: forked child pid = %d\n", c.Process.Pid)
	}

	for {
		sig := <-sigs

		if debugReaping {
			fmt.Printf("reaper: received signal %s\n", sig)
		}

		switch sig {
		case syscall.SIGCHLD:
			childExited, err := cmd.reapAll()
			if err != nil {
				return err
			}
			if childExited {
				return nil
			}
		default:
			if debugReaping {
				fmt.Printf("reaper: sending %s to pid %d\n", sig, c.Process.Pid)
			}
			_ = c.Process.Signal(sig)
		}
	}
}

func (cmd *EntrypointCmd) reapAll() (bool, error) {
	childReaped := false
	for {
		pid, wstatus, err := cmd.reapOne()
		if err != nil {
			return false, err
		}
		if pid == -1 {
			return childReaped, nil
		}

		if pid == cmd.childProcess.Pid {
			childReaped = true
		}

		if debugReaping {
			status := -1
			if wstatus != nil {
				status = wstatus.ExitStatus()
			}
			fmt.Printf("reaper: cleanup pid=%d, status=%d\n", pid, status)
		}
	}
}

func (cmd *EntrypointCmd) reapOne() (int, *syscall.WaitStatus, error) {
	var wstatus syscall.WaitStatus

	for {
		pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
		if err == nil {
			return pid, &wstatus, nil
		}
		if errors.Is(err, syscall.ECHILD) {
			return -1, nil, nil
		}
		if !errors.Is(err, syscall.EINTR) {
			return -1, nil, err
		}
	}
}

func (cmd *EntrypointCmd) startNetnsHolder(ctx context.Context, exitCh chan struct{}, wg *sync.WaitGroup) error {
	var ul *net.UnixListener

	hostNsFd, err := sendnshandle.ReadNetNsFD(ctx, consts.NetNsInitialUnixSocket)
	if err != nil {
		return err
	}

	ul, err = sendnshandle.ListenSCMSocket(consts.NetNsHolderUnixSocket)
	if err != nil {
		hostNsFd.Close()
		return err
	}

	var isDone atomic.Bool
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-exitCh
		isDone.Store(true)
		_ = ul.Close()
	}()

	wg.Add(1)
	go func() {
		defer hostNsFd.Close()
		defer wg.Done()
		for !isDone.Load() {
			slog.Info("waiting for unix socket connection")

			uc, err := ul.AcceptUnix()
			if err != nil {
				if !isDone.Load() {
					slog.Error("ul.AcceptUnix returned error", "error", err)
				}
				return
			}

			slog.Info("sending host netns handle")
			err = sendnshandle.SendNetNsFD(uc, hostNsFd)
			_ = uc.Close()
			if err != nil {
				slog.Error("SendNetNsFD failed", "error", err)
				if isDone.Load() {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return nil
}

func (cmd *EntrypointCmd) startDockerd(ctx context.Context, exitCh chan struct{}, wg *sync.WaitGroup) error {
	rlog := logs.BuildRotatingLogger(filepath.Join(consts.LogsDir, "dockerd.log"))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			exit, err := cmd.runDockerd(ctx, rlog, exitCh)
			if err != nil {
				slog.Error("runDockerd returned error", "error", err, "exit", exit)
			} else {
				slog.Info("runDockerd exited", "exit", exit)
			}
			if exit {
				return
			}
		}
	}()

	slog.InfoContext(ctx, "waiting for dockerd to become ready")
	for {
		c := exec.CommandContext(ctx, "docker", "ps")
		err := c.Run()
		if err != nil {
			slog.ErrorContext(ctx, "docker ps returned error", "error", err)
		} else {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (cmd *EntrypointCmd) runDockerd(ctx context.Context, rlog io.Writer, exitCh chan struct{}) (bool, error) {
	args := []string{
		"--host=unix:///var/run/docker.sock",
		"--log-format=json",
	}

	slog.InfoContext(ctx, "starting dockerd", "args", args)

	c := exec.CommandContext(ctx, "dockerd", args...)
	c.Stdout = rlog
	c.Stderr = rlog
	err := c.Start()
	if err != nil {
		return false, err
	}

	errCh := make(chan error)
	go func() {
		err := c.Wait()
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return false, err
	case <-exitCh:
		_ = c.Process.Signal(unix.SIGTERM)
		return true, <-errCh
	}
}

func (cmd *EntrypointCmd) loadBusyboxImage(ctx context.Context) error {
	slog.InfoContext(ctx, "loading busybox image")

	f, err := os.Open("/busybox-image.tar")
	if err != nil {
		return err
	}
	defer f.Close()

	c := command_helper.CommandHelper{
		Command: "docker",
		Args:    []string{"image", "load"},
		Stdin:   f,
		Logger:  slog.Default(),
		LogCmd:  true,
	}
	err = c.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *EntrypointCmd) startDockerContainers(ctx context.Context) error {
	slog.InfoContext(ctx, "starting dboxed containers")

	err := cmd.startDockerContainer(ctx,
		"dboxed-dns-proxy",
		nil,
		"dboxed-busybox",
		[]string{"chroot", "/hostfs", "dboxed", "sandbox", "dns-proxy"},
	)
	if err != nil {
		return err
	}

	err = cmd.startDockerContainer(ctx,
		"dboxed-run-in-sandbox-status",
		nil,
		"dboxed-busybox",
		[]string{"chroot", "/hostfs", "dboxed", "sandbox", "run-in-sandbox-status"},
	)
	if err != nil {
		return err
	}

	err = cmd.startDockerContainer(ctx,
		"dboxed-run-in-sandbox",
		nil,
		"dboxed-busybox",
		[]string{"chroot", "/hostfs", "dboxed", "sandbox", "run-in-sandbox"},
	)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *EntrypointCmd) stopDockerContainers(ctx context.Context) error {
	slog.Info("stopping containers")
	err := cmd.stopDockerContainer(ctx, "dboxed-run-in-sandbox")
	if err != nil {
		return err
	}
	err = cmd.stopDockerContainer(ctx, "dboxed-run-in-sandbox-status")
	if err != nil {
		return err
	}
	return nil
}

func (cmd *EntrypointCmd) startDockerContainer(ctx context.Context, name string, dockerArgs []string, image string, containerArgs []string) error {
	var args []string
	args = append(args, "run")
	args = append(args, "-d", "-v/:/hostfs", "--privileged", "--init", "--restart=on-failure", "--net=host", "--pid=host")
	args = append(args, "--name", name)
	args = append(args, dockerArgs...)
	args = append(args, image)
	args = append(args, containerArgs...)

	c := command_helper.CommandHelper{
		Command: "docker",
		Args:    args,
		Logger:  slog.Default(),
		LogCmd:  true,
	}
	err := c.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *EntrypointCmd) stopDockerContainer(ctx context.Context, name string) error {
	if !cmd.isDockerContainerRunning(ctx, name) {
		return nil
	}

	var args []string
	args = append(args, "stop", name)

	c := command_helper.CommandHelper{
		Command: "docker",
		Args:    args,
		Logger:  slog.Default(),
		LogCmd:  true,
	}
	err := c.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *EntrypointCmd) runDockerPs(ctx context.Context) ([]dockercli.DockerPS, error) {
	c := command_helper.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "--format=json"},
		Logger:  slog.Default(),
		LogCmd:  true,
	}
	var psList []dockercli.DockerPS
	err := c.RunStdoutJsonLines(ctx, &psList)
	if err != nil {
		return nil, err
	}
	return psList, nil
}

func (cmd *EntrypointCmd) isDockerContainerRunning(ctx context.Context, name string) bool {
	l, err := cmd.runDockerPs(ctx)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(l, func(ps dockercli.DockerPS) bool {
		return ps.Names == name && ps.State == "running"
	})
}
