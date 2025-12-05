package cli

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/box"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/load-balancer"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/login"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/machine"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/network"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/s3-bucket"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/sandbox"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/server"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/token"
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

	Login     login.LoginCmd              `cmd:"" help:"login to the dboxed api"`
	Token     token.TokenCommands         `cmd:"" help:"manage api tokens"`
	Workspace workspace.WorkspaceCommands `cmd:"" help:"manage workspaces"`

	Network      network.NetworkCommands            `cmd:"" help:"manage networks"`
	LoadBalancer load_balancer.LoadBalancerCommands `cmd:"" aliases:"lb" help:"manage load balancers"`

	S3Bucket s3_bucket.S3BucketCommands `cmd:"" name:"s3-bucket" aliases:"s3bucket,s3" help:"manage S3 bucket configurations"`

	VolumeProvider volume_provider.VolumeProviderCommands `cmd:"" help:"manage volume providers"`
	Volume         volume.VolumeCommands                  `cmd:"" help:"manage volumes"`
	VolumeMount    volume_mount.VolumeMountCommands       `cmd:"" help:"manage volume mounts"`

	Box     box.BoxCommands         `cmd:"" help:"manage boxes"`
	Machine machine.MachineCommands `cmd:"" help:"manage machines"`
	Sandbox sandbox.SandboxCommands `cmd:"" help:"manage sandboxes" aliases:"sb"`

	Version VersionCmd `cmd:"" help:"Print version"`

	cliOnlyLinux
}

func Execute() {
	cli := &Cli{}

	ctx := kong.Parse(cli,
		kong.Name("dboxed"),
		kong.Description("Run your cloud workloads on any server you like"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			NoExpandSubcommands: true,
			FlagsLast:           true,
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
		slog.Error("command exited with error", "error", err.Error())
		os.Exit(1)
	}
}
