package volume

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dustin/go-humanize"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintVolume struct {
	ID                 string `col:"ID" id:"true"`
	Name               string `col:"Name"`
	Type               string `col:"Type"`
	Provider           string `col:"Provider"`
	MountTime          string `col:"Mount Time"`
	MountBox           string `col:"Mounted by Box"`
	Attachment         string `col:"Box attachment"`
	LatestSnapshotId   string `col:"Snapshot ID"`
	LatestSnapshotTime string `col:"Snapshot Time"`
	LatestSnapshotSize string `col:"Snapshot Size"`
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
			Provider: ct.VolumeProviders.GetColumn(ctx, v.VolumeProviderId, false),
		}
		if v.MountId != nil && v.MountStatus != nil {
			p.MountTime = v.MountStatus.MountTime.String()
			if v.MountStatus.BoxId != nil {
				p.MountBox = ct.Boxes.GetColumn(ctx, *v.MountStatus.BoxId, false)
			}
		}
		if v.Attachment != nil {
			p.Attachment = ct.Boxes.GetColumn(ctx, v.Attachment.BoxID, false)
		}
		if v.LatestSnapshotId != nil {
			snapshot, err := c2.GetVolumeSnapshotById(ctx, v.ID, *v.LatestSnapshotId)
			if err == nil && snapshot != nil {
				if snapshot.Rustic != nil {
					p.LatestSnapshotId = snapshot.ID
					p.LatestSnapshotTime = snapshot.Rustic.SnapshotTime.String()
					p.LatestSnapshotSize = humanize.Bytes(uint64(snapshot.Rustic.TotalBytesProcessed))
				}
			}
		}

		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
