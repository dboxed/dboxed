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

func (s *BoxesServer) restListLoadBalancerServices(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.LoadBalancerService], error) {
	q := querier2.GetQuerier(c)

	err := auth_middleware.CheckResourceAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	lbServices, err := dmodel.ListLoadBalancerServices(q, i.Id)
	if err != nil {
		return nil, err
	}

	var ret []models.LoadBalancerService
	for _, ing := range lbServices {
		ret = append(ret, *models.LoadBalancerServiceFromDB(ing))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restCreateLoadBalancerServiceInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateLoadBalancerService]
}

func (s *BoxesServer) restCreateLoadBalancerService(c context.Context, i *restCreateLoadBalancerServiceInput) (*huma_utils.JsonBody[models.LoadBalancerService], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Body.LoadBalancerID, true)
	if err != nil {
		return nil, err
	}

	if box.NetworkID == nil || *box.NetworkID != lb.NetworkId {
		return nil, huma.Error400BadRequest("box is not in the same network as the load balancer", nil)
	}

	err = s.validateLoadBalancerServiceParams(&i.Body.Hostname, &i.Body.PathPrefix, &i.Body.Port)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(c, "creating load balancer service", slog.Any("boxId", box.ID), slog.Any("loadBalancerId", lb.ID))

	sbService := &dmodel.LoadBalancerService{
		LoadBalancerId: lb.ID,
		BoxID:          box.ID,
		Description:    i.Body.Description,
		Hostname:       i.Body.Hostname,
		PathPrefix:     i.Body.PathPrefix,
		Port:           i.Body.Port,
	}

	err = sbService.Create(q)
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

	err = dmodel.BumpChangeSeq(q, lb)
	if err != nil {
		return nil, err
	}

	ret := models.LoadBalancerServiceFromDB(*sbService)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdateLoadBalancerServiceInput struct {
	huma_utils.IdByPath
	LoadBalancerServiceId string `path:"loadBalancerServiceId"`
	huma_utils.JsonBody[models.UpdateLoadBalancerService]
}

func (s *BoxesServer) restUpdateLoadBalancerService(c context.Context, i *restUpdateLoadBalancerServiceInput) (*huma_utils.JsonBody[models.LoadBalancerService], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	lbService, err := dmodel.GetLoadBalancerService(q, box.ID, i.LoadBalancerServiceId)
	if err != nil {
		return nil, err
	}

	err = s.validateLoadBalancerServiceParams(i.Body.Hostname, i.Body.PathPrefix, i.Body.Port)
	if err != nil {
		return nil, err
	}

	err = lbService.Update(q, i.Body.Description, i.Body.Hostname, i.Body.PathPrefix, i.Body.Port)
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

	err = dmodel.BumpChangeSeqForId[*dmodel.LoadBalancer](q, lbService.LoadBalancerId)
	if err != nil {
		return nil, err
	}

	ret := models.LoadBalancerServiceFromDB(*lbService)
	return huma_utils.NewJsonBody(*ret), nil
}

type restDeleteLoadBalancerServiceInput struct {
	huma_utils.IdByPath
	LoadBalancerServiceId string `path:"loadBalancerServiceId"`
}

func (s *BoxesServer) restDeleteLoadBalancerService(c context.Context, i *restDeleteLoadBalancerServiceInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = querier2.DeleteOneByFields[dmodel.LoadBalancerService](q, map[string]any{
		"box_id": box.ID,
		"id":     i.LoadBalancerServiceId,
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

	err = dmodel.BumpChangeSeqForId[*dmodel.LoadBalancer](q, i.LoadBalancerServiceId)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) validateLoadBalancerServiceParams(hostname *string, pathPrefix *string, port *int) error {
	if hostname != nil {
		if len(*hostname) == 0 {
			return huma.Error400BadRequest("hostname cannot be empty", nil)
		}
	}
	if pathPrefix != nil {
		if len(*pathPrefix) > 0 && (*pathPrefix)[0] != '/' {
			return huma.Error400BadRequest("path_prefix must start with /", nil)
		}
	}
	if port != nil {
		if *port < 1 || *port > 65535 {
			return huma.Error400BadRequest("port must be between 1 and 65535", nil)
		}
	}
	return nil
}
