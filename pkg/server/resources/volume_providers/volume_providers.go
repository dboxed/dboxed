package volume_providers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type VolumeProviderServer struct {
}

func New() *VolumeProviderServer {
	return &VolumeProviderServer{}
}

func (s *VolumeProviderServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/volume-providers", s.restCreateVolumeProvider)
	huma.Get(workspacesGroup, "/volume-providers", s.restListVolumeProviders)
	huma.Get(workspacesGroup, "/volume-providers/{id}", s.restGetVolumeProvider)
	huma.Patch(workspacesGroup, "/volume-providers/{id}", s.restUpdateVolumeProvider)
	huma.Delete(workspacesGroup, "/volume-providers/{id}", s.restDeleteVolumeProvider)

	return nil
}

func (s *VolumeProviderServer) restCreateVolumeProvider(c context.Context, i *huma_utils.JsonBody[models.CreateVolumeProvider]) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	workspace := global.GetWorkspace(c)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	log := slog.With(slog.Any("workspace", workspace.ID), slog.Any("type", i.Body.Type), slog.Any("name", i.Body.Name))
	log.InfoContext(c, "creating new volume provider")

	vp := &dmodel.VolumeProvider{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspace.ID,
		},
		Type: string(i.Body.Type),
		Name: i.Body.Name,
	}

	err = vp.Create(q)
	if err != nil {
		return nil, err
	}

	switch i.Body.Type {
	case global.VolumeProviderDboxed:
		if i.Body.Dboxed == nil {
			return nil, huma.Error400BadRequest("dboxed field not set")
		}
		err = s.restCreateVolumeProviderDboxed(c, log, vp, i.Body.Dboxed)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid type %s", i.Body.Type))
	}

	mcp, err := s.postprocessVolumeProvider(c, *vp)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, vp)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*mcp), nil
}

func (s *VolumeProviderServer) restListVolumeProviders(c context.Context, i *struct{}) (*huma_utils.List[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	l, err := dmodel.ListVolumeProviders(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.VolumeProvider
	for _, mp := range l {
		mcp, err := s.postprocessVolumeProvider(c, mp)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mcp)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *VolumeProviderServer) restGetVolumeProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	vp, err := dmodel.GetVolumeProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	mcp, err := s.postprocessVolumeProvider(c, *vp)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mcp), nil
}

type restUpdateVolumeProviderInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateVolumeProvider]
}

func (s *VolumeProviderServer) restUpdateVolumeProvider(c context.Context, i *restUpdateVolumeProviderInput) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	mp, err := dmodel.GetVolumeProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.doUpdateVolumeProvider(c, mp, i.Body)
	if err != nil {
		return nil, err
	}

	m, err := s.postprocessVolumeProvider(c, *mp)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, mp)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*m), nil
}

func (s *VolumeProviderServer) doUpdateVolumeProvider(c context.Context, mp *dmodel.VolumeProvider, body models.UpdateVolumeProvider) error {
	log := slog.With(slog.Any("workspace", mp.WorkspaceID), slog.Any("type", mp.Type), slog.Any("name", mp.Name))
	log.InfoContext(c, "updating volume provider")

	switch global.VolumeProviderType(mp.Type) {
	case global.VolumeProviderDboxed:
		if body.Dboxed != nil {
			var err error
			err = s.restUpdateVolumeProviderDboxed(c, log, mp, body.Dboxed)
			if err != nil {
				return err
			}
		}
	default:
		return huma.Error400BadRequest("one of the volume specific sub-structs must be set")
	}
	return nil
}

func (s *VolumeProviderServer) restDeleteVolumeProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[dmodel.VolumeProvider](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *VolumeProviderServer) postprocessVolumeProvider(c context.Context, mp dmodel.VolumeProvider) (*models.VolumeProvider, error) {
	ret := models.VolumeProviderFromDB(mp)

	switch global.VolumeProviderType(mp.Type) {
	case global.VolumeProviderDboxed:
		err := s.postprocessVolumeProviderDboxed(c, &mp, ret)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest("all volume provider structs are nil")
	}
	return ret, nil
}

func (s *VolumeProviderServer) postprocessVolumeProviderDboxed(c context.Context, mp *dmodel.VolumeProvider, ret *models.VolumeProvider) error {
	ret.Dboxed = models.VolumeProviderDboxedFromDB(*mp.Dboxed)
	return nil
}
