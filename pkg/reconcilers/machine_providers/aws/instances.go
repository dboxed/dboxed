package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers/userdata"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/resources/machines"
	"github.com/dboxed/dboxed/pkg/util"
)

func (r *Reconciler) queryAwsInstances(ctx context.Context) base.ReconcileResult {
	r.awsInstancesByID = map[string]*types.Instance{}
	r.awsInstancesByName = map[string]*types.Instance{}
	l, err := cloud_utils.AwsAllHelper(&ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   util.Ptr("tag:" + cloud_utils.MachineProviderIdTagName),
				Values: []string{r.mp.ID},
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
		return base.ErrorWithMessage(err, "failed to describe AWS instances: %s", err.Error())
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

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileMachine(ctx context.Context, log *slog.Logger, m *dmodel.Machine) {
	result := r.doReconcileMachine(ctx, m)
	base.LogReconcileResultError(ctx, log, result)
	base.SetReconcileResult(ctx, log, m.Aws, result)
}

func (r *Reconciler) doReconcileMachine(ctx context.Context, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	log := r.log.With(slog.Any("awsInstanceType", m.Aws.InstanceType))
	if m.Aws.Status.InstanceID != nil {
		log = log.With(slog.Any("awsInstanceId", m.Aws.Status.InstanceID))
	}

	if m.GetDeletedAt() != nil {
		return r.reconcileDeleteAwsInstance(ctx, log, m)
	}

	if r.mp.Aws.Status.SecurityGroupID == nil {
		return base.ErrorFromMessage("aws machine provider has no security group")
	}

	err := dmodel.AddFinalizer(q, m, "aws-machine")
	if err != nil {
		return base.InternalError(err)
	}

	if m.Aws.Status.InstanceID == nil {
		result := r.createAwsInstance(ctx, log, m)
		if result.ExitReconcile() {
			return result
		}
	}

	result := r.updateAwsInstance(ctx, log, m)
	if result.ExitReconcile() {
		return result
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) reconcileDeleteAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	if m.Aws.Status.InstanceID == nil {
		err := dmodel.RemoveFinalizer(q, m, "aws-machine")
		if err != nil {
			return base.InternalError(err)
		}
		return base.ReconcileResult{}
	}

	instance, ok := r.awsInstancesByID[*m.Aws.Status.InstanceID]
	if !ok {
		log.InfoContext(ctx, "aws instance vanished")
		err := dmodel.RemoveFinalizer(q, m, "aws-machine")
		if err != nil {
			return base.InternalError(err)
		}
		return base.ReconcileResult{}
	}

	log.InfoContext(ctx, "deleting aws instance")
	_, err := r.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{*m.Aws.Status.InstanceID},
	})
	if err != nil {
		return base.ErrorWithMessage(err, "failed to terminate AWS instance %s: %s", *m.Aws.Status.InstanceID, err.Error())
	}

	err = m.Aws.Status.UpdateInstanceID(q, nil)
	if err != nil {
		return base.InternalError(err)
	}

	name := cloud_utils.AwsGetNameFromTags(instance.Tags)
	delete(r.awsInstancesByID, *instance.InstanceId)
	if name != nil {
		delete(r.awsInstancesByID, *name)
	}

	err = dmodel.RemoveFinalizer(q, m, "aws-machine")
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) createAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	config := config.GetConfig(ctx)

	var sshKeyName *string
	if r.mp.SshKeyPublic != nil {
		sshKeyName = util.Ptr(cloud_utils.BuildAwsSshKeyName(ctx, r.mp.Name, r.mp.ID))
	}

	token, err := machines.GetFirstValidMachineToken(ctx, m.WorkspaceID, m.ID)
	if err != nil {
		return base.InternalError(err)
	}

	ud := userdata.GetUserdata(
		m.DboxedVersion,
		config.Server.BaseUrl,
		token.Token,
		m.ID,
	)

	image, result := r.selectAwsImage(ctx, *m)
	if result.ExitReconcile() {
		return result
	}

	bdm, err := util.CopyViaJson(image.BlockDeviceMappings)
	if err != nil {
		return base.InternalError(err)
	}
	bdm[0].Ebs.VolumeSize = util.Ptr(int32(m.Aws.RootVolumeSize.V))

	networkInterfaces := []types.InstanceNetworkInterfaceSpecification{
		{
			AssociatePublicIpAddress: util.Ptr(true),
			DeleteOnTermination:      util.Ptr(true),
			Description:              util.Ptr(fmt.Sprintf("network interface for dboxed machine %s (id %s)", m.Name, m.ID)),
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
		return base.ErrorWithMessage(err, "failed to run AWS instance: %s", err.Error())
	}
	instance := resp.Instances[0]

	r.awsInstancesByID[*instance.InstanceId] = &instance
	r.awsInstancesByName[instanceName] = &instance

	log = log.With(slog.Any("awsInstanceId", *instance.InstanceId))
	log.InfoContext(ctx, "created aws instance")

	err = m.Aws.Status.UpdateInstanceID(q, instance.InstanceId)
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) updateAwsInstance(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
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
			return base.InternalError(err)
		}
	}

	if !util.EqualsViaJson(instance.PublicIpAddress, m.Aws.Status.PublicIp4) {
		err := m.Aws.Status.UpdatePublicIP4(q, instance.PublicIpAddress)
		if err != nil {
			return base.InternalError(err)
		}
	}

	return base.ReconcileResult{}
}
