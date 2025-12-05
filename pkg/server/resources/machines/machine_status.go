package machines

import (
	"context"
	"time"

	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (s *MachinesServer) restGetMachineStatus(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.MachineRunStatus], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeMachine, i.Id)
	if err != nil {
		return nil, err
	}

	machine, err := dmodel.GetMachineWithRunStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret := models.MachineRunStatusFromDB(machine.RunStatus)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *MachinesServer) restUpdateMachineStatus(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.UpdateMachineRunStatus]) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeMachine, i.Id)
	if err != nil {
		return nil, err
	}

	machine, err := dmodel.GetMachineWithRunStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	if machine.RunStatus == nil || !machine.RunStatus.ID.Valid {
		machine.RunStatus = &dmodel.MachineRunStatus{
			ID:         querier2.N(machine.ID),
			StatusTime: util.Ptr(time.Now()),
		}
		err = machine.RunStatus.Create(q)
		if err != nil {
			return nil, err
		}
	}

	oldStatusTime := machine.RunStatus.StatusTime

	if i.Body.RunStatus != nil {
		if i.Body.RunStatus != nil {
			err = machine.RunStatus.UpdateRunStatus(q, i.Body.RunStatus)
			if err != nil {
				return nil, err
			}
		}
		if i.Body.StartTime != nil {
			err = machine.RunStatus.UpdateStartTime(q, i.Body.StartTime)
			if err != nil {
				return nil, err
			}
		}

		if i.Body.StopTime != nil {
			err = machine.RunStatus.UpdateStopTime(q, i.Body.StopTime)
			if err != nil {
				return nil, err
			}
		}
	}

	if oldStatusTime != nil && machine.RunStatus.StatusTime != nil {
		// if we didn't update status for some time, do immediate reconciliation so that the overall machine status gets
		// updates asap
		if machine.RunStatus.StatusTime.Sub(*oldStatusTime) >= time.Second*30 {
			err = dmodel.AddChangeTracking(q, machine)
			if err != nil {
				return nil, err
			}
		}
	}

	return &huma_utils.Empty{}, nil
}
