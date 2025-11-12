//go:build linux

package sandbox

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/network"
)

type NetnsHolder struct {
}

func (cmd *NetnsHolder) Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	slog.Info("receiving host netns handle")
	hostNsFd, err := network.ReadFD(consts.NetNsInitialUnixSocket)
	if err != nil {
		return err
	}

	ul, err := network.ListenSCMSocket(consts.NetNsHolderUnixSocket)
	if err != nil {
		return err
	}
	defer ul.Close()

	closed := false
	go func() {
		sig := <-sigs
		slog.Info("received signal", "signal", sig.String())
		closed = true
		_ = ul.Close()
	}()

	for {
		slog.Info("waiting for unix socket connection")

		uc, err := ul.AcceptUnix()
		if err != nil {
			if closed {
				return nil
			}
			return err
		}

		slog.Info("sending host netns handle")
		err = network.SendFD(uc, hostNsFd)
		_ = uc.Close()
		if err != nil {
			slog.Error("SendFD failed", "error", err)
			if closed {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
