package machine_providers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/crypto/ssh"
)

type MachineProviderServer struct {
}

func New() *MachineProviderServer {
	return &MachineProviderServer{}
}

func (s *MachineProviderServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/machine-providers", s.restCreateMachineProvider)
	huma.Get(workspacesGroup, "/machine-providers", s.restListMachineProviders)
	huma.Get(workspacesGroup, "/machine-providers/{id}", s.restGetMachineProvider)
	huma.Patch(workspacesGroup, "/machine-providers/{id}", s.restUpdateMachineProvider)
	huma.Delete(workspacesGroup, "/machine-providers/{id}", s.restDeleteMachineProvider)

	huma.Get(rootGroup, "/v1/machine-provider-info/aws/regions", func(ctx context.Context, i *struct{}) (*huma_utils.List[models.AwsRegion], error) {
		return huma_utils.NewList(awsRegions, len(awsRegions)), nil
	})
	huma.Get(rootGroup, "/v1/machine-provider-info/hetzner/locations", func(ctx context.Context, i *struct{}) (*huma_utils.List[models.HetznerLocation], error) {
		return huma_utils.NewList(hetznerLocations, len(hetznerLocations)), nil
	})
	huma.Get(rootGroup, "/v1/machine-provider-info/hetzner/server-types", func(ctx context.Context, i *struct{}) (*huma_utils.List[models.HetznerServerType], error) {
		return huma_utils.NewList(hetznerServerTypes, len(hetznerServerTypes)), nil
	})

	return nil
}

func (s *MachineProviderServer) restCreateMachineProvider(c context.Context, i *huma_utils.JsonBody[models.CreateMachineProvider]) (*huma_utils.JsonBody[models.MachineProvider], error) {
	q := querier.GetQuerier(c)
	workspace := auth_middleware.GetWorkspace(c)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	log := slog.With(slog.Any("workspace", workspace.ID), slog.Any("type", i.Body.Type), slog.Any("name", i.Body.Name))
	log.InfoContext(c, "creating new machine provider")

	mp := &dmodel.MachineProvider{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspace.ID,
		},
		Type: i.Body.Type,
		Name: i.Body.Name,
	}

	if i.Body.SshKeyPublic != nil {
		_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(*i.Body.SshKeyPublic))
		if err != nil {
			return nil, err
		}
		mp.SshKeyPublic = i.Body.SshKeyPublic
	}

	err = mp.Create(q)
	if err != nil {
		return nil, err
	}

	switch i.Body.Type {
	case dmodel.MachineProviderTypeAws:
		if i.Body.Aws == nil {
			return nil, huma.Error400BadRequest("aws field not set")
		}
		err = s.restCreateMachineProviderAws(c, log, mp, i.Body.Aws)
		if err != nil {
			return nil, err
		}
	case dmodel.MachineProviderTypeHetzner:
		if i.Body.Hetzner == nil {
			return nil, huma.Error400BadRequest("hetzner field not set")
		}
		err = s.restCreateMachineProviderHetzner(c, log, mp, i.Body.Hetzner)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid type %s", i.Body.Type))
	}

	mcp, err := s.postprocessMachineProvider(c, *mp)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*mcp), nil
}

func (s *MachineProviderServer) restListMachineProviders(c context.Context, i *struct{}) (*huma_utils.List[models.MachineProvider], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListMachineProviders(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.MachineProvider
	for _, mp := range l {
		mcp, err := s.postprocessMachineProvider(c, mp)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mcp)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *MachineProviderServer) restGetMachineProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.MachineProvider], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	mp, err := dmodel.GetMachineProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	mcp, err := s.postprocessMachineProvider(c, *mp)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mcp), nil
}

type restUpdateMachineProviderInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateMachineProvider]
}

func (s *MachineProviderServer) restUpdateMachineProvider(c context.Context, i *restUpdateMachineProviderInput) (*huma_utils.JsonBody[models.MachineProvider], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	mp, err := dmodel.GetMachineProviderById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.doUpdateMachineProvider(c, mp, i.Body)
	if err != nil {
		return nil, err
	}

	m, err := s.postprocessMachineProvider(c, *mp)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, mp)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*m), nil
}

func (s *MachineProviderServer) doUpdateMachineProvider(c context.Context, mp *dmodel.MachineProvider, body models.UpdateMachineProvider) error {
	q := querier.GetQuerier(c)
	log := slog.With(slog.Any("workspace", mp.WorkspaceID), slog.Any("type", mp.Type), slog.Any("name", mp.Name))
	log.InfoContext(c, "updating machine provider")

	if body.SshKeyPublic != nil {
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*body.SshKeyPublic))
		if err != nil {
			return huma.Error400BadRequest("invalid ssh key", err)
		}
		err = mp.UpdateSshKeyPublic(q, body.SshKeyPublic)
		if err != nil {
			return err
		}
	}

	switch mp.Type {
	case dmodel.MachineProviderTypeAws:
		if body.Aws != nil {
			var err error
			err = s.restUpdateMachineProviderAws(c, log, mp, body.Aws)
			if err != nil {
				return err
			}
		}
	case dmodel.MachineProviderTypeHetzner:
		if body.Hetzner != nil {
			var err error
			err = s.restUpdateMachineProviderHetzner(c, log, mp, body.Hetzner)
			if err != nil {
				return err
			}
		}
	default:
		return huma.Error400BadRequest("one of the machine specific sub-structs must be set")
	}
	return nil
}

func (s *MachineProviderServer) restDeleteMachineProvider(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.MachineProvider](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *MachineProviderServer) postprocessMachineProvider(c context.Context, mp dmodel.MachineProvider) (*models.MachineProvider, error) {
	ret := models.MachineProviderFromDB(mp)

	if mp.SshKeyPublic != nil {
		pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*mp.SshKeyPublic))
		if err == nil {
			ret.SshKeyFingerprint = util.Ptr(ssh.FingerprintSHA256(pk))
		}
	}

	switch mp.Type {
	case dmodel.MachineProviderTypeAws:
		err := s.postprocessMachineProviderAws(c, &mp, ret)
		if err != nil {
			return nil, err
		}
	case dmodel.MachineProviderTypeHetzner:
		err := s.postprocessMachineProviderHetzner(c, &mp, ret)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest("all machine provider structs are nil")
	}
	return ret, nil
}

func (s *MachineProviderServer) postprocessMachineProviderAws(c context.Context, mp *dmodel.MachineProvider, ret *models.MachineProvider) error {
	ret.Aws = models.MachineProviderAwsFromDB(*mp.Aws)
	for _, subnet := range mp.Aws.Status.Subnets {
		ret.Aws.Subnets = append(ret.Aws.Subnets, *models.MachineProviderAwsSubnetFromDB(subnet))
	}
	return nil
}

func (s *MachineProviderServer) postprocessMachineProviderHetzner(c context.Context, mp *dmodel.MachineProvider, ret *models.MachineProvider) error {
	ret.Hetzner = models.MachineProviderHetznerFromDB(*mp.Hetzner)
	return nil
}
