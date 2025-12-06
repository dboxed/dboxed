package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

type Box struct {
	OwnedByWorkspace
	ReconcileStatus

	Name    string `db:"name"`
	BoxType string `db:"box_type"`

	NetworkID   *string `db:"network_id"`
	NetworkType *string `db:"network_type"`

	MachineID *string `db:"machine_id"`

	Enabled              bool       `db:"enabled"`
	ReconcileRequestedAt *time.Time `db:"reconcile_requested_at"`

	Netbird *BoxNetbird `join:"true"`
}

type BoxSandboxStatus struct {
	ID querier2.NullForJoin[string] `db:"id"`

	StatusTime *time.Time `db:"status_time"`

	RunStatus *string    `db:"run_status"`
	StartTime *time.Time `db:"start_time"`
	StopTime  *time.Time `db:"stop_time"`

	DockerPs []byte `db:"docker_ps"`

	NetworkIP4 *string `db:"network_ip4"`
}

type BoxWithSandboxStatus struct {
	Box

	SandboxStatus *BoxSandboxStatus `db:"sandbox_status" join:"true" join_left_field:"id" join_right_table:"box_sandbox_status" join_right_field:"id"`
}

func (x *BoxWithSandboxStatus) GetTableName() string {
	return "box"
}

type BoxNetbird struct {
	ID querier2.NullForJoin[string] `db:"id"`
	ReconcileStatus

	SetupKey   *string `db:"setup_key"`
	SetupKeyID *string `db:"setup_key_id"`
}

func (v *BoxNetbird) GetId() string {
	return v.ID.V
}

type BoxComposeProject struct {
	BoxID          string `db:"box_id"`
	Name           string `db:"name"`
	ComposeProject string `db:"compose_project"`
}

type BoxPortForward struct {
	ID querier2.NullForJoin[string] `db:"id" uuid:"true"`
	Times

	BoxID       string  `db:"box_id"`
	Description *string `db:"description"`

	Protocol      string `db:"protocol"`
	HostPortFirst int    `db:"host_port_first"`
	HostPortLast  int    `db:"host_port_last"`
	SandboxPort   int    `db:"sandbox_port"`
}

func (v *Box) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *Box) UpdateEnabled(q *querier2.Querier, enabled bool) error {
	v.Enabled = enabled
	return querier2.UpdateOneFromStruct(q, v, "enabled")
}

func (v *Box) RequestReconcile(q *querier2.Querier) error {
	v.ReconcileRequestedAt = util.Ptr(time.Now())
	return querier2.UpdateOneFromStruct(q, v, "reconcile_requested_at")
}

func (v *Box) UpdateMachineID(q *querier2.Querier, machineId *string) error {
	v.MachineID = machineId
	return querier2.UpdateOneFromStruct(q, v, "machine_id")
}

func (v *BoxNetbird) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *BoxSandboxStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetBoxById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxWithSandboxStatusById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*BoxWithSandboxStatus, error) {
	return querier2.GetOne[BoxWithSandboxStatus](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxWithSandboxStatusByName(q *querier2.Querier, workspaceId string, name string, skipDeleted bool) (*BoxWithSandboxStatus, error) {
	return querier2.GetOne[BoxWithSandboxStatus](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxesWithSandboxStatusForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]BoxWithSandboxStatus, error) {
	return querier2.GetMany[BoxWithSandboxStatus](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListBoxesForNetwork(q *querier2.Querier, networkId string, skipDeleted bool) ([]BoxWithSandboxStatus, error) {
	return querier2.GetMany[BoxWithSandboxStatus](q, map[string]any{
		"network_id": networkId,
		"deleted_at": querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListBoxesForMachine(q *querier2.Querier, machineId string, skipDeleted bool) ([]BoxWithSandboxStatus, error) {
	return querier2.GetMany[BoxWithSandboxStatus](q, map[string]any{
		"machine_id": machineId,
		"deleted_at": querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func GetSandboxStatus(q *querier2.Querier, boxId string) (*BoxSandboxStatus, error) {
	return querier2.GetOne[BoxSandboxStatus](q, map[string]any{
		"id": boxId,
	})
}

func (v *BoxSandboxStatus) UpdateStatusTime(q *querier2.Querier) error {
	v.StatusTime = util.Ptr(time.Now())
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
	)
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

func (v *BoxSandboxStatus) UpdateNetworkIp4(q *querier2.Querier, ip4 *string) error {
	v.StatusTime = util.Ptr(time.Now())
	v.NetworkIP4 = ip4
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"network_ip4",
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

func ListBoxComposeProjects(q *querier2.Querier, boxId string) ([]BoxComposeProject, error) {
	return querier2.GetMany[BoxComposeProject](q, map[string]any{
		"box_id": boxId,
	}, nil)
}

func GetBoxComposeProjectByName(q *querier2.Querier, boxId string, name string) (*BoxComposeProject, error) {
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

func (v *BoxPortForward) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListBoxPortForwards(q *querier2.Querier, boxId string) ([]BoxPortForward, error) {
	return querier2.GetMany[BoxPortForward](q, map[string]any{
		"box_id": boxId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("id", querier2.SortOrderAsc),
	})
}

func GetBoxPortForward(q *querier2.Querier, boxId string, id string) (*BoxPortForward, error) {
	return querier2.GetOne[BoxPortForward](q, map[string]any{
		"box_id": boxId,
		"id":     id,
	})
}

func (v *BoxPortForward) Update(q *querier2.Querier, description *string, protocol *string, hostPortFirst *int, hostPortLast *int, sandboxPort *int) error {
	var fields []string
	if description != nil {
		v.Description = description
		fields = append(fields, "description")
	}
	if protocol != nil {
		v.Protocol = *protocol
		fields = append(fields, "protocol")
	}
	if hostPortFirst != nil {
		v.HostPortFirst = *hostPortFirst
		fields = append(fields, "host_port_first")
	}
	if hostPortLast != nil {
		v.HostPortLast = *hostPortLast
		fields = append(fields, "host_port_last")
	}
	if sandboxPort != nil {
		v.SandboxPort = *sandboxPort
		fields = append(fields, "sandbox_port")
	}
	if len(fields) == 0 {
		return nil
	}
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id": v.BoxID,
		"id":     v.ID.V,
	}, v, fields...)
}
