package volume_providers

import (
	"context"
	"fmt"
	"net/url"
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
	if i.Body.Rustic.StorageS3 == nil {
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
		Type: string(dmodel.VolumeProviderTypeRustic),
		Name: i.Body.Name,
	}

	err = r.Create(q)
	if err != nil {
		return nil, err
	}

	if i.Body.Rustic != nil {
		r.Rustic = &dmodel.VolumeProviderRustic{
			ID:          querier.N(r.ID),
			Password:    querier.N(i.Body.Rustic.Password),
			StorageType: string(dmodel.VolumeProviderStorageTypeS3),
		}
		err = r.Rustic.Create(q)
		if err != nil {
			return nil, err
		}

		if i.Body.Rustic.StorageS3 != nil {
			err = s.checkEndpoint(i.Body.Rustic.StorageS3.Endpoint)
			if err != nil {
				return nil, err
			}
			if i.Body.Rustic.StorageS3.Prefix != "" {
				err = s.checkPrefix(i.Body.Rustic.StorageS3.Prefix)
				if err != nil {
					return nil, err
				}
			}

			r.Rustic.StorageS3 = &dmodel.VolumeProviderStorageS3{
				ID:              querier.N(r.ID),
				Endpoint:        querier.N(i.Body.Rustic.StorageS3.Endpoint),
				Region:          i.Body.Rustic.StorageS3.Region,
				Bucket:          querier.N(i.Body.Rustic.StorageS3.Bucket),
				AccessKeyId:     querier.N(i.Body.Rustic.StorageS3.AccessKeyId),
				SecretAccessKey: querier.N(i.Body.Rustic.StorageS3.SecretAccessKey),
				Prefix:          querier.N(i.Body.Rustic.StorageS3.Prefix),
			}
			err = r.Rustic.StorageS3.Create(q)
			if err != nil {
				return nil, err
			}
		}
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
			if body.Rustic.StorageS3.Endpoint != nil {
				err := s.checkEndpoint(*body.Rustic.StorageS3.Endpoint)
				if err != nil {
					return err
				}
				err = r.Rustic.StorageS3.UpdateEndpoint(q, *body.Rustic.StorageS3.Endpoint)
				if err != nil {
					return err
				}
			}
			if body.Rustic.StorageS3.Region != nil {
				err := r.Rustic.StorageS3.UpdateRegion(q, body.Rustic.StorageS3.Region)
				if err != nil {
					return err
				}
			}
			if body.Rustic.StorageS3.Bucket != nil {
				err := r.Rustic.StorageS3.UpdateBucket(q, *body.Rustic.StorageS3.Bucket)
				if err != nil {
					return err
				}
			}
			if body.Rustic.StorageS3.Prefix != nil {
				err := s.checkPrefix(*body.Rustic.StorageS3.Prefix)
				if err != nil {
					return err
				}
				err = r.Rustic.StorageS3.UpdatePrefix(q, *body.Rustic.StorageS3.Prefix)
				if err != nil {
					return err
				}
			}

			if body.Rustic.StorageS3.AccessKeyId != nil || body.Rustic.StorageS3.SecretAccessKey != nil {
				if body.Rustic.StorageS3.AccessKeyId == nil || body.Rustic.StorageS3.SecretAccessKey == nil {
					return huma.Error400BadRequest("either all or none of accessKeyId and secretAccessKey must be set")
				}
				err := r.Rustic.StorageS3.UpdateKeys(q, *body.Rustic.StorageS3.AccessKeyId, *body.Rustic.StorageS3.SecretAccessKey)
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

	err := dmodel.SoftDeleteWithConstraintsByIds[dmodel.VolumeProvider](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *VolumeProviderServer) checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return huma.Error400BadRequest("invalid endpoint", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return huma.Error400BadRequest("invalid endpoint scheme")
	}
	return nil
}

var prefixRegex = regexp.MustCompile(`^([a-zA-Z0-9]*)(/([a-zA-Z0-9]+))*/?$`)

func (s *VolumeProviderServer) checkPrefix(prefix string) error {
	if !prefixRegex.MatchString(prefix) {
		return fmt.Errorf("invalid prefix")
	}
	return nil
}
