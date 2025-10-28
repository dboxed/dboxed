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

func (rn *RunInSandbox) updateSandboxStatusSimple(ctx context.Context, status string, send bool) {
	rn.updateSandboxStatus(ctx, models.UpdateBoxSandboxStatus2{
		RunStatus: &status,
	}, send)
}

func (rn *RunInSandbox) updateSandboxStatus(ctx context.Context, s models.UpdateBoxSandboxStatus2, send bool) {
	rn.statusMutex.Lock()
	defer rn.statusMutex.Unlock()
	if s.RunStatus != nil {
		rn.sandboxStatus.RunStatus = s.RunStatus
	}
	if s.StartTime != nil {
		rn.sandboxStatus.StartTime = s.StartTime
	}
	if s.StopTime != nil {
		rn.sandboxStatus.StopTime = s.StopTime
	}
	if send {
		rn.sendSandboxStatus(ctx, false)
	}
}

func (rn *RunInSandbox) startUpdateSandboxStatusLoop(ctx context.Context) func() {
	rn.sandboxStatus = models.UpdateBoxSandboxStatus2{
		RunStatus: util.Ptr("starting"),
		StartTime: util.Ptr(time.Now()),
	}

	rn.sendSandboxStatusDockerPs(ctx)
	rn.sendSandboxStatus(ctx, true)

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
				rn.sendSandboxStatusDockerPs(ctx)
				rn.sendSandboxStatus(ctx, true)
			}
		}
	}()
	return func() {
		close(stopCh)
		<-doneCh
	}
}

func (rn *RunInSandbox) sendSandboxStatus(ctx context.Context, lock bool) {
	if lock {
		rn.statusMutex.Lock()
		defer rn.statusMutex.Unlock()
	}
	if rn.sandboxStatusTime.Add(time.Second*30).Before(time.Now()) && util.EqualsViaJson(rn.sandboxStatus, rn.sandboxStatusSent) {
		return
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err := boxesClient.UpdateSandboxStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxSandboxStatus{
		SandboxStatus: &rn.sandboxStatus,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to report run status", "error", err)
	} else {
		rn.sandboxStatusSent = rn.sandboxStatus
	}
}

func (rn *RunInSandbox) sendSandboxStatusDockerPs(ctx context.Context) {
	b, err := rn.runDockerPS(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run docker ps", "error", err)
		return
	}

	rn.statusMutex.Lock()
	if bytes.Equal(b, rn.dockerPSSent) {
		rn.statusMutex.Unlock()
		return
	}
	defer rn.statusMutex.Unlock()

	err = rn.doSendSandboxStatusDockerPs(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "failed to report docker ps result", "error", err)
	} else {
		rn.dockerPSSent = b
	}
	return
}

func (rn *RunInSandbox) doSendSandboxStatusDockerPs(ctx context.Context, b []byte) error {
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
	err = boxesClient.UpdateSandboxStatus(ctx, rn.sandboxInfo.Box.ID, models.UpdateBoxSandboxStatus{
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
