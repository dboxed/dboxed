package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

type Machine struct {
	OwnedByWorkspace
	ReconcileStatus

	Name string `db:"name"`

	DboxedVersion string `db:"dboxed_version"`

	MachineProviderID   *string              `db:"machine_provider_id"`
	MachineProviderType *MachineProviderType `db:"machine_provider_type"`
	MachineProvider     *MachineProvider

	Aws     *MachineAws     `join:"true"`
	Hetzner *MachineHetzner `join:"true"`
}

type MachineRunStatus struct {
	ID querier2.NullForJoin[string] `db:"id"`

	StatusTime *time.Time `db:"status_time"`

	RunStatus *string    `db:"run_status"`
	StartTime *time.Time `db:"start_time"`
	StopTime  *time.Time `db:"stop_time"`
}

type MachineWithRunStatus struct {
	Machine

	RunStatus *MachineRunStatus `db:"run_status" join:"true" join_left_field:"id" join_right_table:"machine_run_status" join_right_field:"id"`
}

func (x *MachineWithRunStatus) GetTableName() string {
	return "machine"
}

func (v *Machine) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineRunStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetMachineById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*Machine, error) {
	return querier2.GetOne[Machine](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetMachineWithRunStatusById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*MachineWithRunStatus, error) {
	return querier2.GetOne[MachineWithRunStatus](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetMachineRunStatusById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*MachineWithRunStatus, error) {
	return querier2.GetOne[MachineWithRunStatus](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func listMachines[T any](q *querier2.Querier, workspaceId *string, machineProviderId *string, skipDeleted bool) ([]T, error) {
	return querier2.GetMany[T](q, map[string]any{
		"workspace_id":        querier2.OmitIfNull(workspaceId),
		"machine_provider_id": querier2.OmitIfNull(machineProviderId),
		"deleted_at":          querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListMachinesForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]Machine, error) {
	return listMachines[Machine](q, &workspaceId, nil, skipDeleted)
}

func ListMachinesWithRunStatusForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]MachineWithRunStatus, error) {
	return listMachines[MachineWithRunStatus](q, &workspaceId, nil, skipDeleted)
}

func ListMachinesForMachineProvider(q *querier2.Querier, machineProviderId string, skipDeleted bool) ([]Machine, error) {
	return listMachines[Machine](q, nil, &machineProviderId, skipDeleted)
}

func (v *MachineRunStatus) UpdateRunStatus(q *querier2.Querier, runStatus *string) error {
	v.StatusTime = util.Ptr(time.Now())
	v.RunStatus = runStatus
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"run_status",
	)
}

func (v *MachineRunStatus) UpdateStartTime(q *querier2.Querier, startTime *time.Time) error {
	v.StatusTime = util.Ptr(time.Now())
	v.StartTime = startTime
	v.StopTime = nil
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"start_time",
		"stop_time",
	)
}

func (v *MachineRunStatus) UpdateStopTime(q *querier2.Querier, stopTime *time.Time) error {
	v.StatusTime = util.Ptr(time.Now())
	v.StopTime = stopTime
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"stop_time",
	)
}
