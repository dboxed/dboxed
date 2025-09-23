package volumes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
)

type VolumesServer struct {
	config config.Config
}

func New(config config.Config) *VolumesServer {
	return &VolumesServer{
		config: config,
	}
}

func (s *VolumesServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/volumes", s.restCreateVolume)
	huma.Get(workspacesGroup, "/volumes", s.restListVolumes)
	huma.Get(workspacesGroup, "/volumes/{id}", s.restGetVolume)
	huma.Patch(workspacesGroup, "/volumes/{id}", s.restUpdateVolume)
	huma.Delete(workspacesGroup, "/volumes/{id}", s.restDeleteVolume)

	return nil
}

func (s *VolumesServer) restCreateVolume(c context.Context, i *huma_utils.JsonBody[models.CreateVolume]) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier2.GetQuerier(c)

	volume, inputErr, err := s.createVolume(c, i.Body)
	if err != nil {
		return nil, err
	}
	if inputErr != "" {
		return nil, huma.Error400BadRequest(inputErr)
	}

	ret := models.VolumeFromDB(*volume, nil)

	err = dmodel.AddChangeTracking(q, volume)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(ret), nil
}

func (s *VolumesServer) createVolume(c context.Context, body models.CreateVolume) (*dmodel.Volume, string, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err.Error(), nil
	}

	vp, err := dmodel.GetVolumeProviderById(q, &w.ID, body.VolumeProvider, true)
	if err != nil {
		return nil, "", err
	}

	m := &dmodel.Volume{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Uuid: uuid.NewString(),
		Name: body.Name,

		VolumeProviderID:   vp.ID,
		VolumeProviderType: vp.Type,
	}

	err = m.Create(q)
	if err != nil {
		return nil, "", err
	}

	switch global.VolumeProviderType(vp.Type) {
	case global.VolumeProviderDboxed:
		if body.Dboxed == nil {
			return nil, "missing dboxed config", nil
		}
		err = s.createVolumeDboxed(c, m, *body.Dboxed)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "unknown volume provider type", nil
	}

	return m, "", nil
}

func (s *VolumesServer) createVolumeDboxed(c context.Context, volume *dmodel.Volume, body models.CreateVolumeDboxed) error {
	q := querier2.GetQuerier(c)
	volume.Dboxed = &dmodel.VolumeDboxed{
		ID:     querier2.N(volume.ID),
		FsSize: querier2.N(body.FsSize),
		FsType: querier2.N(body.FsType),
		Status: &dmodel.VolumeDboxedStatus{
			ID: querier2.N(volume.ID),
		},
	}
	err := volume.Dboxed.Create(q)
	if err != nil {
		return err
	}
	err = volume.Dboxed.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *VolumesServer) restListVolumes(c context.Context, i *struct{}) (*huma_utils.List[models.Volume], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	l, err := dmodel.ListVolumesForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.Volume
	for _, m := range l {
		mm, err := s.postprocessVolume(c, m.Volume, m.Attachment)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *VolumesServer) restGetVolume(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	m, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	mm, err := s.postprocessVolume(c, m.Volume, m.Attachment)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mm), nil
}

type restUpdateVolumeInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateVolume]
}

func (s *VolumesServer) restUpdateVolume(c context.Context, i *restUpdateVolumeInput) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	m, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	// TODO nothing to do for now

	mm, err := s.postprocessVolume(c, m.Volume, m.Attachment)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, m)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mm), nil
}

func (s *VolumesServer) restDeleteVolume(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[dmodel.Volume](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *VolumesServer) postprocessVolume(c context.Context, volume dmodel.Volume, attachment *dmodel.BoxVolumeAttachment) (*models.Volume, error) {
	ret := models.VolumeFromDB(volume, attachment)
	return &ret, nil
}
