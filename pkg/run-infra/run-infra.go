package run_infra

import (
	"context"
	"encoding/json"
	"github.com/koobox/unboxed/pkg/netbird"
	"github.com/koobox/unboxed/pkg/types"
	"log/slog"
	"os"
)

type RunInfra struct {
	boxSpec types.BoxSpec

	runNetbird *netbird.RunNetbird
}

func (rn *RunInfra) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "running infra")

	boxSpecBytes, err := os.ReadFile("/etc/unboxed/box-spec.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(boxSpecBytes, &rn.boxSpec)
	if err != nil {
		return err
	}

	rn.runNetbird = &netbird.RunNetbird{
		NetbirdManagementUrl: rn.boxSpec.Netbird.ManagementUrl,
		NetbirdSetupKey:      rn.boxSpec.Netbird.SetupKey,
		NetbirdPeerName:      rn.boxSpec.Hostname,
	}
	err = rn.runNetbird.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
