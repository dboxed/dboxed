//go:build linux

package run_machine

import (
	"context"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
)

func (rn *RunMachine) initLogsPublishing(ctx context.Context) error {
	logsDir := filepath.Join(rn.WorkDir, "machine", "logs")
	tta, err := logs.NewTailToApi(ctx, rn.Client, filepath.Join(logsDir, consts.LogsTailDbFilename))
	if err != nil {
		return err
	}

	rn.logsPublisher = &LogsPublisher{
		MachineId: rn.MachineId,
	}

	err = rn.logsPublisher.Start(ctx, tta.MultiTail)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishMachineLogsDir(logsDir)
	if err != nil {
		return err
	}

	return nil
}
