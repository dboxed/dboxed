package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

type Box struct {
	OwnedByWorkspace
	ReconcileStatus

	Uuid string `db:"uuid"`
	Name string `db:"name"`

	NetworkID   *int64  `db:"network_id"`
	NetworkType *string `db:"network_type"`

	DboxedVersion string `db:"dboxed_version"`

	MachineID *int64 `db:"machine_id"`

	DesiredState string `db:"desired_state"`

	Netbird *BoxNetbird `join:"true"`
}

type BoxSandboxStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	StatusTime *time.Time `db:"status_time"`

	RunStatus *string    `db:"run_status"`
	StartTime *time.Time `db:"start_time"`
	StopTime  *time.Time `db:"stop_time"`

	DockerPs []byte `db:"docker_ps"`
}

type BoxNetbird struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	SetupKey   *string `db:"setup_key"`
	SetupKeyID *string `db:"setup_key_id"`
}

type BoxComposeProject struct {
	BoxID          int64  `db:"box_id"`
	Name           string `db:"name"`
	ComposeProject string `db:"compose_project"`
}

func (v *Box) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *Box) UpdateDesiredState(q *querier2.Querier, desiredState string) error {
	v.DesiredState = desiredState
	return querier2.UpdateOneFromStruct(q, v, "desired_state")
}

func (v *BoxNetbird) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *BoxSandboxStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetBoxById(q *querier2.Querier, workspaceId *int64, id int64, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxByName(q *querier2.Querier, workspaceId int64, name string, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxByUuid(q *querier2.Querier, workspaceId int64, uuid string, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"uuid":         uuid,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxesForWorkspace(q *querier2.Querier, workspaceId int64, skipDeleted bool) ([]Box, error) {
	return querier2.GetMany[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListBoxesForNetwork(q *querier2.Querier, networkId int64, skipDeleted bool) ([]Box, error) {
	return querier2.GetMany[Box](q, map[string]any{
		"network_id": networkId,
		"deleted_at": querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func GetSandboxStatus(q *querier2.Querier, boxId int64) (*BoxSandboxStatus, error) {
	return querier2.GetOne[BoxSandboxStatus](q, map[string]any{
		"id": boxId,
	})
}

func (v *BoxSandboxStatus) UpdateRunStatus(q *querier2.Querier, runStatus *string) error {
	v.StatusTime = util.Ptr(time.Now())
	v.RunStatus = runStatus
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"run_status",
	)
}

func (v *BoxSandboxStatus) UpdateStartTime(q *querier2.Querier, startTime *time.Time) error {
	v.StatusTime = util.Ptr(time.Now())
	v.StartTime = startTime
	v.StopTime = nil
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"start_time",
		"stop_time",
	)
}

func (v *BoxSandboxStatus) UpdateStopTime(q *querier2.Querier, stopTime *time.Time) error {
	v.StatusTime = util.Ptr(time.Now())
	v.StopTime = stopTime
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"stop_time",
	)
}

func (v *BoxSandboxStatus) UpdateDockerPs(q *querier2.Querier, dockerPs []byte) error {
	v.StatusTime = util.Ptr(time.Now())
	v.DockerPs = dockerPs
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"docker_ps",
	)
}

func (v *BoxNetbird) UpdateSetupKey(q *querier2.Querier, setupKey *string, setupKeyId *string) error {
	v.SetupKey = setupKey
	v.SetupKeyID = setupKeyId
	return querier2.UpdateOneFromStruct(q, v,
		"setup_key",
		"setup_key_id",
	)
}

func (v *BoxComposeProject) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListBoxComposeProjects(q *querier2.Querier, boxId int64) ([]BoxComposeProject, error) {
	return querier2.GetMany[BoxComposeProject](q, map[string]any{
		"box_id": boxId,
	}, nil)
}

func GetBoxComposeProjectByName(q *querier2.Querier, boxId int64, name string) (*BoxComposeProject, error) {
	return querier2.GetOne[BoxComposeProject](q, map[string]any{
		"box_id": boxId,
		"name":   name,
	})
}

func (v *BoxComposeProject) UpdateComposeProject(q *querier2.Querier, composeProject string) error {
	v.ComposeProject = composeProject
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id": v.BoxID,
		"name":   v.Name,
	}, v, "compose_project")
}
