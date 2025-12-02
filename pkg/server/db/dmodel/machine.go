package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Machine struct {
	OwnedByWorkspace
	ReconcileStatus

	Name string `db:"name"`

	MachineProviderID   *string `db:"machine_provider_id"`
	MachineProviderType *string `db:"machine_provider_type"`
	MachineProvider     *MachineProvider

	Aws     *MachineAws     `join:"true"`
	Hetzner *MachineHetzner `join:"true"`
}

func (v *Machine) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetMachineById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*Machine, error) {
	return querier2.GetOne[Machine](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func listMachines(q *querier2.Querier, workspaceId *string, machineProviderId *string, skipDeleted bool) ([]Machine, error) {
	return querier2.GetMany[Machine](q, map[string]any{
		"workspace_id":        querier2.OmitIfNull(workspaceId),
		"machine_provider_id": querier2.OmitIfNull(machineProviderId),
		"deleted_at":          querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListMachinesForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]Machine, error) {
	return listMachines(q, &workspaceId, nil, skipDeleted)
}

func ListMachinesForMachineProvider(q *querier2.Querier, machineProviderId string, skipDeleted bool) ([]Machine, error) {
	return listMachines(q, nil, &machineProviderId, skipDeleted)
}
