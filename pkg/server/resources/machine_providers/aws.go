package machine_providers

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (s *MachineProviderServer) restCreateMachineProviderAws(c context.Context, log *slog.Logger, mp *dmodel.MachineProvider, body *models.CreateMachineProviderAws) error {
	if body.Region == "" {
		return huma.Error400BadRequest("region can not be empty")
	}

	err := s.checkAwsVpcId(body.VpcId)
	if err != nil {
		return err
	}
	if body.AwsAccessKeyId == "" || body.AwsSecretAccessKey == "" {
		return huma.Error400BadRequest("both aws_access_key_id and aws_secret_access_key must be provided")
	}

	q := querier2.GetQuerier(c)
	log = log.With(slog.Any("vpcId", body.VpcId))

	mp.Aws = &dmodel.MachineProviderAws{
		ID:                 querier2.N(mp.ID),
		Region:             querier2.N(body.Region),
		VpcID:              &body.VpcId,
		AwsAccessKeyID:     &body.AwsAccessKeyId,
		AwsSecretAccessKey: &body.AwsSecretAccessKey,
		Status: &dmodel.MachineProviderAwsStatus{
			ID: querier2.N(mp.ID),
		},
	}

	err = mp.Aws.Create(q)
	if err != nil {
		return err
	}

	err = mp.Aws.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *MachineProviderServer) restUpdateMachineProviderAws(c context.Context, log *slog.Logger, mp *dmodel.MachineProvider, body *models.UpdateMachineProviderAws) error {
	q := querier2.GetQuerier(c)
	if body.AwsAccessKeyId != nil || body.AwsSecretAccessKey != nil {
		if body.AwsAccessKeyId == nil || body.AwsSecretAccessKey == nil {
			return huma.Error400BadRequest("either both aws_access_key_id and aws_secret_access_key mut be provided or none of both")
		}

		log.InfoContext(c, "updating access key")
		err := mp.Aws.UpdateAccessKeys(q, body.AwsAccessKeyId, body.AwsSecretAccessKey)
		if err != nil {
			return err
		}
	}
	return nil
}

var vpcRegex = regexp.MustCompile(`^vpc-[a-z0-9]+$`)

func (s *MachineProviderServer) checkAwsVpcId(vpcId string) error {
	if vpcId == "" {
		return huma.Error400BadRequest("empty vpc_id is not allowed")
	}
	if !vpcRegex.MatchString(vpcId) {
		return huma.Error400BadRequest("invalid vpc id")
	}
	return nil
}
