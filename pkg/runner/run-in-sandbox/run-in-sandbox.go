package run_in_sandbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/box-spec-runner"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/util"
	util2 "github.com/dboxed/dboxed/pkg/util"
)

type RunInSandbox struct {
	WorkDir string
	Client  *baseclient.Client

	sandboxInfo *sandbox.SandboxInfo

	networkConfig *boxspec.NetworkConfig
	dnsProxy      *dns_proxy.DnsProxy

	logsPublisher logs.LogsPublisher
}

func (rn *RunInSandbox) Run(ctx context.Context) error {
	util2.LoadMod(ctx, "dm-mod")
	util2.LoadMod(ctx, "dm-thin-pool")
	util2.LoadMod(ctx, "dm-snapshot")
	util2.LoadMod(ctx, "dm-zero")

	var err error
	rn.sandboxInfo, err = sandbox.ReadSandboxInfo(consts.DboxedDataDir)
	if err != nil {
		return err
	}

	rn.networkConfig, err = util.UnmarshalYamlFile[boxspec.NetworkConfig](consts.NetworkConfFile)
	if err != nil {
		return err
	}

	err = rn.startDnsProxy(ctx)
	if err != nil {
		return err
	}

	err = rn.initLogsPublishing(ctx)
	if err != nil {
		return err
	}
	defer rn.logsPublisher.Stop(false)

	slog.InfoContext(ctx, "waiting for docker to become available")
	for {
		_, err := dockercli.RunDockerCli(ctx, slog.Default(), true, "", "info")
		if err == nil {
			break
		}
		if !util.SleepWithContext(ctx, 2*time.Second) {
			return ctx.Err()
		}
	}
	slog.InfoContext(ctx, "docker is up and running")

	boxesClient := clients.BoxClient{Client: rn.Client}

	lastBoxSpecHash := ""
	var lastBoxSpec *boxspec.BoxSpec
	for {
		boxSpec, err := boxesClient.GetBoxSpecById(ctx, rn.sandboxInfo.Box.ID)
		if err != nil {
			if baseclient.IsNotFound(err) {
				slog.InfoContext(ctx, "box was deleted, exiting")
				err = rn.shutdown(ctx, lastBoxSpec)
				if err != nil {
					return err
				}
				return nil
			}
			slog.ErrorContext(ctx, "error in GetBoxSpecById", slog.Any("error", err))
		} else {
			newHash, err := util.Sha256SumJson(boxSpec)
			if err != nil {
				return err
			}
			if newHash != lastBoxSpecHash {
				slog.InfoContext(ctx, "a new box spec was received")
				err = rn.reconcileBoxSpec(ctx, boxSpec)
				if err != nil {
					slog.ErrorContext(ctx, "error while reconciling box spec", slog.Any("error", err))
				}
				lastBoxSpecHash = newHash
				lastBoxSpec = boxSpec
			}
		}

		if !util.SleepWithContext(ctx, time.Second*5) {
			return ctx.Err()
		}
	}
}

func (rn *RunInSandbox) reconcileBoxSpec(ctx context.Context, boxSpec *boxspec.BoxSpec) error {
	boxSpecRunner := box_spec_runner.BoxSpecRunner{
		WorkDir: rn.WorkDir,
		BoxSpec: boxSpec,
	}
	err := boxSpecRunner.Reconcile(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunInSandbox) shutdown(ctx context.Context, lastBoxSpec *boxspec.BoxSpec) error {
	if lastBoxSpec != nil {
		boxSpecRunner := box_spec_runner.BoxSpecRunner{
			BoxSpec: lastBoxSpec,
		}

		slog.InfoContext(ctx, "shutting down compose projects")
		err := boxSpecRunner.Down(ctx)
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "shutting down dockerd")
	err := rn.S6SvcDown(ctx, "dockerd")
	if err != nil {
		return err
	}

	// if the box got deleted, we won't be able to upload remaining logs
	rn.logsPublisher.Stop(true)
	// ensure we don't restart the sandbox
	err = util.RunCommand(ctx, "/run/s6/basedir/bin/halt")
	if err != nil {
		return err
	}
	return nil
}
