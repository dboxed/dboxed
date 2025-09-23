package hetzner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (r *Reconciler) reconcileHetznerNetwork(ctx context.Context) error {
	q := querier2.GetQuerier(ctx)
	if r.mp.Hetzner.HetznerNetworkName.V == "" {
		return fmt.Errorf("unexpected missing hetzner network name")
	}

	var err error
	var hetznerNetwork *hcloud.Network
	hetznerNetwork, _, err = r.hcloudClient.Network.GetByName(ctx, r.mp.Hetzner.HetznerNetworkName.V)
	if err != nil {
		return err
	}
	if hetznerNetwork == nil {
		return fmt.Errorf("network '%s' not found", r.mp.Hetzner.HetznerNetworkName.V)
	}

	log := r.log

	if len(hetznerNetwork.Subnets) == 0 {
		return fmt.Errorf("network '%s' has no subnets", r.mp.Hetzner.HetznerNetworkName.V)
	}

	networkZone := string(hetznerNetwork.Subnets[0].NetworkZone)
	var cloudCidr *string
	var robotCidr *string
	var robotVSwitchId *int64

	cloudCidr = util.Ptr(hetznerNetwork.Subnets[0].IPRange.String())
	if len(hetznerNetwork.Subnets) > 1 && hetznerNetwork.Subnets[1].Type == hcloud.NetworkSubnetTypeVSwitch {
		robotCidr = util.Ptr(hetznerNetwork.Subnets[1].IPRange.String())
		robotVSwitchId = &hetznerNetwork.Subnets[1].VSwitchID
	}

	status := dmodel.MachineProviderHetznerStatus{
		ID:                 querier2.N(r.mp.ID),
		HetznerNetworkID:   &hetznerNetwork.ID,
		HetznerNetworkZone: &networkZone,
		HetznerNetworkCidr: util.Ptr(hetznerNetwork.IPRange.String()),
		CloudSubnetCidr:    cloudCidr,
		RobotSubnetCidr:    robotCidr,
		RobotVswitchID:     robotVSwitchId,
	}
	if !util.EqualsViaJson(status, r.mp.Hetzner.Status) {
		log.InfoContext(ctx, "updating hetzner network in DB", slog.Any("params", util.MustJson(r.mp.Hetzner)))

		r.mp.Hetzner.Status = &status
		err = r.mp.Hetzner.Status.UpdateStatus(q)
		if err != nil {
			return err
		}
	}

	return nil
}
