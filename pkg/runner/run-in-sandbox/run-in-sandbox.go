package run_in_sandbox

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/box-spec-runner"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/server/models"
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

	reconcileLogger *slog.Logger
	lastBoxSpecHash string
	lastBoxSpec     *boxspec.BoxSpec

	sandboxStatus     models.UpdateBoxSandboxStatus2
	sandboxStatusSent models.UpdateBoxSandboxStatus2
	sandboxStatusTime time.Time
	dockerPSSent      []byte
	statusMutex       sync.Mutex

	sendStatusStopCh chan struct{}
	sendStatusDoneCh chan struct{}
}

func (rn *RunInSandbox) Run(ctx context.Context) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	// stop the publisher at the very end of Run, so that we try our best to publish all logs, including shutdown logs
	defer rn.logsPublisher.Stop(false)
	defer rn.stopUpdateSandboxStatusLoop()

	shutdown, err := rn.doRun(ctx, sigs)
	if err != nil {
		return err
	}

	if !shutdown {
		if _, err := os.Stat(consts.ShutdownSandboxMarkerFile); err == nil {
			shutdown = true
			slog.InfoContext(ctx, "detected stop marker file, shutting down sandbox")
		}
	}

	if shutdown {
		err = rn.shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *RunInSandbox) doRun(ctx context.Context, sigs chan os.Signal) (bool, error) {
	startTime := time.Now()

	var err error
	rn.sandboxInfo, err = sandbox.ReadSandboxInfo(consts.DboxedDataDir)
	if err != nil {
		return false, err
	}
	rn.networkConfig, err = util.UnmarshalYamlFile[boxspec.NetworkConfig](consts.NetworkConfFile)
	if err != nil {
		return false, err
	}

	// dns proxy must start as early as possible, as otherwise things will fail
	err = rn.startDnsProxy(ctx)
	if err != nil {
		return false, err
	}

	util2.LoadMod(ctx, "dm-mod")
	util2.LoadMod(ctx, "dm-thin-pool")
	util2.LoadMod(ctx, "dm-snapshot")
	util2.LoadMod(ctx, "dm-zero")

	sleepWithSignals := func(d time.Duration) (bool, error) {
		select {
		case <-ctx.Done():
			return true, ctx.Err()
		case <-time.After(2 * time.Second):
			return false, nil
		case s := <-sigs:
			slog.InfoContext(ctx, "received signal, exiting now", slog.Any("signal", s.String()))
			return true, nil
		}
	}

	var reconcileLogWriter io.WriteCloser
	rn.reconcileLogger, reconcileLogWriter = rn.buildReconcileLogger()
	defer reconcileLogWriter.Close()

	rn.updateSandboxStatus(ctx, models.UpdateBoxSandboxStatus2{
		RunStatus: util.Ptr("starting"),
		StartTime: &startTime,
	}, false)

	rn.startUpdateSandboxStatusLoop(ctx)

	err = rn.initLogsPublishing(ctx)
	if err != nil {
		return false, err
	}

	slog.InfoContext(ctx, "waiting for docker to become available")
	for {
		_, err := util.RunCommandStdout(ctx, "docker", "info")
		if err == nil {
			break
		}

		exit, err := sleepWithSignals(2 * time.Second)
		if err != nil {
			return false, err
		}
		if exit {
			return false, nil
		}
	}
	slog.InfoContext(ctx, "docker is up and running")

	rn.updateSandboxStatusSimple(ctx, "reconciling", true)

	for {
		boxesClient := clients.BoxClient{Client: rn.Client}
		boxSpec, err := boxesClient.GetBoxSpecById(ctx, rn.sandboxInfo.Box.ID)
		if err != nil {
			if baseclient.IsNotFound(err) || baseclient.IsUnauthorized(err) {
				slog.InfoContext(ctx, "box was deleted, exiting")
				// if the box got deleted, we won't be able to upload remaining logs, so we cancel immediately to avoid spamming local logs
				rn.logsPublisher.Stop(true)
				return true, nil
			}
			slog.ErrorContext(ctx, "error in GetBoxSpecById", slog.Any("error", err))
		} else {
			if boxSpec.DesiredState != "up" {
				rn.reconcileLogger.InfoContext(ctx, "desired state is not 'up', shutting down", "desiredState", boxSpec.DesiredState)
				return true, nil
			}

			newHash, err := util.Sha256SumJson(boxSpec)
			if err != nil {
				return false, err
			}
			if newHash != rn.lastBoxSpecHash {
				slog.InfoContext(ctx, "a new box spec was received")
				err = rn.reconcileBoxSpec(ctx, boxSpec)
				if err != nil {
					slog.ErrorContext(ctx, "error while reconciling box spec", slog.Any("error", err))
				}
				rn.lastBoxSpecHash = newHash
				rn.lastBoxSpec = boxSpec
			}
		}

		exit, err := sleepWithSignals(5 * time.Second)
		if err != nil {
			return false, err
		}
		if exit {
			return false, nil
		}
	}
}

func (rn *RunInSandbox) reconcileBoxSpec(ctx context.Context, boxSpec *boxspec.BoxSpec) error {
	rn.reconcileLogger.InfoContext(ctx, "starting reconcile of box spec")

	rn.updateSandboxStatusSimple(ctx, "reconciling", true)

	boxSpecRunner := box_spec_runner.BoxSpecRunner{
		WorkDir: rn.WorkDir,
		BoxSpec: boxSpec,
		Log:     rn.reconcileLogger,
	}
	err := boxSpecRunner.Reconcile(ctx)
	if err != nil {
		rn.updateSandboxStatusSimple(ctx, "reconciling failed", true)
		return err
	}

	rn.updateSandboxStatusSimple(ctx, "running", true)
	return nil
}

func (rn *RunInSandbox) shutdown(ctx context.Context) error {
	rn.updateSandboxStatus(ctx, models.UpdateBoxSandboxStatus2{
		RunStatus: util.Ptr("stopping"),
		StopTime:  util.Ptr(time.Now()),
	}, true)
	rn.sendSandboxStatusDockerPs(ctx)

	if rn.lastBoxSpec != nil {
		boxSpecRunner := box_spec_runner.BoxSpecRunner{
			WorkDir: rn.WorkDir,
			BoxSpec: rn.lastBoxSpec,
			Log:     rn.reconcileLogger,
		}

		rn.reconcileLogger.InfoContext(ctx, "shutting down compose projects")
		err := boxSpecRunner.Down(ctx, false, true)
		if err != nil {
			return err
		}
	}

	// final docker ps report
	rn.sendSandboxStatusDockerPs(ctx)

	s6, err := rn.getS6Helper()
	if err != nil {
		return err
	}

	rn.reconcileLogger.InfoContext(ctx, "shutting down dockerd")
	err = s6.S6SvcDown(ctx, "dockerd")
	if err != nil {
		return err
	}
	rn.reconcileLogger.InfoContext(ctx, "dockerd has exited")

	// ensure we don't restart the sandbox
	rn.reconcileLogger.InfoContext(ctx, "running s6 halt")
	err = util.RunCommand(ctx, "/run/s6/basedir/bin/halt")
	if err != nil {
		return err
	}

	rn.updateSandboxStatusSimple(ctx, "stopped", true)

	rn.reconcileLogger.InfoContext(ctx, "shutdown finished")
	return nil
}

func (rn *RunInSandbox) getS6Helper() (*service.S6Helper, error) {
	s6 := &service.S6Helper{}
	return s6, nil
}
