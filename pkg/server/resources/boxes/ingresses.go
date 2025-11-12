package boxes

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
)

func (s *BoxesServer) restListBoxIngresses(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.BoxIngress], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := s.checkBoxToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ingresses, err := dmodel.ListBoxIngresses(q, box.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.BoxIngress
	for _, ing := range ingresses {
		ret = append(ret, *models.BoxIngressFromDB(ing))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restCreateBoxIngressInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateBoxIngress]
}

func (s *BoxesServer) restCreateBoxIngress(c context.Context, i *restCreateBoxIngressInput) (*huma_utils.JsonBody[models.BoxIngress], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	proxy, err := dmodel.GetIngressProxyById(q, &w.ID, i.Body.ProxyID, true)
	if err != nil {
		return nil, err
	}

	err = s.validateBoxIngressParams(&i.Body.Hostname, &i.Body.PathPrefix, &i.Body.Port)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(c, "creating box ingress", slog.Any("boxId", box.ID), slog.Any("proxyId", proxy.ID))

	ingress := &dmodel.BoxIngress{
		ProxyID:     proxy.ID,
		BoxID:       box.ID,
		Description: i.Body.Description,
		Hostname:    i.Body.Hostname,
		PathPrefix:  i.Body.PathPrefix,
		Port:        i.Body.Port,
	}

	err = ingress.Create(q)
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.BoxIngressFromDB(*ingress)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdateBoxIngressInput struct {
	huma_utils.IdByPath
	IngressId string `path:"ingressId"`
	huma_utils.JsonBody[models.UpdateBoxIngress]
}

func (s *BoxesServer) restUpdateBoxIngress(c context.Context, i *restUpdateBoxIngressInput) (*huma_utils.JsonBody[models.BoxIngress], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	ingress, err := dmodel.GetBoxIngress(q, box.ID, i.IngressId)
	if err != nil {
		return nil, err
	}

	err = s.validateBoxIngressParams(i.Body.Hostname, i.Body.PathPrefix, i.Body.Port)
	if err != nil {
		return nil, err
	}

	err = ingress.Update(q, i.Body.Description, i.Body.Hostname, i.Body.PathPrefix, i.Body.Port)
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.BoxIngressFromDB(*ingress)
	return huma_utils.NewJsonBody(*ret), nil
}

type restDeleteBoxIngressInput struct {
	huma_utils.IdByPath
	IngressId string `path:"ingressId"`
}

func (s *BoxesServer) restDeleteBoxIngress(c context.Context, i *restDeleteBoxIngressInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = querier2.DeleteOneByFields[dmodel.BoxIngress](q, map[string]any{
		"box_id": box.ID,
		"id":     i.IngressId,
	})
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) validateBoxIngressParams(hostname *string, pathPrefix *string, port *int) error {
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
