//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type ListCmd struct {
	flags.ListFlags

	All bool `help:"List all sandboxes, not only the ones from the local machine"`
}

type PrintSandbox struct {
	ID          string `col:"ID"`
	Workspace   string `col:"Workspace"`
	Box         string `col:"Box"`
	MachineId   string `col:"Machine Id"`
	Host        string `col:"Host"`
	LocalStatus string `col:"Local Status"`
	ApiStatus   string `col:"API Status"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	sc := clients.SandboxClient{Client: c}
	ct := commandutils.NewClientTool(c)

	apiSandboxes, err := sc.ListSandboxes(ctx)
	if err != nil {
		return err
	}
	apiSandboxesById := map[string]*models.BoxSandbox{}
	for _, sb := range apiSandboxes {
		apiSandboxesById[sb.ID] = &sb
	}

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	localSandboxInfos, err := sandbox.ListSandboxes(sandboxBaseDir)
	if err != nil {
		return err
	}

	machineIdBytes, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return err
	}
	machineId := strings.TrimSpace(string(machineIdBytes))
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	formatMachineId := func(id string) string {
		if id == machineId {
			return fmt.Sprintf("%s (local)", id)
		} else {
			return id
		}
	}

	var table []PrintSandbox
	tableContains := map[string]struct{}{}
	for _, si := range localSandboxInfos {
		s := sandbox.Sandbox{
			Debug:       g.Debug,
			HostWorkDir: g.WorkDir,
			SandboxDir:  filepath.Join(sandboxBaseDir, si.SandboxId),
		}

		statusStr := "unknown"
		cs, err := s.GetSandboxContainerStatus()
		if err == nil {
			statusStr = cs.String()
		}

		te := PrintSandbox{
			ID:          si.SandboxId,
			Workspace:   ct.Workspaces.GetColumn(ctx, si.Box.Workspace, false),
			Box:         si.Box.Name,
			MachineId:   formatMachineId(machineId),
			Host:        hostname,
			LocalStatus: statusStr,
		}

		apiSb, ok := apiSandboxesById[si.SandboxId]
		if !ok {
			te.ApiStatus = "only local"
		} else {
			if apiSb.RunStatus != nil {
				te.ApiStatus = *apiSb.RunStatus
			} else {
				te.ApiStatus = "N/A"
			}
		}

		tableContains[si.SandboxId] = struct{}{}
		table = append(table, te)
	}

	for _, apiSb := range apiSandboxes {
		if _, ok := tableContains[apiSb.ID]; ok {
			continue
		}

		if apiSb.MachineID != machineId && !cmd.All {
			continue
		}

		te := PrintSandbox{
			ID:        apiSb.ID,
			Workspace: ct.Workspaces.GetColumn(ctx, *c.GetWorkspaceId(), false),
			Box:       apiSb.BoxId,
			MachineId: formatMachineId(apiSb.MachineID),
			Host:      apiSb.Hostname,
		}
		if apiSb.MachineID == machineId {
			te.LocalStatus = "deleted"
		} else {
			te.LocalStatus = "N/A"
		}
		if apiSb.RunStatus != nil {
			te.ApiStatus = *apiSb.RunStatus
		} else {
			te.ApiStatus = "N/A"
		}
		table = append(table, te)
	}

	sort.Slice(table, func(i, j int) bool {
		return table[i].ID < table[j].ID
	})

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
