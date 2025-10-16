//go:build linux

package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type ListCmd struct {
}

type PrintSandbox struct {
	SandboxName string `json:"sandboxName"`
	Workspace   string `json:"workspace"`
	Box         string `json:"box"`
	Status      string `json:"status"`
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

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	sandboxInfos, err := sandbox.ListSandboxes(sandboxBaseDir)
	if err != nil {
		return err
	}

	var printList []PrintSandbox
	for _, si := range sandboxInfos {
		s := sandbox.Sandbox{
			Debug:           g.Debug,
			HostWorkDir:     g.WorkDir,
			SandboxDir:      filepath.Join(sandboxBaseDir, si.Box.Uuid),
			VethNetworkCidr: si.VethNetworkCidr,
		}

		statusStr := "unknown"
		cs, err := s.GetSandboxContainerStatus()
		if err == nil {
			statusStr = cs.String()
		}

		printList = append(printList, PrintSandbox{
			SandboxName: si.SandboxName,
			Box:         fmt.Sprintf("%s (id=%d)", si.Box.Name, si.Box.ID),
			Workspace:   ct.GetWorkspaceColumn(ctx, si.Box.Workspace),
			Status:      statusStr,
		})
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
