package main

import (
	"fmt"
	pid1_init "github.com/koobox/unboxed/cmd/unboxed/init-wrapper"
	run_infra_host "github.com/koobox/unboxed/cmd/unboxed/run-infra-host"
	run_infra_sandbox "github.com/koobox/unboxed/cmd/unboxed/run-infra-sandbox"
	"github.com/koobox/unboxed/cmd/unboxed/runc"
	"github.com/koobox/unboxed/cmd/unboxed/start"
	"github.com/koobox/unboxed/cmd/unboxed/systemd"
	"github.com/koobox/unboxed/pkg/version"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "unboxed",
	Short: "",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(pid1_init.InitWrapperCmd)
	rootCmd.AddCommand(start.StartCmd)
	rootCmd.AddCommand(run_infra_sandbox.RunInfraSandboxCmd)
	rootCmd.AddCommand(run_infra_host.RunInfraHostCmd)
	rootCmd.AddCommand(systemd.SystemdCmd)
	rootCmd.AddCommand(runc.RuncCmd)
}

func Execute() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var Version = ""

func main() {
	// was it set via -ldflags -X
	if //goland:noinspection ALL
	Version != "" {
		version.Version = Version
	}

	Execute()
}
