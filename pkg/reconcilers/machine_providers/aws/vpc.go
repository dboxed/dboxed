package aws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

func (r *Reconciler) reconcileVpc(ctx context.Context) error {
	q := querier2.GetQuerier(ctx)

	if r.mp.Aws.VpcID == nil {
		return fmt.Errorf("unexpected missing aws vpc id")
	}

	vpcs, err := r.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{*r.mp.Aws.VpcID},
	})
	if err != nil {
		return err
	}
	if len(vpcs.Vpcs) == 0 {
		return fmt.Errorf("vpc not found")
	}
	vpc := vpcs.Vpcs[0]

	var vpcName *string
	for _, t := range vpc.Tags {
		if *t.Key == "Name" {
			vpcName = t.Value
		}
	}

	err = r.mp.Aws.Status.UpdateVpcInfo(q, vpcName, vpc.CidrBlock)
	if err != nil {
		return err
	}

	err = r.reconcileSubnets(ctx)
	if err != nil {
		return err
	}

	err = r.reconcileSecurityGroups(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileSubnets(ctx context.Context) error {
	q := querier2.GetQuerier(ctx)
	subnets, err := r.ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{{Name: util.Ptr("vpc-id"), Values: []string{*r.mp.Aws.VpcID}}},
	})
	if err != nil {
		return err
	}
	found := map[string]struct{}{}
	for _, subnet := range subnets.Subnets {
		var subnetName *string
		for _, t := range subnet.Tags {
			if *t.Key == "Name" {
				subnetName = t.Value
			}
		}

		log := r.log.With(slog.Any("subnetID", *subnet.SubnetId), slog.Any("subnetName", subnetName), slog.Any("subnetCidr", *subnet.CidrBlock))

		curSubnet := r.subnets[*subnet.SubnetId]

		newDboxedSubnet := &dmodel.MachineProviderAwsSubnet{
			MachineProviderID: querier2.N(r.mp.ID),
			SubnetID:          querier2.N(*subnet.SubnetId),
			SubnetName:        subnetName,
			AvailabilityZone:  querier2.N(*subnet.AvailabilityZone),
			Cidr:              querier2.N(*subnet.CidrBlock),
		}

		if curSubnet == nil {
			log.InfoContext(ctx, "creating aws subnet")
		} else {
			if util.EqualsViaJson(curSubnet, newDboxedSubnet) {
				found[*subnet.SubnetId] = struct{}{}
				continue
			}
			log.InfoContext(ctx, "updating aws subnet")
		}

		err = newDboxedSubnet.CreateOrUpdate(q)
		if err != nil {
			return err
		}
		r.subnets[*subnet.SubnetId] = newDboxedSubnet
		found[*subnet.SubnetId] = struct{}{}
	}
	for _, dboxedSubnet := range r.subnets {
		if _, ok := found[dboxedSubnet.SubnetID.V]; ok {
			continue
		}
		log := r.log.With(slog.Any("subnetID", dboxedSubnet.SubnetID), slog.Any("subnetName", dboxedSubnet.SubnetName), slog.Any("subnetCidr", dboxedSubnet.Cidr))

		log.InfoContext(ctx, "deleting aws subnet")

		err = dmodel.DeleteMachineProviderAwsSubnet(q, r.mp.ID, dboxedSubnet.SubnetID.V)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) reconcileSecurityGroups(ctx context.Context) error {
	q := querier2.GetQuerier(ctx)
	config := config.GetConfig(ctx)
	groupName := fmt.Sprintf("%s-nodes-%d", config.InstanceName, r.mp.Aws.ID.V)

	log := r.log.With(slog.Any("securityGroupName", groupName))

	resp, err := r.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: util.Ptr("vpc-id"), Values: []string{*r.mp.Aws.VpcID}},
			{Name: util.Ptr("group-name"), Values: []string{groupName}},
		},
	})
	if err != nil {
		return err
	}

	var sg *types.SecurityGroup
	if resp == nil || len(resp.SecurityGroups) == 0 {
		sg, err = r.createSecurityGroup(ctx, log, groupName)
		if err != nil {
			return err
		}
	} else {
		sg = &resp.SecurityGroups[0]
	}
	log = log.With(slog.Any("securityGroupId", *sg.GroupId))

	if r.mp.Aws.Status.SecurityGroupID == nil || *r.mp.Aws.Status.SecurityGroupID != *sg.GroupId {
		log.InfoContext(ctx, "updating security group id")
		err = r.mp.Aws.Status.UpdateSecurityGroupID(q, sg.GroupId)
		if err != nil {
			return err
		}
	}

	err = r.reconcileSecurityGroupRules(ctx, log, sg)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) createSecurityGroup(ctx context.Context, log *slog.Logger, groupName string) (*types.SecurityGroup, error) {
	var tags []types.Tag
	for k, v := range cloud_utils.BuildCloudBaseTags(r.mp.Aws.ID.V, r.mp.WorkspaceID) {
		tags = append(tags, types.Tag{Key: &k, Value: &v})
	}

	log.InfoContext(ctx, "creating security group")
	resp, err := r.ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: util.Ptr(fmt.Sprintf("dboxed security group for nodes in machine provider %d", r.mp.Aws.ID.V)),
		GroupName:   &groupName,
		VpcId:       r.mp.Aws.VpcID,
		TagSpecifications: []types.TagSpecification{
			{ResourceType: types.ResourceTypeSecurityGroup, Tags: tags},
		},
	})
	if err != nil {
		return nil, err
	}

	resp2, err := r.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{*resp.GroupId},
	})
	if err != nil {
		return nil, err
	}
	if len(resp2.SecurityGroups) == 0 {
		return nil, fmt.Errorf("failed to find security group even after creating it")
	}

	return &resp2.SecurityGroups[0], nil
}

