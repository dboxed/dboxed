package aws

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
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

func (r *Reconciler) reconcileCommon(log *slog.Logger, mp *dmodel.MachineProvider) base.ReconcileResult {
	r.log = log
	r.mp = mp

	r.subnets = map[string]*dmodel.MachineProviderAwsSubnet{}
	for _, subnet := range r.mp.Aws.Status.Subnets {
		r.subnets[subnet.SubnetID.V] = &subnet
	}

	r.buildAWSClients()

	if r.mp.Aws.VpcID != nil {
		r.log = slog.With(slog.Any("vpcID", *r.mp.Aws.VpcID))
	}
	if r.mp.Aws.Status.VpcName != nil {
		r.log = slog.With(slog.Any("vpcName", *r.mp.Aws.Status.VpcName))
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	result := r.reconcileCommon(log, mp)
	if result.Error != nil {
		return result
	}

	if mp.GetDeletedAt() != nil {
		return r.reconcileDeleteMachineProvider(ctx)
	}

	err := dmodel.AddFinalizer(q, mp, "cleanup-aws")
	if err != nil {
		return base.InternalError(err)
	}

	result = r.reconcileSshKey(ctx)
	if result.Error != nil {
		return result
	}

	result = r.reconcileVpc(ctx)
	if result.Error != nil {
		return result
	}

	result = r.queryAwsInstances(ctx)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) buildAWSClients() {
	r.ec2Client = cloud_utils.BuildAwsClient(cloud_utils.AwsCreds{
		AwsAccessKeyID:     r.mp.Aws.AwsAccessKeyID,
		AwsSecretAccessKey: r.mp.Aws.AwsSecretAccessKey,
		Region:             r.mp.Aws.Region.V,
	})
}

func (r *Reconciler) reconcileDeleteMachineProvider(ctx context.Context) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	if !r.mp.HasFinalizer("cleanup-aws") {
		return base.ReconcileResult{}
	}

	result := r.deleteSecurityGroup(ctx)
	if result.Error != nil {
		return result
	}

	result = r.deleteSshKeyPair(ctx)
	if result.Error != nil {
		return result
	}

	err := dmodel.RemoveFinalizer(q, r.mp, "cleanup-aws")
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}
