package cloud_utils

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

const MachineProviderIdTagName = "dboxed.io/machine-provider-id"

func BuildCloudBaseTags(machineProviderId int64, workspaceId int64) map[string]string {
	m := map[string]string{
		MachineProviderIdTagName: fmt.Sprintf("%d", machineProviderId),
		"dboxed.io/workspace-id": fmt.Sprintf("%d", workspaceId),
	}
	return m
}

func BuildCloudMachineTags(machineProviderId int64, machine *dmodel.Machine) map[string]string {
	m := map[string]string{
		"dboxed.io/machine-id":   fmt.Sprintf("%d", machine.ID),
		"dboxed.io/machine-name": machine.Name,
	}
	for k, v := range BuildCloudBaseTags(machineProviderId, machine.WorkspaceID) {
		m[k] = v
	}
	return m
}