func (r *Reconciler) reconcileSecurityGroupRules(ctx context.Context, log *slog.Logger, sg *types.SecurityGroup) error {
	outPermissions := []types.IpPermission{
		{
			IpProtocol:       util.Ptr("-1"),
			IpRanges:         []types.IpRange{{CidrIp: util.Ptr("0.0.0.0/0")}},
			Ipv6Ranges:       make([]types.Ipv6Range, 0),
			PrefixListIds:    make([]types.PrefixListId, 0),
			UserIdGroupPairs: make([]types.UserIdGroupPair, 0),
		},
	}

	inPermissions := []types.IpPermission{
		{
			// ssh
			IpProtocol:       util.Ptr("tcp"),
			FromPort:         util.Ptr(int32(22)),
			ToPort:           util.Ptr(int32(22)),
			IpRanges:         []types.IpRange{{CidrIp: util.Ptr("0.0.0.0/0")}},
			Ipv6Ranges:       make([]types.Ipv6Range, 0),
			PrefixListIds:    make([]types.PrefixListId, 0),
			UserIdGroupPairs: make([]types.UserIdGroupPair, 0),
		},
		{
			// tailscale
			IpProtocol:       util.Ptr("udp"),
			FromPort:         util.Ptr(int32(41641)),
			ToPort:           util.Ptr(int32(41641)),
			IpRanges:         []types.IpRange{{CidrIp: util.Ptr("0.0.0.0/0")}},
			Ipv6Ranges:       make([]types.Ipv6Range, 0),
			PrefixListIds:    make([]types.PrefixListId, 0),
			UserIdGroupPairs: make([]types.UserIdGroupPair, 0),
		},
	}

	a := reflect.DeepEqual(inPermissions, sg.IpPermissions)
	b := reflect.DeepEqual(outPermissions, sg.IpPermissionsEgress)

	if a && b {
		return nil
	}

	log.InfoContext(ctx, "updating security group rules")

	if len(sg.IpPermissionsEgress) != 0 {
		_, err := r.ec2Client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: sg.IpPermissionsEgress,
		})
		if err != nil {
			return err
		}
	}
	if len(sg.IpPermissions) != 0 {
		_, err := r.ec2Client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       sg.GroupId,
			IpPermissions: sg.IpPermissions,
		})
		if err != nil {
			return err
		}
	}

	_, err := r.ec2Client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       sg.GroupId,
		IpPermissions: outPermissions,
	})
	if err != nil {
		return err
	}
	_, err = r.ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       sg.GroupId,
		IpPermissions: inPermissions,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) deleteSecurityGroup(ctx context.Context) error {
	q := querier2.GetQuerier(ctx)
	if r.mp.Aws.Status.SecurityGroupID == nil {
		return nil
	}

	r.log.InfoContext(ctx, "deleting aws security group", slog.Any("securityGroupId", *r.mp.Aws.Status.SecurityGroupID))
	_, err := r.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: r.mp.Aws.Status.SecurityGroupID,
	})
	if err != nil {
		var err2 *smithy.GenericAPIError
		if !errors.As(err, &err2) || err2.Code != "InvalidGroup.NotFound" {
			return err
		}
	}

	err = r.mp.Aws.Status.UpdateSecurityGroupID(q, nil)
	if err != nil {
		return err
	}
	return nil
}
