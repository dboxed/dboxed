package aws

import (
	"context"
	"slices"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

// see https://wiki.debian.org/Cloud/AmazonEC2Image/Trixie
const debianAccount = "136693071363"

func (r *Reconciler) describeAwsInstanceTypes(ctx context.Context) ([]types.InstanceTypeInfo, base.ReconcileResult) {
	if r.awsInstanceTypes != nil {
		return r.awsInstanceTypes, base.ReconcileResult{}
	}

	var nextToken *string
	for {
		resp, err := r.ec2Client.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, base.ErrorWithMessage(err, "failed to describe instance types: %s", err.Error())
		}
		r.awsInstanceTypes = append(r.awsInstanceTypes, resp.InstanceTypes...)
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	return r.awsInstanceTypes, base.ReconcileResult{}
}

func (r *Reconciler) describeAwsImages(ctx context.Context) ([]types.Image, base.ReconcileResult) {
	if r.awsImages != nil {
		return r.awsImages, base.ReconcileResult{}
	}

	var nextToken *string
	for {
		resp, err := r.ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Owners: []string{debianAccount},
			Filters: []types.Filter{
				{
					Name:   util.Ptr("name"),
					Values: []string{"debian-13-arm64-*", "debian-13-amd64-*"},
				},
				{
					Name:   util.Ptr("state"),
					Values: []string{"available"},
				},
			},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, base.ErrorWithMessage(err, "failed to describe images: %s", err.Error())
		}
		r.awsImages = append(r.awsImages, resp.Images...)
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	return r.awsImages, base.ReconcileResult{}
}

func (r *Reconciler) selectAwsImage(ctx context.Context, m dmodel.Machine) (*types.Image, base.ReconcileResult) {
	var wg sync.WaitGroup

	var instanceTypes []types.InstanceTypeInfo
	var images []types.Image

	var result1, result2 base.ReconcileResult

	wg.Add(2)
	go func() {
		defer wg.Done()
		instanceTypes, result1 = r.describeAwsInstanceTypes(ctx)
	}()
	go func() {
		defer wg.Done()
		images, result2 = r.describeAwsImages(ctx)
	}()
	wg.Wait()

	if result1.Error != nil {
		return nil, result1
	} else if result2.Error != nil {
		return nil, result2
	}

	var instanceType *types.InstanceTypeInfo
	for _, it := range instanceTypes {
		if it.InstanceType == types.InstanceType(m.Aws.InstanceType.V) {
			instanceType = &it
			break
		}
	}
	if instanceType == nil {
		return nil, base.ErrorFromMessage("instance type %s not found", m.Aws.InstanceType.V)
	}

	var filteredImages []types.Image
	for _, image := range images {
		if slices.Index(instanceType.ProcessorInfo.SupportedArchitectures, types.ArchitectureType(image.Architecture)) == -1 {
			continue
		}
		filteredImages = append(filteredImages, image)
	}

	sort.Slice(filteredImages, func(i, j int) bool {
		i1 := filteredImages[i]
		i2 := filteredImages[j]
		return *i1.Name < *i2.Name
	})

	return &filteredImages[len(filteredImages)-1], base.ReconcileResult{}
}
