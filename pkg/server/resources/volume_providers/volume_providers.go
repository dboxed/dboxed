package volume_providers

import (
	"context"
	"fmt"
	"regexp"

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
	s := &VolumeProviderServer{}
	return s
}

func (s *VolumeProviderServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/volume-providers", s.restCreateVolumeProvider)
	huma.Get(workspacesGroup, "/volume-providers", s.restListVolumeProviders)
	huma.Get(workspacesGroup, "/volume-providers/{id}", s.restGetVolumeProvider)
	huma.Get(workspacesGroup, "/volume-providers/by-name/{volumeProviderName}", s.restGetVolumeProviderByName)
	huma.Patch(workspacesGroup, "/volume-providers/{id}", s.restUpdateVolumeProvider)
	huma.Delete(workspacesGroup, "/volume-providers/{id}", s.restDeleteVolumeProvider)

	return nil
}

func (s *VolumeProviderServer) restCreateVolumeProvider(ctx context.Context, i *huma_utils.JsonBody[models.CreateVolumeProvider]) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid name", err)
	}

	if i.Body.Rustic == nil {
		return nil, huma.Error400BadRequest("currently only rustic is supported")
	}
	if i.Body.Rustic.StorageType != dmodel.VolumeProviderStorageTypeS3 {
		return nil, huma.Error400BadRequest("currently only S3 storage is supported")
	}

	if i.Body.Rustic != nil {
		if i.Body.Rustic.Password == "" {
			return nil, huma.Error400BadRequest("rustic password is missing")
		}
	}

	r := dmodel.VolumeProvider{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Type: i.Body.Type,
		Name: i.Body.Name,
	}

	err = r.Create(q)
	if err != nil {
		return nil, err
	}

	checkS3Bucket := func(bucketId string) error {
		_, err := dmodel.GetS3BucketById(q, &w.ID, bucketId, true)
		if err != nil {
			return err
		}
		return nil
	}

	switch i.Body.Type {
	case dmodel.VolumeProviderTypeRustic:
		if i.Body.Rustic == nil {
			return nil, fmt.Errorf("missing rustic config")
		}

		err = s.checkPrefix(i.Body.Rustic.StoragePrefix)
		if err != nil {
			return nil, err
		}

		r.Rustic = &dmodel.VolumeProviderRustic{
			ID:            querier.N(r.ID),
			StorageType:   querier.N(i.Body.Rustic.StorageType),
			Password:      querier.N(i.Body.Rustic.Password),
			StoragePrefix: querier.N(i.Body.Rustic.StoragePrefix),
		}

		switch i.Body.Rustic.StorageType {
		case dmodel.VolumeProviderStorageTypeS3:
			if i.Body.Rustic.S3BucketId == nil {
				return nil, fmt.Errorf("missing S3 bucket id")
			}

			err = checkS3Bucket(*i.Body.Rustic.S3BucketId)
			if err != nil {
				return nil, err
			}

			r.Rustic.S3BucketID = i.Body.Rustic.S3BucketId
		default:
			return nil, fmt.Errorf("unsupported storage type %s", i.Body.Rustic.StorageType)
		}

		err = r.Rustic.Create(q)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported volume provider type %s", i.Body.Type)
	}

	return huma_utils.NewJsonBody(models.VolumeProviderFromDB(r)), nil
}

func (s *VolumeProviderServer) restListVolumeProviders(ctx context.Context, i *struct{}) (*huma_utils.List[models.VolumeProvider], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	l, err := dmodel.ListVolumeProviders(q, &w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.VolumeProvider
	for _, r := range l {
		mm := models.VolumeProviderFromDB(r)
		ret = append(ret, mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *VolumeProviderServer) restGetVolumeProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetVolumeProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	m := models.VolumeProviderFromDB(*r)
	return huma_utils.NewJsonBody(m), nil
}

type VolumeProviderName struct {
	VolumeProviderName string `path:"volumeProviderName"`
}

func (s *VolumeProviderServer) restGetVolumeProviderByName(c context.Context, i *VolumeProviderName) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetVolumeProviderByName(q, w.ID, i.VolumeProviderName, true)
	if err != nil {
		return nil, err
	}

	m := models.VolumeProviderFromDB(*r)
	return huma_utils.NewJsonBody(m), nil
}

type restUpdateVolumeProviderInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateVolumeProvider]
}

func (s *VolumeProviderServer) restUpdateVolumeProvider(c context.Context, i *restUpdateVolumeProviderInput) (*huma_utils.JsonBody[models.VolumeProvider], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetVolumeProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.doUpdateVolumeProvider(c, r, i.Body)
	if err != nil {
		return nil, err
	}

	m := models.VolumeProviderFromDB(*r)

	return huma_utils.NewJsonBody(m), nil
}

func (s *VolumeProviderServer) doUpdateVolumeProvider(c context.Context, r *dmodel.VolumeProvider, body models.UpdateVolumeProvider) error {
	q := querier.GetQuerier(c)
	if body.Rustic != nil {
		if dmodel.VolumeProviderType(r.Type) != dmodel.VolumeProviderTypeRustic {
			return huma.Error400BadRequest("invalid update, not a rustic volume provider")
		}

		if body.Rustic.Password != nil {
			if *body.Rustic.Password == "" {
				return huma.Error400BadRequest("rustic password can not be empty")
			}
			err := r.Rustic.UpdatePassword(q, *body.Rustic.Password)
			if err != nil {
				return err
			}
		}

		if body.Rustic.StorageS3 != nil {
			if r.Rustic.StorageType.V != dmodel.VolumeProviderStorageTypeS3 {
				return huma.Error400BadRequest("invalid update, not a S3 based volume provider")
			}
			if body.Rustic.StorageS3.S3BucketId != nil {
				err := r.Rustic.UpdateS3Bucket(q, *body.Rustic.StorageS3.S3BucketId)
				if err != nil {
					return err
				}
			}
			if body.Rustic.StorageS3.StoragePrefix != nil {
				err := s.checkPrefix(*body.Rustic.StorageS3.StoragePrefix)
				if err != nil {
					return err
				}
				err = r.Rustic.UpdateStoragePrefix(q, *body.Rustic.StorageS3.StoragePrefix)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *VolumeProviderServer) restDeleteVolumeProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.VolumeProvider](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

var prefixRegex = regexp.MustCompile(`^([-a-zA-Z0-9]*)(/([-a-zA-Z0-9]+))*/?$`)

func (s *VolumeProviderServer) checkPrefix(prefix string) error {
	if !prefixRegex.MatchString(prefix) {
		return fmt.Errorf("invalid prefix")
	}
	return nil
}
