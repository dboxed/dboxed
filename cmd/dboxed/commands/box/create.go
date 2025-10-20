package box

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the box name. Must be unique." required:""`

	Network      *string  `help:"Attach box to specified network (ID or name)."`
	AttachVolume []string `help:"Attach specified volume to new box."`
	ComposeFile  []string `help:"Add specified docker-compose.yml file to new box. Example: --compose-file=name=path/to/docker-compose.yml"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	req := models.CreateBox{
		Name: cmd.Name,
	}

	if cmd.Network != nil {
		n, err := commandutils.GetNetwork(ctx, c, *cmd.Network)
		if err != nil {
			return err
		}
		req.Network = &n.ID
	}

	for _, av := range cmd.AttachVolume {
		v, err := commandutils.GetVolume(ctx, c, av)
		if err != nil {
			return err
		}
		req.VolumeAttachments = append(req.VolumeAttachments, models.AttachVolumeRequest{
			VolumeId: v.ID,
		})
	}
	for _, cp := range cmd.ComposeFile {
		s := strings.SplitN(cp, "=", 2)
		if len(s) < 2 {
			return fmt.Errorf("invalid --compose-project, must be in format '--compose-file name=path/to/docker-compose.yml'")
		}
		p := kong.ExpandPath(s[1])
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		req.ComposeProjects = append(req.ComposeProjects, models.CreateBoxComposeProject{
			Name:           s[0],
			ComposeProject: string(content),
		})
	}

	c2 := &clients.BoxClient{Client: c}

	b, err := c2.CreateBox(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("box created", slog.Any("id", b.ID), slog.Any("name", b.Name))

	return nil
}
