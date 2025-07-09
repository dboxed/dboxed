package run_infra_sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"log/slog"
)

type RunInfraSandbox struct {
	conf *types.InfraConfig
}

func (rn *RunInfraSandbox) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "running infra in sandbox namespace")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	err = network.WaitForInterface(ctx, "wt0")
	if err != nil {
		return err
	}

	return nil
}
