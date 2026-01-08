package machine_providers

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"go4.org/netipx"
)

func (s *MachineProviderServer) restCreateMachineProviderHetzner(c context.Context, log *slog.Logger, mp *dmodel.MachineProvider, body *models.CreateMachineProviderHetzner) error {
	q := querier2.GetQuerier(c)
	if body.CloudToken == "" {
		return huma.Error400BadRequest("cloud token must be provided")
	}
	if body.RobotUsername != nil || body.RobotPassword != nil && (body.RobotUsername == nil || body.RobotPassword == nil) {
		return huma.Error400BadRequest("either both of robot_username/robot_password must be set or none of them")
	}

	if body.HetznerNetworkName == "" {
		return huma.Error400BadRequest("must specify Hetzner network name")
	}
	if err := util.CheckName(body.HetznerNetworkName); err != nil {
		return huma.Error400BadRequest(err.Error())
	}

	log.InfoContext(c, "creating machine_provider_hetzner")

	mp.Hetzner = &dmodel.MachineProviderHetzner{
		ID:                 querier2.N(mp.ID),
		HcloudToken:        querier2.N(body.CloudToken),
		HetznerNetworkName: querier2.N(body.HetznerNetworkName),
		Status: &dmodel.MachineProviderHetznerStatus{
			ID: querier2.N(mp.ID),
		},
	}

	if body.RobotUsername != nil {
		mp.Hetzner.RobotUser = body.RobotUsername
		mp.Hetzner.RobotPassword = body.RobotPassword
	}

	err := mp.Hetzner.Create(q)
	if err != nil {
		return err
	}

	err = mp.Hetzner.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *MachineProviderServer) restUpdateMachineProviderHetzner(c context.Context, log *slog.Logger, mp *dmodel.MachineProvider, body *models.UpdateMachineProviderHetzner) error {
	q := querier2.GetQuerier(c)

	if body.CloudToken != nil {
		if *body.CloudToken == "" {
			return huma.Error400BadRequest("empty cloud token is not allowed")
		}
		err := mp.Hetzner.UpdateHCloudToken(q, *body.CloudToken)
		if err != nil {
			return err
		}
	}

	if body.RobotUsername != nil || body.RobotPassword != nil {
		if body.RobotUsername == nil || body.RobotPassword == nil {
			return huma.Error400BadRequest("either both of robot_username/robot_password must be set or none of them")
		}
		u := strings.TrimSpace(*body.RobotUsername)
		p := strings.TrimSpace(*body.RobotPassword)
		var updateUser, updatePassword *string
		if u == "" || p == "" {
			// trying to remove login
			if u != "" || p != "" {
				return huma.Error400BadRequest("either both of robot_username/robot_password must be empty or none of them")
			}
			updateUser = nil
			updatePassword = nil
		} else {
			updateUser = &u
			updatePassword = &p
		}

		err := mp.Hetzner.UpdateRobotCredentials(q, updateUser, updatePassword)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MachineProviderServer) checkSubnets(networkCidr *string, cloudSubnetCidr *string, robotSubnetCidr *string) error {
	networkCidr2, err := netip.ParsePrefix(*networkCidr)
	if err != nil {
		return huma.Error400BadRequest("invalid cidr")
	}

	if !networkCidr2.Addr().IsPrivate() {
		return huma.Error400BadRequest("only private CIDRs are allowed")
	}

	cloudSubnetCidr2, err := netip.ParsePrefix(*cloudSubnetCidr)
	if err != nil {
		return huma.Error400BadRequest("invalid cidr")
	}
	cloudSubnetCidrRange := netipx.RangeOfPrefix(cloudSubnetCidr2)

	if !networkCidr2.Contains(cloudSubnetCidrRange.From()) || !networkCidr2.Contains(cloudSubnetCidrRange.To()) {
		return huma.Error400BadRequest("cloud subnet CIDR not part of network CIDR")
	}

	if robotSubnetCidr != nil {
		robotSubnetCidr2, err := netip.ParsePrefix(*robotSubnetCidr)
		if err != nil {
			return huma.Error400BadRequest("invalid cidr")
		}
		robotSubnetCidrRange := netipx.RangeOfPrefix(robotSubnetCidr2)
		if !networkCidr2.Contains(robotSubnetCidrRange.From()) || !networkCidr2.Contains(robotSubnetCidrRange.To()) {
			return huma.Error400BadRequest("cloud subnet CIDR not part of network CIDR")
		}
		if cloudSubnetCidr2.Overlaps(robotSubnetCidr2) {
			return huma.Error400BadRequest("robot and cloud subnet CIDRs overlap")
		}
	}
	return nil
}

func (s *MachineProviderServer) restListHetznerServerTypes(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.HetznerServerType], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	mp, err := dmodel.GetMachineProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	if mp.Type != dmodel.MachineProviderTypeHetzner {
		return nil, huma.Error400BadRequest("machine provider is not a hetzner provider")
	}

	hcloudClient := hcloud.NewClient(hcloud.WithToken(mp.Hetzner.HcloudToken.V))

	l, _, err := hcloudClient.ServerType.List(c, hcloud.ServerTypeListOpts{})
	if err != nil {
		return nil, huma.Error400BadRequest(fmt.Sprintf("failed to retrieve hetzner server types: %s", err.Error()), err)
	}

	var ret []models.HetznerServerType
	for _, x := range l {
		e := models.HetznerServerType{
			ID:           x.ID,
			Name:         x.Name,
			Description:  x.Description,
			Category:     x.Category,
			Cores:        x.Cores,
			Memory:       x.Memory,
			Disk:         x.Disk,
			StorageType:  string(x.StorageType),
			CPUType:      string(x.CPUType),
			Architecture: string(x.Architecture),
		}
		for _, p := range x.Pricings {
			e.Pricings = append(e.Pricings, models.HetznerServerTypeLocationPricing{
				Location:        p.Location.Name,
				Hourly:          convertHetznerPrice(p.Hourly),
				Monthly:         convertHetznerPrice(p.Monthly),
				IncludedTraffic: p.IncludedTraffic,
				PerTBTraffic:    convertHetznerPrice(p.PerTBTraffic),
			})
		}
		for _, l := range x.Locations {
			e2 := models.HetznerServerTypeLocation{
				Location: l.Location.Name,
			}
			if l.DeprecatableResource.Deprecation != nil {
				e2.Deprecation = &models.HetznerDeprecation{
					Announced:        l.Deprecation.Announced,
					UnavailableAfter: l.Deprecation.UnavailableAfter,
				}
			}
			e.Locations = append(e.Locations, e2)
		}
		ret = append(ret, e)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

func convertHetznerPrice(p hcloud.Price) models.HetznerPrice {
	return models.HetznerPrice{
		Net:   p.Net,
		Gross: p.Gross,
	}
}
