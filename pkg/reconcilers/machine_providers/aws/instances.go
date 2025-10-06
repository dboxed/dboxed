package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers/userdata"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

func (r *Reconciler) queryAwsInstances(ctx context.Context) error {
	r.awsInstancesByID = map[string]*types.Instance{}
	r.awsInstancesByName = map[string]*types.Instance{}
	l, err := cloud_utils.AwsAllHelper(&ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   util.Ptr("tag:" + cloud_utils.MachineProviderIdTagName),
				Values: []string{fmt.Sprintf("%d", r.mp.ID)},
			},
		},
		MaxResults: util.Ptr(int32(5)),
	}, func(params *ec2.DescribeInstancesInput) ([]types.Reservation, *string, error) {
		o, err := r.ec2Client.DescribeInstances(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return o.Reservations, o.NextToken, nil
	})
	if err != nil {
		return err
	}
	for _, reservation := range l {
		for _, instance := range reservation.Instances {
			r.awsInstancesByID[*instance.InstanceId] = &instance
			name := cloud_utils.AwsGetNameFromTags(instance.Tags)
			if name != nil {
				r.awsInstancesByName[*name] = &instance
			}
		}
	}

	return nil
}

func (r *Reconciler) ReconcileMachine(ctx context.Context, m *dmodel.Machine) error {
	q := querier.GetQuerier(ctx)
	log := r.log.With(slog.Any("awsInstanceType", m.Aws.InstanceType))
	if m.Aws.Status.InstanceID != nil {
		log = log.With(slog.Any("awsInstanceId", m.Aws.Status.InstanceID))
	}

	if m.DeletedAt.Valid {
		return r.reconcileDeleteAwsInstance(ctx, log, m)
	}

	if r.mp.Aws.Status.SecurityGroupID == nil {
		return fmt.Errorf("aws machine provider has no security group")
	}

	err := dmodel.AddFinalizer(q, m, "aws-machine")
	if err != nil {
		return err
	}

	if m.Aws.Status.InstanceID == nil {
		err := r.createAwsInstance(ctx, log, m)
		if err != nil {
			return err
		}
	}

	err = r.updateAwsInstance(ctx, log, m)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileDeleteAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) error {
	q := querier.GetQuerier(ctx)

	if m.Aws.Status.InstanceID == nil {
		err := dmodel.RemoveFinalizer(q, m, "aws-machine")
		if err != nil {
			return err
		}
		return nil
	}

	instance, ok := r.awsInstancesByID[*m.Aws.Status.InstanceID]
	if !ok {
		log.InfoContext(ctx, "aws instance vanished")
		err := dmodel.RemoveFinalizer(q, m, "aws-machine")
		if err != nil {
			return err
		}
		return nil
	}

	log.InfoContext(ctx, "deleting aws instance")
	_, err := r.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*m.Aws.Status.InstanceID},
	})
	if err != nil {
		return err
	}

	err = m.Aws.Status.UpdateInstanceID(q, nil)
	if err != nil {
		return err
	}

	name := cloud_utils.AwsGetNameFromTags(instance.Tags)
	delete(r.awsInstancesByID, *instance.InstanceId)
	if name != nil {
		delete(r.awsInstancesByID, *name)
	}

	err = dmodel.RemoveFinalizer(q, m, "aws-machine")
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) createAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) error {
	q := querier.GetQuerier(ctx)
	config := config.GetConfig(ctx)

	var sshKeyName *string
	if r.mp.SshKeyPublic != nil {
		sshKeyName = util.Ptr(cloud_utils.BuildAwsSshKeyName(ctx, r.mp.Name, r.mp.ID))
	}

	box, err := dmodel.GetBoxById(q, nil, m.BoxID, true)
	if err != nil {
		return err
	}
	ud := userdata.GetUserdata(
		box.DboxedVersion,
		"dummy",
		m.Name,
	)

	image, err := r.selectAwsImage(ctx, *m)
	if err != nil {
		return err
	}

	bdm, err := util.CopyViaJson(image.BlockDeviceMappings)
	if err != nil {
		return err
	}
	bdm[0].Ebs.VolumeSize = util.Ptr(int32(m.Aws.RootVolumeSize.V))

	networkInterfaces := []types.InstanceNetworkInterfaceSpecification{
		{
			AssociatePublicIpAddress: util.Ptr(true),
			DeleteOnTermination:      util.Ptr(true),
			Description:              util.Ptr(fmt.Sprintf("network interface for dboxed machine %s (id %d)", m.Name, m.ID)),
			DeviceIndex:              util.Ptr(int32(0)),
			SubnetId:                 &m.Aws.SubnetID.V,
			Groups:                   []string{*r.mp.Aws.Status.SecurityGroupID},
		},
	}

	log = log.With(slog.Any("amiId", image.ImageId))
	log.InfoContext(ctx, "creating aws instance")

	var tags []types.Tag
	for k, v := range cloud_utils.BuildCloudMachineTags(r.mp.ID, m) {
		tags = append(tags, types.Tag{Key: &k, Value: &v})
	}
	instanceName := fmt.Sprintf("%s-machine-%s", config.InstanceName, m.Name)
	tags = append(tags, types.Tag{Key: util.Ptr("Name"), Value: &instanceName})

	resp, err := r.ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:             image.ImageId,
		InstanceType:        types.InstanceType(m.Aws.InstanceType.V),
		KeyName:             sshKeyName,
		MinCount:            util.Ptr(int32(1)),
		MaxCount:            util.Ptr(int32(1)),
		UserData:            util.Ptr(base64.StdEncoding.EncodeToString([]byte(ud))),
		BlockDeviceMappings: bdm,
		NetworkInterfaces:   networkInterfaces,

		TagSpecifications: []types.TagSpecification{
			{ResourceType: types.ResourceTypeInstance, Tags: tags},
			{ResourceType: types.ResourceTypeVolume, Tags: tags},
			//{ResourceType: types.ResourceTypeSpotInstancesRequest, Tags: tags},
			{ResourceType: types.ResourceTypeNetworkInterface, Tags: tags},
		},
	})
	if err != nil {
		return err
	}
	instance := resp.Instances[0]

	r.awsInstancesByID[*instance.InstanceId] = &instance
	r.awsInstancesByName[instanceName] = &instance

	log = log.With(slog.Any("awsInstanceId", *instance.InstanceId))
	log.InfoContext(ctx, "created aws instance")

	err = m.Aws.Status.UpdateInstanceID(q, instance.InstanceId)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) updateAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) error {
	q := querier.GetQuerier(ctx)
	removeInstanceId := false
	instance, ok := r.awsInstancesByID[*m.Aws.Status.InstanceID]
	if !ok {
		log.InfoContext(ctx, "aws instance disappeared, removing instance id and expecting it to be re-created")
		removeInstanceId = true
	} else if instance.State.Name == types.InstanceStateNameTerminated {
		log.InfoContext(ctx, "aws instance terminated, removing instance id and expecting it to be re-created")
		removeInstanceId = true
	}

	if removeInstanceId {
		err := m.Aws.Status.UpdateInstanceID(q, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
