//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	run_in_sandbox_status "github.com/dboxed/dboxed/pkg/runner/run-in-sandbox-status"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sendnshandle"
	"github.com/dboxed/dboxed/pkg/util"
)

type RunInSandboxStatus struct {
}

func (cmd *RunInSandboxStatus) Run() error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	hostNetNsFd, err := sendnshandle.ReadNetNsFD(ctx, consts.NetNsHolderUnixSocket)
	if err != nil {
		return err
	}
	defer hostNetNsFd.Close()

	clientAuth, err := baseclient.ReadClientAuth(util.Ptr(consts.SandboxClientAuthFile))
	if err != nil {
		return err
	}
	// this client will use the host namespace
	client, err := baseclient.New(nil, clientAuth, false, baseclient.WithNetworkNamespace(nil, &hostNetNsFd))
	if err != nil {
		return err
	}

	sandboxInfo, err := sandbox.ReadSandboxInfo(consts.DboxedDataDir)
	if err != nil {
		return err
	}

	lp := run_in_sandbox_status.LogsPublisher{
		Client: client,
		BoxId:  sandboxInfo.Box.ID,
	}

	err = lp.Start(ctx)
	if err != nil {
		return err
	}

	sig := <-sigs
	slog.Info("received signal", "signal", sig.String())

	lp.Stop(util.Ptr(time.Second * 5))

	return nil
}
