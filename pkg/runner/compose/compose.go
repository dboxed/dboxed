package compose

import (
	"context"

	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/util"
)

func ListRunningComposeProjects(ctx context.Context) ([]dockercli.DockerComposeListEntry, error) {
	cmd := util.CommandHelper{
		Command: "docker",
		Args:    []string{"compose", "ls", "-a", "--format", "json"},
	}
	var l []dockercli.DockerComposeListEntry
	err := cmd.RunStdoutJson(ctx, &l)
	if err != nil {
		return nil, err
	}
	return l, nil
}
