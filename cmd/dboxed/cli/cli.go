package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dboxed/dboxed/cmd/dboxed/commands"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/consts"
)

type Cli struct {
	flags.GlobalFlags

	Run     commands.RunCmd     `cmd:"" help:"Download, unpack and run a box"`
	Systemd commands.SystemdCmd `cmd:"" help:"Sub commands to control dboxed systemd integration"`
	Runc    commands.RuncCmd    `cmd:"" help:"Run runc for a box"`

	RunBoxInSandbox commands.RunBoxInSandbox `cmd:"" hidden:""`
}

func Execute() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	cli := &Cli{}

	ctx := kong.Parse(cli,
		kong.Name("dboxed"),
		kong.Description("A simple container orchestrator."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"default_infra_image": consts.GetDefaultInfraImage(),
		})

	err := ctx.Run(&cli.GlobalFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
