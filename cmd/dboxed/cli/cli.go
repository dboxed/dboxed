package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/auth"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/box"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/server"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/volume"
	volume_provider "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-provider"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/workspace"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/consts"
)

type Cli struct {
	flags.GlobalFlags

	Server server.ServerCommands `cmd:"" help:"server commands"`

	Auth      auth.AuthCommands           `cmd:"" help:"manage authentication"`
	Workspace workspace.WorkspaceCommands `cmd:"" help:"manage workspaces"`

	VolumeProvider volume_provider.VolumeProviderCommands `cmd:"" help:"manage volume providers"`
	Volume         volume.VolumeCommands                  `cmd:"" help:"manage volumes"`

	Box box.BoxCommands `cmd:"" help:"manage boxes"`

	cliOnlyLinux
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
