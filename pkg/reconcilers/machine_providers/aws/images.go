package aws

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/sync/errgroup"
)

func (r *Reconciler) describeAwsInstanceTypes(ctx context.Context) ([]types.InstanceTypeInfo, error) {
	if r.awsInstanceTypes != nil {
		return r.awsInstanceTypes, nil
	}

	var nextToken *string
	for {
		resp, err := r.ec2Client.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		r.awsInstanceTypes = append(r.awsInstanceTypes, resp.InstanceTypes...)
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	return r.awsInstanceTypes, nil
}

func (r *Reconciler) describeAwsImages(ctx context.Context) ([]types.Image, error) {
	if r.awsImages != nil {
		return r.awsImages, nil
	}

	var nextToken *string
	for {
		resp, err := r.ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name:   util.Ptr("name"),
					Values: []string{"al2023-ami-2023.*"},
				},
				{
					Name:   util.Ptr("state"),
					Values: []string{"available"},
				},
			},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		r.awsImages = append(r.awsImages, resp.Images...)
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}
	return r.awsImages, nil
}

func (r *Reconciler) selectAwsImage(ctx context.Context, m dmodel.Machine) (*types.Image, error) {
	var errGrp errgroup.Group

	var instanceTypes []types.InstanceTypeInfo
	var images []types.Image

	errGrp.Go(func() error {
		var err error
		instanceTypes, err = r.describeAwsInstanceTypes(ctx)
		return err
	})
	errGrp.Go(func() error {
		var err error
		images, err = r.describeAwsImages(ctx)
		return err
	})
	err := errGrp.Wait()
	if err != nil {
		return nil, err
	}

	var instanceType *types.InstanceTypeInfo
	for _, it := range instanceTypes {
		if it.InstanceType == types.InstanceType(m.Aws.InstanceType.V) {
			instanceType = &it
			break
		}
	}
	if instanceType == nil {
		return nil, fmt.Errorf("instance type %s not found", m.Aws.InstanceType.V)
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

	return &filteredImages[len(filteredImages)-1], nil
}
