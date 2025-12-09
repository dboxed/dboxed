package boxes

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
)

func (s *BoxesServer) restListPortForwards(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.BoxPortForward], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	portForwards, err := dmodel.ListBoxPortForwards(q, box.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.BoxPortForward
	for _, pf := range portForwards {
		ret = append(ret, *models.BoxPortForwardFromDB(pf))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restCreatePortForwardInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateBoxPortForward]
}

func (s *BoxesServer) restCreatePortForward(c context.Context, i *restCreatePortForwardInput) (*huma_utils.JsonBody[models.BoxPortForward], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	// Validate port forward params
	err = s.validatePortForwardParams(&i.Body.Protocol, &i.Body.HostPortFirst, &i.Body.HostPortLast, &i.Body.SandboxPort)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(c, "creating port forward for box", slog.Any("boxId", box.ID))

	portForward := &dmodel.BoxPortForward{
		BoxID:         box.ID,
		Description:   i.Body.Description,
		Protocol:      i.Body.Protocol,
		HostPortFirst: i.Body.HostPortFirst,
		HostPortLast:  i.Body.HostPortLast,
		SandboxPort:   i.Body.SandboxPort,
	}

	err = portForward.Create(q)
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.BoxPortForwardFromDB(*portForward)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdatePortForwardInput struct {
	huma_utils.IdByPath
	PortForwardId string `path:"portForwardId"`
	huma_utils.JsonBody[models.UpdateBoxPortForward]
}

func (s *BoxesServer) restUpdatePortForward(c context.Context, i *restUpdatePortForwardInput) (*huma_utils.JsonBody[models.BoxPortForward], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	portForward, err := dmodel.GetBoxPortForward(q, box.ID, i.PortForwardId)
	if err != nil {
		return nil, err
	}

	// Validate port forward params
	err = s.validatePortForwardParams(i.Body.Protocol, i.Body.HostPortFirst, i.Body.HostPortLast, i.Body.SandboxPort)
	if err != nil {
		return nil, err
	}

	err = portForward.Update(q, i.Body.Description, i.Body.Protocol, i.Body.HostPortFirst, i.Body.HostPortLast, i.Body.SandboxPort)
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.BoxPortForwardFromDB(*portForward)
	return huma_utils.NewJsonBody(*ret), nil
}

type restDeletePortForwardInput struct {
	huma_utils.IdByPath
	PortForwardId string `path:"portForwardId"`
}

func (s *BoxesServer) restDeletePortForward(c context.Context, i *restDeletePortForwardInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = querier2.DeleteOneByFields[dmodel.BoxPortForward](q, map[string]any{
		"box_id": box.ID,
		"id":     i.PortForwardId,
	})
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) validatePortForwardParams(protocol *string, hostPortFirst *int, hostPortLast *int, sandboxPort *int) error {
	if protocol != nil {
		if *protocol != "tcp" && *protocol != "udp" {
			return huma.Error400BadRequest("invalid protocol, must be 'tcp' or 'udp'", nil)
		}
	}
	if hostPortFirst != nil {
		if hostPortLast == nil {
			return huma.Error400BadRequest("host_port_first and host_port_last must always be specified together", nil)
		}
		if *hostPortFirst < 1 || *hostPortFirst > 65535 {
			return huma.Error400BadRequest("invalid host_port_first, must be between 1 and 65535", nil)
		}
	}
	if hostPortLast != nil {
		if hostPortFirst == nil {
			return huma.Error400BadRequest("host_port_first and host_port_last must always be specified together", nil)
		}
		if *hostPortLast < 1 || *hostPortLast > 65535 {
			return huma.Error400BadRequest("invalid host_port_last, must be between 1 and 65535", nil)
		}
	}
	if sandboxPort != nil {
		if *sandboxPort < 1 || *sandboxPort > 65535 {
			return huma.Error400BadRequest("invalid sandbox_port, must be between 1 and 65535", nil)
		}
	}
	if hostPortFirst != nil && hostPortLast != nil {
		if *hostPortFirst > *hostPortLast {
			return huma.Error400BadRequest("host_port_first must be less than or equal to host_port_last", nil)
		}
	}
	return nil
}
