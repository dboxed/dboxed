package machines

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type MachinesServer struct {
	config config.Config
}

func New(config config.Config) *MachinesServer {
	return &MachinesServer{
		config: config,
	}
}

func (s *MachinesServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/machines", s.restCreateMachine)
	huma.Get(workspacesGroup, "/machines", s.restListMachines)
	huma.Get(workspacesGroup, "/machines/{id}", s.restGetMachine)
	huma.Patch(workspacesGroup, "/machines/{id}", s.restUpdateMachine)
	huma.Delete(workspacesGroup, "/machines/{id}", s.restDeleteMachine)

	return nil
}

func (s *MachinesServer) restCreateMachine(c context.Context, i *huma_utils.JsonBody[models.CreateMachine]) (*huma_utils.JsonBody[models.Machine], error) {
	q := querier2.GetQuerier(c)

	machine, inputErr, err := s.createMachine(c, i.Body)
	if err != nil {
		return nil, err
	}
	if inputErr != "" {
		return nil, huma.Error400BadRequest(inputErr)
	}

	ret, err := models.MachineFromDB(*machine)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, machine)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *MachinesServer) createMachine(c context.Context, body models.CreateMachine) (*dmodel.Machine, string, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err.Error(), nil
	}

	m := &dmodel.Machine{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Name: body.Name,
	}

	if body.MachineProvider != nil {
		mp, err := dmodel.GetMachineProviderById(q, &w.ID, *body.MachineProvider, true)
		if err != nil {
			return nil, "", err
		}
		m.MachineProviderID = &mp.ID
		m.MachineProvider = mp
		m.MachineProviderType = &mp.Type
	}

	err = m.Create(q)
	if err != nil {
		return nil, "", err
	}

	if body.MachineProvider != nil {
		switch global.MachineProviderType(*m.MachineProviderType) {
		case global.MachineProviderHetzner:
			if body.Hetzner == nil {
				return nil, "missing hetzner config", nil
			}
			err = s.createMachineHetzner(c, m, *body.Hetzner)
			if err != nil {
				return nil, "", err
			}
		case global.MachineProviderAws:
			if body.Aws == nil {
				return nil, "missing aws config", nil
			}
			err = s.createMachineAws(c, m, *body.Aws)
			if err != nil {
				return nil, "", err
			}
		default:
			return nil, "unknown machine provider type", nil
		}
	}

	return m, "", nil
}

func (s *MachinesServer) createMachineHetzner(c context.Context, machine *dmodel.Machine, body models.CreateMachineHetzner) error {
	q := querier2.GetQuerier(c)
	machine.Hetzner = &dmodel.MachineHetzner{
		ID:             querier2.N(machine.ID),
		ServerType:     querier2.N(body.ServerType),
		ServerLocation: querier2.N(body.ServerLocation),
		Status: &dmodel.MachineHetznerStatus{
			ID: querier2.N(machine.ID),
		},
	}
	err := machine.Hetzner.Create(q)
	if err != nil {
		return err
	}
	err = machine.Hetzner.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *MachinesServer) createMachineAws(c context.Context, machine *dmodel.Machine, body models.CreateMachineAws) error {
	q := querier2.GetQuerier(c)

	if body.SubnetId == "" {
		return huma.Error400BadRequest("subnet_id can not be empty")
	}

	if body.RootVolumeSize == nil {
		body.RootVolumeSize = util.Ptr(int64(40))
	}
	if *body.RootVolumeSize < 8 {
		return huma.Error400BadRequest("root_volume_size must be at least 8 GB")
	}

	subnet, err := dmodel.GetMachineProviderSubnet(q, *machine.MachineProviderID, body.SubnetId)
	if err != nil {
		return err
	}

	machine.Aws = &dmodel.MachineAws{
		ID:             querier2.N(machine.ID),
		InstanceType:   querier2.N(body.InstanceType),
		SubnetID:       querier2.N(subnet.SubnetID.V),
		RootVolumeSize: querier2.N(*body.RootVolumeSize),
		Status: &dmodel.MachineAwsStatus{
			ID: querier2.N(machine.ID),
		},
	}

	err = machine.Aws.Create(q)
	if err != nil {
		return err
	}
	err = machine.Aws.Status.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *MachinesServer) restListMachines(c context.Context, i *struct{}) (*huma_utils.List[models.Machine], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListMachinesForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.Machine
	for _, m := range l {
		mm, err := s.postprocessMachine(c, m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *MachinesServer) restGetMachine(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Machine], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	m, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	mm, err := s.postprocessMachine(c, *m)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mm), nil
}

type restUpdateMachineInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateMachine]
}

func (s *MachinesServer) restUpdateMachine(c context.Context, i *restUpdateMachineInput) (*huma_utils.JsonBody[models.Machine], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	m, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	// TODO nothing to do for now

	mm, err := s.postprocessMachine(c, *m)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, m)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mm), nil
}

func (s *MachinesServer) restDeleteMachine(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Machine](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *MachinesServer) postprocessMachine(c context.Context, machine dmodel.Machine) (*models.Machine, error) {
	ret, err := models.MachineFromDB(machine)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
