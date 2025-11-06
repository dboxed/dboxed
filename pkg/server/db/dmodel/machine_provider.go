package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type MachineProvider struct {
	OwnedByWorkspace
	ReconcileStatus

	Type string `db:"type"`
	Name string `db:"name"`

	SshKeyPublic *string `db:"ssh_key_public"`

	Aws     *MachineProviderAws     `join:"true"`
	Hetzner *MachineProviderHetzner `join:"true"`
}

func postprocessMachineProvider(q *querier2.Querier, mp *MachineProvider) error {
	switch mp.Type {
	case "aws":
		s, err := getMachineProviderAwsSubnets(q, mp.ID)
		if err != nil {
			return err
		}
		mp.Aws.Status.Subnets = s
	case "hetzner":
	}
	return nil
}

func GetMachineProviderById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*MachineProvider, error) {
	v, err := querier2.GetOne[MachineProvider](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	err = postprocessMachineProvider(q, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func ListMachineProviders(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]MachineProvider, error) {
	l, err := querier2.GetMany[MachineProvider](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
	if err != nil {
		return nil, err
	}

	var ret []MachineProvider
	for _, n := range l {
		err = postprocessMachineProvider(q, &n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func getMachineProviderAwsSubnets(q *querier2.Querier, machineProviderId string) ([]MachineProviderAwsSubnet, error) {
	return querier2.GetMany[MachineProviderAwsSubnet](q, map[string]any{
		"machine_provider_id": machineProviderId,
	}, nil)
}

func (v *MachineProvider) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineProvider) UpdateSshKeyPublic(q *querier2.Querier, k *string) error {
	v.SshKeyPublic = k
	return querier2.UpdateOneFromStruct(q, v, "ssh_key_public")
}
