package aws

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Reconciler struct {
	log *slog.Logger

	mp *dmodel.MachineProvider

	subnets map[string]*dmodel.MachineProviderAwsSubnet

	awsImages        []types.Image
	awsInstanceTypes []types.InstanceTypeInfo

	awsInstancesByID   map[string]*types.Instance
	awsInstancesByName map[string]*types.Instance

	ec2Client *ec2.Client
}

func (r *Reconciler) reconcileCommon(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error {
	r.log = log
	r.mp = mp

	r.subnets = map[string]*dmodel.MachineProviderAwsSubnet{}
	for _, subnet := range r.mp.Aws.Status.Subnets {
		r.subnets[subnet.SubnetID.V] = &subnet
	}

	err := r.buildAWSClients(ctx)
	if err != nil {
		return err
	}

	if r.mp.Aws.VpcID != nil {
		r.log = slog.With(slog.Any("vpcID", *r.mp.Aws.VpcID))
	}
	if r.mp.Aws.Status.VpcName != nil {
		r.log = slog.With(slog.Any("vpcName", *r.mp.Aws.Status.VpcName))
	}

	return nil
}

func (r *Reconciler) ReconcileMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error {
	err := r.reconcileCommon(ctx, log, mp)
	if err != nil {
		return err
	}

	err = r.reconcileSshKey(ctx)
	if err != nil {
		return err
	}

	err = r.reconcileVpc(ctx)
	if err != nil {
		return err
	}

	err = r.queryAwsInstances(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildAWSClients(ctx context.Context) error {
	r.ec2Client = cloud_utils.BuildAwsClient(cloud_utils.AwsCreds{
		AwsAccessKeyID:     r.mp.Aws.AwsAccessKeyID,
		AwsSecretAccessKey: r.mp.Aws.AwsSecretAccessKey,
		Region:             r.mp.Aws.Region.V,
	})
	return nil
}

func (r *Reconciler) ReconcileDeleteMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error {
	err := r.reconcileCommon(ctx, log, mp)
	if err != nil {
		return err
	}

	err = r.deleteSecurityGroup(ctx)
	if err != nil {
		return err
	}

	err = r.deleteSshKeyPair(ctx)
	if err != nil {
		return err
	}

	return nil
}
