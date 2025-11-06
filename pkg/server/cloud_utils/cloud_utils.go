package cloud_utils

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

const MachineProviderIdTagName = "dboxed.io/machine-provider-id"

func BuildCloudBaseTags(machineProviderId string, workspaceId string) map[string]string {
	m := map[string]string{
		MachineProviderIdTagName: fmt.Sprintf("%s", machineProviderId),
		"dboxed.io/workspace-id": fmt.Sprintf("%s", workspaceId),
	}
	return m
}

func BuildCloudMachineTags(machineProviderId string, machine *dmodel.Machine) map[string]string {
	m := map[string]string{
		"dboxed.io/machine-id":   fmt.Sprintf("%s", machine.ID),
		"dboxed.io/machine-name": machine.Name,
	}
	for k, v := range BuildCloudBaseTags(machineProviderId, machine.WorkspaceID) {
		m[k] = v
	}
	return m
}
