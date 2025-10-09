//go:build linux

package box

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/run-box"
	"github.com/dboxed/dboxed/pkg/util"
)

type RunCmd struct {
	Box       string  `help:"Specify box name or id" required:"" arg:""`
	LocalName *string `help:"Override local box name. Defaults to the box <name>-<uuid>"`

	InfraImage  string `help:"Specify the infra/sandbox image to use" default:"${default_infra_image}"`
	VethCidrArg string `help:"CIDR to use for veth pairs. dboxed will dynamically allocate 2 IPs from this CIDR per box" default:"1.2.3.0/24"`

	WaitBeforeExit *time.Duration `help:"Wait before finally exiting. This gives the process time to print stdout/stderr messages that might be lost. Especially useful in combination with --debug"`
}

func (cmd *RunCmd) Run(g *flags.GlobalFlags, logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	defer func() {
		if cmd.WaitBeforeExit != nil {
			slog.Info("sleeping before exit")
			time.Sleep(*cmd.WaitBeforeExit)
		}
	}()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	box, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	localName := fmt.Sprintf("%s-%s", box.Name, box.Uuid)
	if cmd.LocalName != nil {
		localName = *cmd.LocalName
	}
	err = util.CheckName(localName)
	if err != nil {
		return err
	}

	runBox := run_box.RunBox{
		Debug:           g.Debug,
		Client:          c,
		BoxId:           box.ID,
		InfraImage:      cmd.InfraImage,
		BoxName:         localName,
		WorkDir:         g.WorkDir,
		VethNetworkCidr: cmd.VethCidrArg,
	}

	err = runBox.Run(ctx, logHandler)
	if err != nil {
		return err
	}

	return nil
}
