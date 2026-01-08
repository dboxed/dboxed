package machine_providers

import (
	"context"
	"log/slog"
	"net/netip"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
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
