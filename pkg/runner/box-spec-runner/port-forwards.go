package box_spec_runner

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/boxspec"
)

func (rn *BoxSpecRunner) reconcilePortForwards(ctx context.Context) error {
	var pfs []boxspec.PortForward
	if rn.BoxSpec.Network != nil {
		pfs = rn.BoxSpec.Network.PortForwards
	}

	slog.InfoContext(ctx, "setting up port forwards", "portForwards", pfs)
	err := rn.PortForwards.SetupPortForwards(ctx, pfs)
	if err != nil {
		return err
	}

	return nil
}
