package run_in_sandbox

import (
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunInSandbox) updateSandboxStatusSimple(status string) {
	rn.updateSandboxStatus(models.UpdateBoxSandboxStatus2{
		RunStatus: &status,
	})
}

func (rn *RunInSandbox) updateSandboxStatus(s models.UpdateBoxSandboxStatus2) {
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

	if util.EqualsViaJson(rn.sandboxStatus, rn.sandboxStatusWritten) {
		return
	}

	_ = util.AtomicWriteFileYaml(consts.SandboxStatusFile, rn.sandboxStatus, 0644)
}
