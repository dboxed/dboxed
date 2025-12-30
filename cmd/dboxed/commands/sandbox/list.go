//go:build linux

package sandbox

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintSandbox struct {
	ID        string `col:"ID"`
	Workspace string `col:"Workspace"`
	Box       string `col:"Box"`
	Status    string `col:"Status"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	ct := commandutils.NewClientTool(c)

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	sandboxInfos, err := sandbox.ListSandboxes(sandboxBaseDir)
	if err != nil {
		return err
	}

	var table []PrintSandbox
	for _, si := range sandboxInfos {
		s := sandbox.Sandbox{
			Debug:       g.Debug,
			HostWorkDir: g.WorkDir,
			SandboxDir:  filepath.Join(sandboxBaseDir, si.Box.ID),
		}

		statusStr := "unknown"
		cs, err := s.GetSandboxContainerStatus()
		if err == nil {
			statusStr = cs.String()
		}

		table = append(table, PrintSandbox{
			ID:        si.SandboxId,
			Box:       si.Box.Name,
			Workspace: ct.Workspaces.GetColumn(ctx, si.Box.Workspace, false),
			Status:    statusStr,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
