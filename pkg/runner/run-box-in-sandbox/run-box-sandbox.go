package run_box_in_sandbox

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/box-spec-runner"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	util2 "github.com/dboxed/dboxed/pkg/util"
)

type RunBoxInSandbox struct {
	Debug bool

	networkConfig *boxspec.NetworkConfig
	dnsProxy      *dns_proxy.DnsProxy

	oldBoxSpecHash string

	dboxedVolumeLog io.WriteCloser
}

func (rn *RunBoxInSandbox) Run(ctx context.Context) error {
	util2.LoadMod(ctx, "dm-mod")
	util2.LoadMod(ctx, "dm-thin-pool")
	util2.LoadMod(ctx, "dm-snapshot")
	util2.LoadMod(ctx, "dm-zero")

	var err error
	rn.networkConfig, err = util.UnmarshalJsonFile[boxspec.NetworkConfig](consts.NetworkConfFile)
	if err != nil {
		return err
	}

	err = rn.startDnsProxy(ctx)
	if err != nil {
		return err
	}

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

	rn.dboxedVolumeLog = logs.BuildRotatingLogger(filepath.Join(consts.LogsDir, "dboxed-volume.log"))
	defer rn.dboxedVolumeLog.Close()

	boxSpecRunner := &box_spec_runner.BoxSpecRunner{
		DboxedVolumeLog: rn.dboxedVolumeLog,
	}
	for {
		err := rn.reconcileBoxSpec(ctx, boxSpecRunner)
		if err != nil {
			slog.ErrorContext(ctx, "error while reconciling box spec", slog.Any("error", err))
		}
		if !util.SleepWithContext(ctx, 2*time.Second) {
			return ctx.Err()
		}
	}
}

func (rn *RunBoxInSandbox) reconcileBoxSpec(ctx context.Context, boxSpecRunner *box_spec_runner.BoxSpecRunner) error {
	boxSpec, err := rn.readBoxSpec()
	if err != nil {
		return err
	}
	hash, err := util.Sha256SumJson(boxSpec)
	if err != nil {
		return err
	}
	if hash == rn.oldBoxSpecHash {
		return nil
	}
	rn.oldBoxSpecHash = hash

	err = boxSpecRunner.Reconcile(ctx, boxSpec)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunBoxInSandbox) readBoxSpec() (*boxspec.BoxSpec, error) {
	b, err := os.ReadFile(consts.BoxSpecFile)
	if err != nil {
		return nil, err
	}

	var boxSpec boxspec.BoxSpec
	err = json.Unmarshal(b, &boxSpec)
	if err != nil {
		return nil, err
	}
	return &boxSpec, nil
}
