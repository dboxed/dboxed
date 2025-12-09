//go:build linux

package volume_mount

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintVolumeMount struct {
	ID        string `col:"ID"`
	Volume    string `col:"Volume"`
	Workspace string `col:"Workspace"`
	Box       string `col:"Box"`
	MountId   string `col:"Mount ID"`
	MountTime string `col:"Mount Time"`
	Restored  string `col:"Restored"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	ct := commandutils.NewClientTool(c)

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumes, err := volume_serve.ListVolumeState(baseDir)
	if err != nil {
		return err
	}

	var table []PrintVolumeMount
	for _, v := range volumes {
		p := PrintVolumeMount{
			ID:        v.Volume.ID,
			Volume:    v.Volume.Name,
			Workspace: ct.Workspaces.GetColumn(ctx, v.Volume.Workspace, false),
		}
		if v.Volume != nil {
			if v.Volume.MountId != nil {
				p.MountId = *v.Volume.MountId
				if v.Volume.MountStatus != nil {
					if v.Volume.MountStatus.BoxId != nil {
						p.Box = ct.Boxes.GetColumn(ctx, *v.Volume.MountStatus.BoxId, false)
					}
					p.MountTime = v.Volume.MountStatus.MountTime.String()
				}
			}
		}
		if v.RestoreSnapshot != nil {
			p.Restored = fmt.Sprintf("from snapshot %s", *v.RestoreSnapshot)
		}

		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
