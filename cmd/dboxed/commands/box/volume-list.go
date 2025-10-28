package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListVolumesCmd struct {
	Box string `help:"Box ID, UUID, or name" required:""`
}

type PrintVolumeAttachment struct {
	Volume   string `col:"Volume"`
	RootUid  int64  `col:"Root UID"`
	RootGid  int64  `col:"Root GID"`
	RootMode string `col:"Root Mode"`
}

func (cmd *ListVolumesCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	ct := commandutils.NewClientTool(c)

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	attachments, err := c2.ListAttachedVolumes(ctx, b.ID)
	if err != nil {
		return err
	}

	var table []PrintVolumeAttachment
	for _, a := range attachments {
		table = append(table, PrintVolumeAttachment{
			Volume:   ct.Boxes.GetColumn(ctx, a.VolumeID),
			RootUid:  a.RootUid,
			RootGid:  a.RootGid,
			RootMode: a.RootMode,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
