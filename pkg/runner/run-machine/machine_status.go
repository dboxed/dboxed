//go:build linux

package run_machine

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunMachine) updateMachineStatusSimple(ctx context.Context, status string, send bool) {
	rn.updateMachineStatus(ctx, models.UpdateMachineRunStatus{
		RunStatus: &status,
	}, send)
}

func (rn *RunMachine) updateMachineStatus(ctx context.Context, s models.UpdateMachineRunStatus, send bool) {
	rn.statusMutex.Lock()
	defer rn.statusMutex.Unlock()
	if s.RunStatus != nil {
		rn.machineStatus.RunStatus = s.RunStatus
	}
	if s.StartTime != nil {
		rn.machineStatus.StartTime = s.StartTime
	}
	if s.StopTime != nil {
		rn.machineStatus.StopTime = s.StopTime
	}
	if send {
		rn.sendMachineStatus(ctx, false)
	}
}

func (rn *RunMachine) startUpdateMachineStatusLoop(ctx context.Context) {
	rn.machineStatus = models.UpdateMachineRunStatus{
		RunStatus: util.Ptr("starting"),
		StartTime: util.Ptr(time.Now()),
	}

	rn.sendMachineStatus(ctx, true)

	rn.sendStatusStopCh = make(chan struct{})
	rn.sendStatusDone.Add(1)
	go func() {
		defer rn.sendStatusDone.Done()
		for {
			select {
			case <-rn.sendStatusStopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 10):
				rn.sendMachineStatus(ctx, true)
			}
		}
	}()
}

func (rn *RunMachine) stopUpdateMachineStatusLoop() {
	if rn.sendStatusStopCh != nil {
		close(rn.sendStatusStopCh)
		rn.sendStatusDone.Wait()
	}
}

func (rn *RunMachine) sendMachineStatus(ctx context.Context, lock bool) {
	if lock {
		rn.statusMutex.Lock()
		defer rn.statusMutex.Unlock()
	}
	if !time.Now().After(rn.machineStatusTime.Add(time.Second*30)) && util.EqualsViaJson(rn.machineStatus, rn.machineStatusSent) {
		return
	}

	machineClient := clients.MachineClient{Client: rn.Client}
	err := machineClient.UpdateMachineStatus(ctx, rn.MachineId, rn.machineStatus)
	if err != nil {
		slog.ErrorContext(ctx, "failed to report machine status", "error", err)
	} else {
		rn.machineStatusSent = rn.machineStatus
		rn.machineStatusTime = time.Now()
	}
}
