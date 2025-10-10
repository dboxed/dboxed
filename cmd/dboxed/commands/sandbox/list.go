//go:build linux

package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
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

	sandboxBaseDir := filepath.Join(g.WorkDir, "sandboxes")

	sandboxInfos, err := sandbox.ListSandboxes(sandboxBaseDir)
	if err != nil {
		return err
	}

	var printList []PrintSandbox
	for _, si := range sandboxInfos {
		s := sandbox.Sandbox{
			Debug:           g.Debug,
			HostWorkDir:     g.WorkDir,
			SandboxName:     si.SandboxName,
			SandboxDir:      filepath.Join(sandboxBaseDir, si.SandboxName),
			VethNetworkCidr: si.VethNetworkCidr,
		}

		runcStatusStr := "unknown"
		runcState, err := s.RuncState(ctx)
		if err == nil && runcState != nil {
			runcStatusStr = runcState.Status
		}

		printList = append(printList, PrintSandbox{
			SandboxName: si.SandboxName,
			Box:         fmt.Sprintf("%s (id=%d)", si.Box.Name, si.Box.ID),
			Workspace:   fmt.Sprintf("%s (id=%d)", si.Box.Name, si.Workspace.ID),
			Status:      runcStatusStr,
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
