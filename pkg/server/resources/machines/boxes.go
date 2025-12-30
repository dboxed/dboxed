package machines

import (
	"context"

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
		box := models.BoxFromDB(b.Box, b.Sandbox)
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

	err = box.UpdateMachineID(q, &machine.ID, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, machine)
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

	if box.MachineFromSpec {
		return nil, huma.Error400BadRequest("box was added via dboxed spec and can't be manually removed")
	}

	err = box.UpdateMachineID(q, nil, false)
	if err != nil {
		return nil, err
	}

	err = InvalidateBoxTokens(c, w.ID, machine.ID, &box.ID)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, machine)
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

	err = InvalidateBoxTokens(c, w.ID, machine.ID, &box.ID)
	if err != nil {
		return nil, err
	}

	tokenName := BuildBoxTokenNamePrefix(machine.ID, &box.ID) + util.RandomString(8)
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
