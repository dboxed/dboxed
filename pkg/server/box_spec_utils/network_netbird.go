package box_spec_utils

import (
	"context"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	box_specs "github.com/dboxed/dboxed/pkg/server/box_spec_utils/box-specs"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

func addNetbirdComposeProject(ctx context.Context, box *dmodel.Box, network *dmodel.Network, boxSpec *boxspec.BoxSpec) error {
	composeProject := &ctypes.Project{
		Name: "dboxed-netbird",
	}
	err := box_specs.AddNetbirdService(*network.Netbird, box, composeProject)
	if err != nil {
		return err
	}

	b, err := composeProject.MarshalYAML()
	if err != nil {
		return err
	}

	boxSpec.ComposeProjects = append([]string{string(b)}, boxSpec.ComposeProjects...)

	return nil
}
