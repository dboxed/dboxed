package run_in_sandbox

import (
	"bytes"
	"compress/gzip"
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunInSandbox) updateBoxRunStatusSimple(ctx context.Context, status string, send bool) {
	rn.updateBoxRunStatus(ctx, models.BoxRunStatusInfo{
		RunStatus: &status,
	}, send)
}

func (rn *RunInSandbox) updateBoxRunStatus(ctx context.Context, s models.BoxRunStatusInfo, send bool) {
	rn.runStatusInfoMutex.Lock()
	defer rn.runStatusInfoMutex.Unlock()
	if s.RunStatus != nil {
		rn.runStatusInfo.RunStatus = s.RunStatus
	}
	if s.StartTime != nil {
		rn.runStatusInfo.StartTime = s.StartTime
	}
	if s.StopTime != nil {
		rn.runStatusInfo.StopTime = s.StopTime
	}
	if send {
		rn.sendBoxRunStatus(ctx, false)
	}
}

func (rn *RunInSandbox) startUpdateBoxRunStatusLoop(ctx context.Context) func() {
	rn.runStatusInfo = models.BoxRunStatusInfo{
		RunStatus: util.Ptr("starting"),
		StartTime: util.Ptr(time.Now()),
	}

	rn.sendBoxRunStatusDockerPs(ctx)
	rn.sendBoxRunStatus(ctx, true)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		for {
			select {
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 10):
				rn.sendBoxRunStatusDockerPs(ctx)
				rn.sendBoxRunStatus(ctx, true)
			}
		}
	}()
	return func() {
		close(stopCh)
		<-doneCh
	}
}

func (rn *RunInSandbox) sendBoxRunStatus(ctx context.Context, lock bool) {
	if lock {
		rn.runStatusInfoMutex.Lock()
		defer rn.runStatusInfoMutex.Unlock()
	}
	if rn.runStatusInfoSentTime.Add(time.Second*30).Before(time.Now()) && util.EqualsViaJson(rn.runStatusInfo, rn.runStatusInfoSent) {
		return
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err := boxesClient.UpdateBoxRunStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxRunStatus{
		RunStatus: &rn.runStatusInfo,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to report run status", "error", err)
	} else {
		rn.runStatusInfoSent = rn.runStatusInfo
	}
}

func (rn *RunInSandbox) sendBoxRunStatusDockerPs(ctx context.Context) {
	b, err := rn.runDockerPS(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run docker ps", "error", err)
		return
	}

	rn.runStatusInfoMutex.Lock()
	if bytes.Equal(b, rn.dockerPSSent) {
		rn.runStatusInfoMutex.Unlock()
		return
	}
	defer rn.runStatusInfoMutex.Unlock()

	err = rn.doSendBoxRunStatusDockerPs(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "failed to report docker ps result", "error", err)
	} else {
		rn.dockerPSSent = b
	}
	return
}

func (rn *RunInSandbox) doSendBoxRunStatusDockerPs(ctx context.Context, b []byte) error {
	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	w := gzip.NewWriter(buf)
	_, err := w.Write(b)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err = boxesClient.UpdateBoxRunStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxRunStatus{
		DockerPs: buf.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunInSandbox) runDockerPS(ctx context.Context) ([]byte, error) {
	c := util.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "-a", "--format", "json"},
	}
	b, err := c.RunStdout(ctx)
	if err != nil {
		return nil, err
	}
	return b, nil
}
