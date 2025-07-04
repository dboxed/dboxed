package main

import (
	"context"
	run_infra "github.com/koobox/unboxed/pkg/run-infra"
	"github.com/spf13/cobra"
	"time"
)

var runInfraCmd = &cobra.Command{
	Use:    "run-infra",
	Short:  "Run infra",
	Long:   ``,
	RunE:   doRunInfra,
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(runInfraCmd)
}

func doRunInfra(cmd *cobra.Command, args []string) error {
	runInfra := run_infra.RunInfra{}

	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := runInfra.Start(ctx)
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
