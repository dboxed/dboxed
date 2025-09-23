package volume_providers

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (s *VolumeProviderServer) restCreateVolumeProviderDboxed(c context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, body *models.CreateVolumeProviderDboxed) error {
	err := s.checkDboxedRepository(&body.ApiUrl, &body.Token, &body.RepositoryId)
	if err != nil {
		return err
	}

	q := querier2.GetQuerier(c)

	vp.Dboxed = &dmodel.VolumeProviderDboxed{
		ID:           querier2.N(vp.ID),
		ApiUrl:       querier2.N(body.ApiUrl),
		Token:        querier2.N(body.Token),
		RepositoryId: querier2.N(body.RepositoryId),
		Status: &dmodel.VolumeProviderDboxedStatus{
			ID: querier2.N(vp.ID),
		},
	}

	err = vp.Dboxed.Create(q)
	if err != nil {
		return err
	}

	err = vp.Dboxed.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *VolumeProviderServer) restUpdateVolumeProviderDboxed(c context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, body *models.UpdateVolumeProviderDboxed) error {
	q := querier2.GetQuerier(c)

	err := s.checkDboxedRepository(body.ApiUrl, body.Token, body.RepositoryId)
	if err != nil {
		return err
	}

	if body.ApiUrl != nil || body.Token != nil || body.RepositoryId != nil {
		err = vp.Dboxed.Update(q, body.ApiUrl, body.Token, body.RepositoryId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VolumeProviderServer) checkDboxedRepository(apiUrl *string, token *string, repositoryId *int64) error {
	if apiUrl != nil {
		_, err := url.Parse(*apiUrl)
		if err != nil {
			return huma.Error400BadRequest("api url is invalid", err)
		}
	}
	if token != nil {
		if *token == "" {
			return huma.Error400BadRequest("invalid/missing token")
		}
	}
	if repositoryId != nil {
		if *repositoryId <= 0 {
			return huma.Error400BadRequest("invalid/missing repository id")
		}
	}

	return nil
}
