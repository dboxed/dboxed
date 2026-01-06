package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

type BoxSandbox struct {
	OwnedByWorkspaceOrNull
	BoxID querier2.NullForJoin[string] `db:"box_id"`

	MachineId querier2.NullForJoin[string] `db:"machine_id"`
	Hostname  querier2.NullForJoin[string] `db:"hostname"`

	StatusTime *time.Time `db:"status_time"`

	RunStatus *string    `db:"run_status"`
	StartTime *time.Time `db:"start_time"`
	StopTime  *time.Time `db:"stop_time"`

	DockerPs []byte `db:"docker_ps"`

	NetworkIP4 *string `db:"network_ip4"`
}

type BoxWithFullSandbox struct {
	Box

	Sandbox *BoxSandbox `db:"sandbox" join:"true" join_left_field:"current_sandbox_id" join_right_table:"box_sandbox" join_right_field:"id"`
}

func (x *BoxWithFullSandbox) GetTableName() string {
	return "box"
}

type BoxWithSandbox struct {
	BoxWithFullSandbox
}

func (x *BoxWithSandbox) GetOmittedColumns() []string {
	return []string{"sandbox.docker_ps"}
}

type BoxSandboxOnlyNetworkIP4 struct {
	ID querier2.NullForJoin[string] `db:"id"`

	NetworkIP4 *string `db:"network_ip4"`
}

func (v *BoxSandbox) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListSandboxesByBox(q *querier2.Querier, boxId string) ([]BoxSandbox, error) {
	return querier2.GetMany[BoxSandbox](q, map[string]any{
		"box_id": boxId,
	}, nil)
}

func ListSandboxesByWorkspace(q *querier2.Querier, workspaceId string) ([]BoxSandbox, error) {
	return querier2.GetMany[BoxSandbox](q, map[string]any{
		"workspace_id": workspaceId,
	}, nil)
}

func GetSandboxById(q *querier2.Querier, workspaceId *string, boxId *string, id string) (*BoxSandbox, error) {
	return querier2.GetOne[BoxSandbox](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"box_id":       querier2.OmitIfNull(boxId),
		"id":           id,
	})
}
func (v *BoxSandbox) UpdateStatusTime(q *querier2.Querier) error {
	v.StatusTime = util.Ptr(time.Now())
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
	)
}

func (v *BoxSandbox) UpdateStatus(q *querier2.Querier, runStatus *string, startTime *time.Time, stopTime *time.Time, networkIp4 *string) error {
	var fields []string
	if runStatus != nil {
		fields = append(fields, "run_status")
		v.RunStatus = runStatus
	}
	if startTime != nil {
		fields = append(fields, "start_time", "stop_time")
		v.StartTime = startTime
		v.StopTime = stopTime
	} else if stopTime != nil {
		fields = append(fields, "stop_time")
		v.StopTime = stopTime
	}
	if networkIp4 != nil {
		fields = append(fields, "network_ip4")
		v.NetworkIP4 = networkIp4
	}

	if len(fields) == 0 {
		return nil
	}

	fields = append(fields, "status_time")
	v.StatusTime = util.Ptr(time.Now())

	return querier2.UpdateOneFromStruct(q, v, fields...)
}

func (v *BoxSandbox) UpdateDockerPs(q *querier2.Querier, dockerPs []byte) error {
	v.StatusTime = util.Ptr(time.Now())
	v.DockerPs = dockerPs
	return querier2.UpdateOneFromStruct(q, v,
		"status_time",
		"docker_ps",
	)
}
