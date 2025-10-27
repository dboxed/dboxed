package volume

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
}

type PrintVolume struct {
	ID         int64  `col:"Id"`
	Name       string `col:"Name"`
	Type       string `col:"Type"`
	Provider   string `col:"Provider"`
	LockTime   string `col:"Lock Time"`
	LockBox    string `col:"Locked by Box"`
	Attachment string `col:"Box attachment"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}
	ct := commandutils.NewClientTool(c)

	volumes, err := c2.ListVolumes(ctx)
	if err != nil {
		return err
	}

	var table []PrintVolume
	for _, v := range volumes {
		p := PrintVolume{
			ID:       v.ID,
			Name:     v.Name,
			Type:     string(v.VolumeProviderType),
			Provider: ct.VolumeProviders.GetColumn(ctx, v.VolumeProviderId),
		}
		if v.LockId != nil && v.LockTime != nil {
			p.LockTime = v.LockTime.String()
		}
		if v.LockBoxId != nil {
			p.LockBox = ct.Boxes.GetColumn(ctx, *v.LockBoxId)
		}
		if v.Attachment != nil {
			p.Attachment = ct.Boxes.GetColumn(ctx, v.Attachment.BoxID)
		}

		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
