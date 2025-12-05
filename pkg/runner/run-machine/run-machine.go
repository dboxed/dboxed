//go:build linux

package run_machine

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type RunMachine struct {
	Debug      bool
	WorkDir    string
	InfraImage string
	VethCidr   string

	Client    *baseclient.Client
	MachineId string

	logsPublisher *LogsPublisher

	machineStatus     models.UpdateMachineRunStatus
	machineStatusSent models.UpdateMachineRunStatus
	machineStatusTime time.Time
	statusMutex       sync.Mutex

	sendStatusStopCh chan struct{}
	sendStatusDone   sync.WaitGroup
}

func (rn *RunMachine) Run(ctx context.Context, logHandler *logs.MultiLogHandler) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	err := os.MkdirAll(consts.MachineLogsDir, 0755)
	if err != nil {
		return err
	}

	logFile := filepath.Join(consts.MachineLogsDir, "machine-run.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)
	defer logHandler.RemoveWriter(logWriter)

	// stop the publisher at the very end of Run, so that we try our best to publish all logs, including shutdown logs
	defer rn.logsPublisher.Stop(util.Ptr(time.Second * 5))
	defer rn.stopUpdateMachineStatusLoop()

	shutdown, err := rn.doRun(ctx, sigs)
	if err != nil {
		return err
	}

	if shutdown {
		err = rn.shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *RunMachine) doRun(ctx context.Context, sigs chan os.Signal) (bool, error) {
	startTime := time.Now()

	err := rn.initLogsPublishing(ctx)
	if err != nil {
		return false, err
	}

	sleepWithSignals := func(d time.Duration) (bool, error) {
		select {
		case <-ctx.Done():
			return true, ctx.Err()
		case <-time.After(2 * time.Second):
			return false, nil
		case s := <-sigs:
			slog.InfoContext(ctx, "received signal, exiting now", slog.Any("signal", s.String()))
			return true, nil
		}
	}

	rn.updateMachineStatus(ctx, models.UpdateMachineRunStatus{
		RunStatus: util.Ptr("starting"),
		StartTime: &startTime,
	}, true)
	rn.startUpdateMachineStatusLoop(ctx)

	mc := clients.MachineClient{Client: rn.Client}
	firstLoop := true
	for {
		if !firstLoop {
			exit, err := sleepWithSignals(5 * time.Second)
			if err != nil {
				return false, err
			}
			if exit {
				return false, nil
			}
		}
		firstLoop = false

		machine, err := mc.GetMachineById(ctx, rn.MachineId)
		if err != nil {
			if baseclient.IsNotFound(err) || baseclient.IsUnauthorized(err) {
				slog.InfoContext(ctx, "machine was deleted, exiting")
				// if the box got deleted, we won't be able to upload remaining logs, so we cancel immediately to avoid spamming local logs
				rn.logsPublisher.Stop(util.Ptr(time.Second * 1))
				return true, nil
			}
			slog.ErrorContext(ctx, "error in GetMachineById", slog.Any("error", err))
			continue
		}
		boxes, err := mc.ListBoxes(ctx, machine.ID)
		if err != nil {
			slog.ErrorContext(ctx, "error in ListBoxes", slog.Any("error", err))
			continue
		}

		err = rn.reconcileMachine(ctx, boxes)
		if err != nil {
			slog.ErrorContext(ctx, "error in reconcileMachine", slog.Any("error", err))
			continue
		}
	}
}
