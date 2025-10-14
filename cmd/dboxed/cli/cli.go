package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/auth"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/box"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/sandbox"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/server"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/volume"
	volume_mount "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-mount"
	volume_provider "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-provider"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/workspace"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
)

type Cli struct {
	flags.GlobalFlags

	Server server.ServerCommands `cmd:"" help:"server commands"`

	Auth      auth.AuthCommands           `cmd:"" help:"manage authentication"`
	Workspace workspace.WorkspaceCommands `cmd:"" help:"manage workspaces"`

	VolumeProvider volume_provider.VolumeProviderCommands `cmd:"" help:"manage volume providers"`
	Volume         volume.VolumeCommands                  `cmd:"" help:"manage volumes"`
	VolumeMount    volume_mount.VolumeMountCommands       `cmd:"" help:"manage volume mounts"`

	Box     box.BoxCommands         `cmd:"" help:"manage boxes"`
	Sandbox sandbox.SandboxCommands `cmd:"" help:"manage sandboxes" aliases:"sb"`

	Version VersionCmd `cmd:"" help:"Print version"`

	cliOnlyLinux
}

func Execute() {
	cli := &Cli{}

	ctx := kong.Parse(cli,
		kong.Name("dboxed"),
		kong.Description("A simple container orchestrator."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.DefaultEnvars("DBOXED"),
		kong.Vars{
			"default_infra_image": consts.GetDefaultInfraImage(),
		})

	logLevel := slog.LevelInfo
	if cli.GlobalFlags.Debug {
		logLevel = slog.LevelDebug
	}

	handler := logs.NewMultiLogHandler(logLevel)
	handler.AddWriter(os.Stderr)
	slog.SetDefault(slog.New(handler))

	err := ctx.Run(&cli.GlobalFlags, handler)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
