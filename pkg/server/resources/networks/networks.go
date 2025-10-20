package networks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type NetworksServer struct {
}

func New() *NetworksServer {
	return &NetworksServer{}
}

func (s *NetworksServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/networks", s.restCreateNetwork)
	huma.Get(workspacesGroup, "/networks", s.restListNetworks)
	huma.Get(workspacesGroup, "/networks/{id}", s.restGetNetwork)
	huma.Get(workspacesGroup, "/networks/by-name/{name}", s.restGetNetworkByName)
	huma.Patch(workspacesGroup, "/networks/{id}", s.restUpdateNetwork)
	huma.Delete(workspacesGroup, "/networks/{id}", s.restDeleteNetwork)

	return nil
}

func (s *NetworksServer) restCreateNetwork(c context.Context, i *huma_utils.JsonBody[models.CreateNetwork]) (*huma_utils.JsonBody[models.Network], error) {
	q := querier.GetQuerier(c)
	workspace := global.GetWorkspace(c)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	log := slog.With(slog.Any("workspace", workspace.ID), slog.Any("type", i.Body.Type), slog.Any("name", i.Body.Name))
	log.InfoContext(c, "creating new network")

	n := &dmodel.Network{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspace.ID,
		},
		Type: string(i.Body.Type),
		Name: i.Body.Name,
	}

	err = n.Create(q)
	if err != nil {
		return nil, err
	}

	switch i.Body.Type {
	case global.NetworkNetbird:
		if i.Body.Netbird == nil {
			return nil, huma.Error400BadRequest("netbird field not set")
		}
		err = s.restCreateNetworkNetbird(c, log, n, i.Body.Netbird)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid type %s", i.Body.Type))
	}

	mcp, err := s.postprocessNetwork(c, *n)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, n)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*mcp), nil
}

func (s *NetworksServer) restListNetworks(c context.Context, i *struct{}) (*huma_utils.List[models.Network], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	l, err := dmodel.ListNetworksForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.Network
	for _, n := range l {
		mcp, err := s.postprocessNetwork(c, n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mcp)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *NetworksServer) restGetNetwork(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Network], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	n, err := dmodel.GetNetworkById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	mcp, err := s.postprocessNetwork(c, *n)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mcp), nil
}

type NetworkName struct {
	NetworkName string `path:"name"`
}

func (s *NetworksServer) restGetNetworkByName(c context.Context, i *NetworkName) (*huma_utils.JsonBody[models.Network], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	n, err := dmodel.GetNetworkByName(q, w.ID, i.NetworkName, true)
	if err != nil {
		return nil, err
	}

	mcp, err := s.postprocessNetwork(c, *n)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*mcp), nil
}

type restUpdateNetworkInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateNetwork]
}

func (s *NetworksServer) restUpdateNetwork(c context.Context, i *restUpdateNetworkInput) (*huma_utils.JsonBody[models.Network], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	n, err := dmodel.GetNetworkById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.doUpdateNetwork(c, n, i.Body)
	if err != nil {
		return nil, err
	}

	m, err := s.postprocessNetwork(c, *n)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, n)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*m), nil
}

func (s *NetworksServer) doUpdateNetwork(c context.Context, n *dmodel.Network, body models.UpdateNetwork) error {
	log := slog.With(slog.Any("workspace", n.WorkspaceID), slog.Any("type", n.Type), slog.Any("name", n.Name))
	log.InfoContext(c, "updating network")

	switch global.NetworkType(n.Type) {
	case global.NetworkNetbird:
		if body.Netbird != nil {
			err := s.restUpdateNetworkNetbird(c, log, n, body.Netbird)
			if err != nil {
				return err
			}
		}
	default:
		return huma.Error400BadRequest("one of the network specific sub-structs must be set")
	}

	return nil
}

func (s *NetworksServer) restDeleteNetwork(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Network](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *NetworksServer) postprocessNetwork(c context.Context, n dmodel.Network) (*models.Network, error) {
	ret := models.NetworkFromDB(n)

	switch global.NetworkType(n.Type) {
	case global.NetworkNetbird:
		err := s.postprocessNetworkNetbird(c, &n, ret)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest("all network structs are nil")
	}
	return ret, nil
}

func (s *NetworksServer) postprocessNetworkNetbird(c context.Context, n *dmodel.Network, ret *models.Network) error {
	if n.Netbird != nil {
		ret.Netbird = models.NetworkNetbirdFromDB(*n.Netbird)
	}
	return nil
}
