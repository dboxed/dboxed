package box_spec_utils

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
)

func BuildBoxSpec(c context.Context, box *dmodel.Box, network *dmodel.Network) (*boxspec.BoxFile, error) {
	boxm, err := models.BoxFromDB(c, *box)
	if err != nil {
		return nil, err
	}

	volumeNames, err := validateBoxSpec(boxm.BoxSpec)
	if err != nil {
		return nil, err
	}

	err = buildAttachedVolumes(c, box, &boxm.BoxSpec, volumeNames)
	if err != nil {
		return nil, err
	}

	boxm.BoxSpec.Uuid = box.Uuid

	boxm.BoxSpec.Logs = &boxspec.LogsSpec{
		Nats: &boxspec.LogsNatsSpec{
			MetadataKVStore: nats_utils.BuildMetadataKVStoreName(c, box.WorkspaceID),
			LogStream:       nats_utils.BuildLogsStreamName(c, box.WorkspaceID),
			LogId:           fmt.Sprintf("box-%d", box.ID),
		},
	}

	if network != nil && box.NetworkType != nil {
		switch global.NetworkType(*box.NetworkType) {
		case global.NetworkNetbird:
			err = addNetbirdComposeProject(c, box, network, &boxm.BoxSpec)
			if err != nil {
				return nil, err
			}
		default:
			return nil, huma.Error400BadRequest(fmt.Sprintf("unknown network type %s", *box.NetworkType))
		}
	}

	file := &boxspec.BoxFile{
		Spec: boxm.BoxSpec,
	}

	return file, nil
}

func buildAttachedVolumes(ctx context.Context, box *dmodel.Box, boxSpec *boxspec.BoxSpec, volumeNames map[string]struct{}) error {
	q := querier.GetQuerier(ctx)

	ats, err := dmodel.ListBoxVolumeAttachments(q, box.ID)
	if err != nil {
		return err
	}
	volumeProviders := map[int64]*dmodel.VolumeProvider{}
	for _, at := range ats {
		if _, ok := volumeNames[at.Volume.Name]; ok {
			return huma.Error400BadRequest(fmt.Sprintf("attached dboxed volume %s clashes with volume from box spec", at.Volume.Name))
		}

		vp, ok := volumeProviders[at.Volume.VolumeProviderID]
		if !ok {
			vp, err = dmodel.GetVolumeProviderById(q, nil, at.Volume.VolumeProviderID, true)
			if err != nil {
				return err
			}
			volumeProviders[at.Volume.VolumeProviderID] = vp
		}

		switch dmodel.VolumeProviderType(vp.Type) {
		case dmodel.VolumeProviderTypeRustic:
			boxSpec.Volumes = append(boxSpec.Volumes, boxspec.BoxVolumeSpec{
				Name:     at.Volume.Name,
				RootUid:  uint32(at.RootUid.V),
				RootGid:  uint32(at.RootGid.V),
				RootMode: at.RootMode.V,
				Dboxed: &boxspec.DboxedVolume{
					VolumeId:       at.Volume.ID,
					BackupInterval: "1m",
				},
			})
		default:
			return huma.Error400BadRequest(fmt.Sprintf("not supported volume provider %s in attachment %s", vp.Type, at.Volume.Name))
		}
	}
	return nil
}

func validateBoxSpec(boxSpec boxspec.BoxSpec) (map[string]struct{}, error) {
	if boxSpec.Uuid != "" {
		return nil, huma.Error400BadRequest("uuid can not be set in box spec")
	}
	if boxSpec.Logs != nil {
		return nil, huma.Error400BadRequest("logs spec can not be set in box spec")
	}

	volumeNames := map[string]struct{}{}
	for _, v := range boxSpec.Volumes {
		if _, ok := volumeNames[v.Name]; ok {
			return nil, huma.Error400BadRequest(fmt.Sprintf("volume name %s not unique", v.Name))
		}
		volumeNames[v.Name] = struct{}{}
	}
	return volumeNames, nil
}
