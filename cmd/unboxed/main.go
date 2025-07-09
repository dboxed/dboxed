package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/koobox/unboxed/cmd/unboxed/commands"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	versionpkg "github.com/koobox/unboxed/pkg/version"
	"log/slog"
	"os"
)

type Cli struct {
	flags.GlobalFlags

	Start   commands.StartCmd   `cmd:"" help:"Download, unpack and start a box"`
	Systemd commands.SystemdCmd `cmd:"" help:"Sub commands to control unboxed systemd integration"`
	Runc    commands.RuncCmd    `cmd:"" help:"Run runc for a box"`

	InitWrapper     commands.InitWrapperCmd     `cmd:"" help:"internal command" hidden:""`
	RunInfraHost    commands.RunInfraHostCmd    `cmd:"" help:"internal command" hidden:""`
	RunInfraSandbox commands.RunInfraSandboxCmd `cmd:"" help:"internal command" hidden:""`
}

func Execute() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	cli := &Cli{}

	ctx := kong.Parse(cli,
		kong.Name("unboxed"),
		kong.Description("A simple container orchestrator."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	err := ctx.Run(&cli.GlobalFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// set via ldflags
var version = ""

func main() {
	// was it set via -ldflags -X
	if //goland:noinspection ALL
	version != "" {
		versionpkg.Version = version
	}

	Execute()
}
