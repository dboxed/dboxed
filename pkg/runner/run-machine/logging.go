//go:build linux

package run_machine

import (
	"context"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
)

func (rn *RunMachine) initLogsPublishing(ctx context.Context) error {
	tta, err := logs.NewTailToApi(ctx, rn.Client, filepath.Join(consts.MachineLogsDir, consts.LogsTailDbFilename))
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

	err = rn.logsPublisher.PublishMachineLogsDir(consts.MachineLogsDir)
	if err != nil {
		return err
	}

	return nil
}
