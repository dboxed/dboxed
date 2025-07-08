package main

import (
	"fmt"
	"github.com/koobox/unboxed/cmd/unboxed/run"
	run_infra "github.com/koobox/unboxed/cmd/unboxed/run-infra"
	"github.com/koobox/unboxed/cmd/unboxed/runc"
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
	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(run_infra.RunInfraCmd)
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
