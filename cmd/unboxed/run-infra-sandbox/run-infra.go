package run_infra_sandbox

import (
	"context"
	run_infra "github.com/koobox/unboxed/pkg/run-infra-sandbox"
	"github.com/spf13/cobra"
	"time"
)

var RunInfraSandboxCmd = &cobra.Command{
	Use:    "run-infra-sandbox",
	Short:  "Run infra inside sandbox",
	Long:   ``,
	RunE:   doRunInfraSandbox,
	Hidden: true,
}

func doRunInfraSandbox(cmd *cobra.Command, args []string) error {
	runInfraSandbox := run_infra.RunInfraSandbox{}

	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := runInfraSandbox.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
