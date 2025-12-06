package machines

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/tokens"
	"github.com/dboxed/dboxed/pkg/util"
)

func (s *MachinesServer) restListBoxes(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	_, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	boxes, err := dmodel.ListBoxesForMachine(q, i.Id, true)
	if err != nil {
		return nil, err
	}

	var ret []models.Box
	for _, b := range boxes {
		box, err := models.BoxFromDB(b.Box, b.SandboxStatus)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *box)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restAddBoxInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.AddBoxToMachineRequest]
}

func (s *MachinesServer) restAddBox(c context.Context, i *restAddBoxInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	machine, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Body.BoxId, true)
	if err != nil {
		return nil, err
	}

	if box.MachineID != nil {
		if *box.MachineID == machine.ID {
			// nothing to do
			return &huma_utils.Empty{}, nil
		} else {
			return nil, huma.Error400BadRequest("box is already assigned to another machine")
		}
	}

	err = box.UpdateMachineID(q, &machine.ID)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, machine)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restRemoveBoxInput struct {
	Id    string `path:"id"`
	BoxId string `path:"boxId"`
}

func (s *MachinesServer) restRemoveBox(c context.Context, i *restRemoveBoxInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	machine, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.BoxId, true)
	if err != nil {
		return nil, err
	}

	if box.MachineID == nil || *box.MachineID != machine.ID {
		return nil, huma.Error400BadRequest("box is not assigned to this machine")
	}

	err = box.UpdateMachineID(q, nil)
	if err != nil {
		return nil, err
	}

	err = s.invalidateBoxTokens(c, machine.ID, box.ID)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, machine)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restCreateBoxTokenInput struct {
	Id    string `path:"id"`
	BoxId string `path:"boxId"`
}

func (s *MachinesServer) restCreateBoxToken(c context.Context, i *restCreateBoxTokenInput) (*huma_utils.JsonBody[models.Token], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	machine, err := dmodel.GetMachineById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.BoxId, true)
	if err != nil {
		return nil, err
	}

	if box.MachineID == nil || *box.MachineID != machine.ID {
		return nil, huma.Error400BadRequest("box is not assigned to this machine")
	}

	err = s.invalidateBoxTokens(c, machine.ID, box.ID)
	if err != nil {
		return nil, err
	}

	tokenName := s.buildBoxTokenNamePrefix(machine.ID, box.ID) + util.RandomString(8)
	token, err := tokens.CreateToken(c, w.ID, models.CreateToken{
		Name:  tokenName,
		Type:  dmodel.TokenTypeBox,
		BoxID: &box.ID,
	}, true, true)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*token), nil
}

func (s *MachinesServer) listBoxTokens(ctx context.Context, machineId string, boxId string) ([]dmodel.Token, error) {
	q := querier2.GetQuerier(ctx)
	tokens, err := dmodel.ListTokensForBox(q, boxId)
	if err != nil {
		return nil, err
	}
	var ret []dmodel.Token
	prefix := s.buildBoxTokenNamePrefix(machineId, boxId)
	for _, t := range tokens {
		if strings.HasPrefix(t.Name, prefix) {
			ret = append(ret, t)
		}
	}
	return ret, nil
}

func (s *MachinesServer) invalidateBoxTokens(ctx context.Context, machineId string, boxId string) error {
	q := querier2.GetQuerier(ctx)
	oldTokens, err := s.listBoxTokens(ctx, machineId, boxId)
	if err != nil {
		return err
	}
	for _, t := range oldTokens {
		if t.ValidUntil == nil {
			err = t.UpdateValidUntil(q, util.Ptr(time.Now().Add(5*time.Minute)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MachinesServer) buildBoxTokenNamePrefix(machineId string, boxId string) string {
	return tokens.InternalTokenNamePrefix + fmt.Sprintf("machine_box_%s_%s_", machineId, boxId)
}
