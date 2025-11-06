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
}

type PrintVolumeMount struct {
	MountName string `col:"Mount Name"`
	Volume    string `col:"Volume"`
	Workspace string `col:"Workspace"`
	Box       string `col:"Box"`
	LockId    string `col:"Lock ID"`
	LockTime  string `col:"Lock Time"`
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
			MountName: v.MountName,
			Volume:    fmt.Sprintf("%s (id=%s)", v.Volume.Name, v.Volume.ID),
			Workspace: ct.Workspaces.GetColumn(ctx, v.Volume.Workspace),
		}
		if v.Volume != nil {
			if v.Volume.LockBoxId != nil {
				p.Box = ct.Boxes.GetColumn(ctx, *v.Volume.LockBoxId)
			}
			if v.Volume.LockId != nil {
				p.LockId = *v.Volume.LockId
			}
			if v.Volume.LockTime != nil {
				p.LockTime = v.Volume.LockTime.String()
			}
		}
		if v.RestoreSnapshot != nil {
			p.Restored = fmt.Sprintf("from snapshot %s", *v.RestoreSnapshot)
		}

		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
