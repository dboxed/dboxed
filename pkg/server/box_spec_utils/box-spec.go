package box_spec_utils

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
)

func BuildBoxSpec(c context.Context, box *dmodel.Box, network *dmodel.Network) (*boxspec.BoxSpec, error) {
	q := querier.GetQuerier(c)

	boxSpec := &boxspec.BoxSpec{
		Uuid:            box.Uuid,
		ComposeProjects: map[string]string{},
	}

	err := buildAttachedVolumes(c, box, boxSpec)
	if err != nil {
		return nil, err
	}

	bcps, err := dmodel.ListBoxComposeProjects(q, box.ID)
	if err != nil {
		return nil, err
	}
	for _, bcp := range bcps {
		boxSpec.ComposeProjects[bcp.Name] = bcp.ComposeProject
	}

	if network != nil && box.NetworkType != nil {
		switch global.NetworkType(*box.NetworkType) {
		case global.NetworkNetbird:
			err = addNetbirdComposeProject(c, box, network, boxSpec)
			if err != nil {
				return nil, err
			}
		default:
			return nil, huma.Error400BadRequest(fmt.Sprintf("unknown network type %s", *box.NetworkType))
		}
	}

	return boxSpec, nil
}

func buildAttachedVolumes(ctx context.Context, box *dmodel.Box, boxSpec *boxspec.BoxSpec) error {
	q := querier.GetQuerier(ctx)

	ats, err := dmodel.ListBoxVolumeAttachments(q, box.ID)
	if err != nil {
		return err
	}
	volumeProviders := map[int64]*dmodel.VolumeProvider{}
	for _, at := range ats {
		vp, ok := volumeProviders[at.Volume.VolumeProviderID]
		if !ok {
			vp, err = dmodel.GetVolumeProviderById(q, nil, at.Volume.VolumeProviderID, true)
			if err != nil {
				return err
			}
			volumeProviders[at.Volume.VolumeProviderID] = vp
		}

		boxSpec.Volumes = append(boxSpec.Volumes, boxspec.DboxedVolume{
			Uuid:           at.Volume.Uuid,
			Name:           at.Volume.Name,
			Id:             at.Volume.ID,
			RootUid:        uint32(at.RootUid.V),
			RootGid:        uint32(at.RootGid.V),
			RootMode:       at.RootMode.V,
			BackupInterval: "1m",
		})
	}
	return nil
}
