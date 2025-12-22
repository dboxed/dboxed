package cloud_utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/dboxed/dboxed/pkg/server/config"
)

type AwsCreds struct {
	AwsAccessKeyID     *string
	AwsSecretAccessKey *string
	Region             string
}

func BuildAwsConfig(creds AwsCreds) *aws.Config {
	creds2 := aws.Credentials{}
	if creds.AwsAccessKeyID != nil {
		creds2.AccessKeyID = *creds.AwsAccessKeyID
	}
	if creds.AwsSecretAccessKey != nil {
		creds2.SecretAccessKey = *creds.AwsSecretAccessKey
	}

	awsConfig := aws.NewConfig()
	awsConfig.Region = creds.Region
	awsConfig.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return creds2, nil
	})

	return awsConfig
}

func AwsAllHelper[P any, R any](params *P, fn func(params *P) ([]R, *string, error)) ([]R, error) {
	paramsV := reflect.Indirect(reflect.ValueOf(params))

	var ret []R
	for {
		l, nextToken, err := fn(params)
		if err != nil {
			return nil, err
		}
		ret = append(ret, l...)
		if len(l) == 0 || nextToken == nil {
			break
		}
		paramsV.FieldByName("NextToken").Set(reflect.ValueOf(nextToken))
	}
	return ret, nil
}

func AwsGetNameFromTags(tags []types.Tag) *string {
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
			return tag.Value
		}
	}
	return nil
}

func BuildAwsSshKeyName(ctx context.Context, machineProviderName string, machineProviderId string) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-%s-%s", config.InstanceName, machineProviderName, machineProviderId)
}
