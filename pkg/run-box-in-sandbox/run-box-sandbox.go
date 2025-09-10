package run_box_in_sandbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/dboxed/dboxed-common/util"
	box_spec_runner "github.com/dboxed/dboxed/pkg/box-spec-runner"
	"github.com/dboxed/dboxed/pkg/dockercli"
	"github.com/dboxed/dboxed/pkg/types"
	util2 "github.com/dboxed/dboxed/pkg/util"
)

type RunBoxInSandbox struct {
	Debug bool

	oldBoxSpecHash string
}

func (rn *RunBoxInSandbox) Run(ctx context.Context) error {
	util2.LoadMod(ctx, "dm-mod")
	util2.LoadMod(ctx, "dm-thin-pool")
	util2.LoadMod(ctx, "dm-snapshot")
	util2.LoadMod(ctx, "dm-zero")

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

	for {
		err := rn.reconcileBoxSpec(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error while reconciling box spec", slog.Any("error", err))
		}
		if !util.SleepWithContext(ctx, 2*time.Second) {
			return ctx.Err()
		}
	}
}

func (rn *RunBoxInSandbox) reconcileBoxSpec(ctx context.Context) error {
	boxSpec, hash, err := rn.readBoxSpec()
	if err != nil {
		return err
	}
	if hash == rn.oldBoxSpecHash {
		return nil
	}
	rn.oldBoxSpecHash = hash
	boxSpecRunner := &box_spec_runner.BoxSpecRunner{
		BoxSpec: boxSpec,
	}
	err = boxSpecRunner.Reconcile(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunBoxInSandbox) readBoxSpec() (*types.BoxSpec, string, error) {
	b, err := os.ReadFile(types.BoxSpecFile)
	if err != nil {
		return nil, "", err
	}

	var boxSpec types.BoxSpec
	err = json.Unmarshal(b, &boxSpec)
	if err != nil {
		return nil, "", err
	}
	return &boxSpec, util.Sha256Sum(b), nil
}
