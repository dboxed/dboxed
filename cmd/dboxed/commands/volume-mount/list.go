//go:build linux

package volume_mount

import (
	"context"
	"encoding/json"
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
	MountName string `json:"mountName"`
	Volume    string `json:"volume"`
	Workspace string `json:"workspace"`
	Box       string `json:"box"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	ct := commandutils.ClientTool{
		Client: c,
	}

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumes, err := volume_serve.ListVolumeState(baseDir)
	if err != nil {
		return err
	}

	var printList []PrintVolumeMount
	for _, v := range volumes {
		p := PrintVolumeMount{
			MountName: v.MountName,
			Volume:    fmt.Sprintf("%s (id=%d)", v.Volume.Name, v.Volume.ID),
			Workspace: ct.GetWorkspaceColumn(ctx, v.Volume.Workspace),
		}
		if v.BoxId != nil {
			p.Box = ct.GetBoxColumn(ctx, *v.BoxId)
		}

		printList = append(printList, p)
	}

	for _, p := range printList {
		j, err := json.Marshal(p)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", string(j))
	}

	return nil
}
