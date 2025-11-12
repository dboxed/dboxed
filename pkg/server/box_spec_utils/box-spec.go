package box_spec_utils

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
)

func BuildBoxSpec(c context.Context, box *dmodel.Box, network *dmodel.Network) (*boxspec.BoxSpec, error) {
	q := querier.GetQuerier(c)
	cfg := config.GetConfig(c)

	boxSpec := &boxspec.BoxSpec{
		ID:              box.ID,
		DesiredState:    box.DesiredState,
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

	portForwards, err := dmodel.ListBoxPortForwards(q, box.ID)
	if err != nil {
		return nil, err
	}

	boxSpec.Network = &boxspec.BoxNetwork{}
	for _, pf := range portForwards {
		boxSpec.Network.PortForwards = append(boxSpec.Network.PortForwards, boxspec.PortForward{
			Protocol:      pf.Protocol,
			HostFirstPort: pf.HostPortFirst,
			HostLastPort:  pf.HostPortLast,
			SandboxPort:   pf.SandboxPort,
		})
	}

	if network != nil && box.NetworkType != nil {
		switch global.NetworkType(*box.NetworkType) {
		case global.NetworkNetbird:
			if box.Netbird.SetupKey == nil {
				return nil, fmt.Errorf("box %s has no setup key", box.ID)
			}
			boxSpec.Network.Netbird = &boxspec.BoxNetworkNetbird{
				ManagementUrl: network.Netbird.ApiUrl.V,
				SetupKey:      *box.Netbird.SetupKey,
				Hostname:      fmt.Sprintf("%s-%s", cfg.InstanceName, box.ID),
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
	volumeProviders := map[string]*dmodel.VolumeProvider{}
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
			ID:             at.Volume.ID,
			Name:           at.Volume.Name,
			RootUid:        uint32(at.RootUid.V),
			RootGid:        uint32(at.RootGid.V),
			RootMode:       at.RootMode.V,
			BackupInterval: "1m",
		})
	}
	return nil
}
