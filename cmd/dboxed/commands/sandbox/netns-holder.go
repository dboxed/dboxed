//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/network"
)

type NetnsHolder struct {
}

func (cmd *NetnsHolder) Run(logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	logFile := filepath.Join(consts.LogsDir, "netns-holder.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)
	defer logHandler.RemoveWriter(logWriter)

	var ul *net.UnixListener
	sigReceived := false
	go func() {
		sig := <-sigs
		sigReceived = true
		slog.Info("received signal", "signal", sig.String())
		_ = ul.Close()
		cancel()
	}()

	hostNsFd, err := network.ReadNetNsFD(ctx, consts.NetNsInitialUnixSocket)
	if err != nil {
		return err
	}
	defer hostNsFd.Close()

	ul, err = network.ListenSCMSocket(consts.NetNsHolderUnixSocket)
	if err != nil {
		return err
	}
	defer ul.Close()

	for {
		slog.Info("waiting for unix socket connection")

		uc, err := ul.AcceptUnix()
		if err != nil {
			if sigReceived {
				return nil
			}
			return err
		}

		slog.Info("sending host netns handle")
		err = network.SendNetNsFD(uc, hostNsFd)
		_ = uc.Close()
		if err != nil {
			slog.Error("SendNetNsFD failed", "error", err)
			if sigReceived {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
