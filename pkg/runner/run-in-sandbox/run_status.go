package run_in_sandbox

import (
	"bytes"
	"compress/gzip"
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunInSandbox) updateBoxRunStatusSimple(ctx context.Context, status string) {
	rn.updateBoxRunStatus(ctx, models.BoxRunStatusInfo{
		RunStatus: &status,
	})
}

func (rn *RunInSandbox) updateBoxRunStatus(ctx context.Context, s models.BoxRunStatusInfo) {
	boxesClient := clients.BoxClient{Client: rn.Client}
	err := boxesClient.UpdateBoxRunStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxRunStatus{
		RunStatus: &s,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to report run status", "error", err)
	}
}

func (rn *RunInSandbox) updateBoxRunStatusDockerPs(ctx context.Context) {
	b, err := rn.runDockerPSCompress(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run docker ps", "error", err)
		return
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err = boxesClient.UpdateBoxRunStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxRunStatus{
		DockerPs: b,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to report docker ps result", "error", err)
	}
}

func (rn *RunInSandbox) runDockerPSCompress(ctx context.Context) ([]byte, error) {
	c := util.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "-a", "--format", "json"},
	}
	b, err := c.RunStdout(ctx)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	w := gzip.NewWriter(buf)
	_, err = w.Write(b)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
