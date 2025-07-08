package run_infra_host

import (
	"context"
	run_infra "github.com/koobox/unboxed/pkg/run-infra-host"
	"github.com/spf13/cobra"
	"log/slog"
	"time"
)

var RunInfraHostCmd = &cobra.Command{
	Use:    "run-infra-host",
	Short:  "Run infra on host",
	Long:   ``,
	RunE:   doRunInfraHost,
	Hidden: true,
}

func doRunInfraHost(cmd *cobra.Command, args []string) error {
	runInfraHost := run_infra.RunInfraHost{}

	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := runInfraHost.Start(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error in runInfraHost.Start", slog.Any("error", err))
		//return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
