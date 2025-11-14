package ingress_proxies

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type IngressProxyServer struct {
	config config.Config
}

func New(config config.Config) *IngressProxyServer {
	return &IngressProxyServer{
		config: config,
	}
}

func (s *IngressProxyServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/ingress-proxies", s.restCreateIngressProxy)
	huma.Get(workspacesGroup, "/ingress-proxies", s.restListIngressProxies)
	huma.Get(workspacesGroup, "/ingress-proxies/{id}", s.restGetIngressProxy)
	huma.Patch(workspacesGroup, "/ingress-proxies/{id}", s.restUpdateIngressProxy)
	huma.Delete(workspacesGroup, "/ingress-proxies/{id}", s.restDeleteIngressProxy)

	return nil
}

func (s *IngressProxyServer) restCreateIngressProxy(c context.Context, i *huma_utils.JsonBody[models.CreateIngressProxy]) (*huma_utils.JsonBody[models.IngressProxy], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), nil)
	}

	if i.Body.ProxyType != "caddy" {
		return nil, huma.Error400BadRequest("invalid proxy_type, must be 'caddy'", nil)
	}

	slog.InfoContext(c, "creating ingress proxy", slog.Any("name", i.Body.Name))

	// Validate port ranges
	if i.Body.HttpPort < 1 || i.Body.HttpPort > 65535 {
		return nil, huma.Error400BadRequest("http_port must be between 1 and 65535", nil)
	}
	if i.Body.HttpsPort < 1 || i.Body.HttpsPort > 65535 {
		return nil, huma.Error400BadRequest("https_port must be between 1 and 65535", nil)
	}
	if i.Body.HttpPort == i.Body.HttpsPort {
		return nil, huma.Error400BadRequest("http_port and https_port can't be the same", nil)
	}

	if i.Body.Replicas < 0 || i.Body.Replicas > 10 {
		return nil, huma.Error400BadRequest("replicas must be between 0 and 10", nil)
	}

	network, err := dmodel.GetNetworkById(q, &w.ID, i.Body.Network, true)
	if err != nil {
		return nil, err
	}

	proxy := &dmodel.IngressProxy{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Name:      i.Body.Name,
		ProxyType: string(i.Body.ProxyType),
		NetworkId: network.ID,
		HttpPort:  i.Body.HttpPort,
		HttpsPort: i.Body.HttpsPort,
		Replicas:  i.Body.Replicas,
	}

	err = proxy.Create(q)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, proxy)
	if err != nil {
		return nil, err
	}

	ret := models.IngressProxyFromDB(*proxy)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *IngressProxyServer) restListIngressProxies(c context.Context, i *struct{}) (*huma_utils.List[models.IngressProxy], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	proxies, err := dmodel.ListIngressProxiesForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.IngressProxy
	for _, p := range proxies {
		ret = append(ret, *models.IngressProxyFromDB(p))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *IngressProxyServer) restGetIngressProxy(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.IngressProxy], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	proxy, err := dmodel.GetIngressProxyById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret := models.IngressProxyFromDB(*proxy)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdateIngressProxyInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateIngressProxy]
}

func (s *IngressProxyServer) restUpdateIngressProxy(c context.Context, i *restUpdateIngressProxyInput) (*huma_utils.JsonBody[models.IngressProxy], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	proxy, err := dmodel.GetIngressProxyById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	// Validate port ranges
	if i.Body.HttpPort != nil {
		if *i.Body.HttpPort < 1 || *i.Body.HttpPort > 65535 {
			return nil, huma.Error400BadRequest("http_port must be between 1 and 65535", nil)
		}
	}
	if i.Body.HttpsPort != nil {
		if *i.Body.HttpsPort < 1 || *i.Body.HttpsPort > 65535 {
			return nil, huma.Error400BadRequest("https_port must be between 1 and 65535", nil)
		}
	}

	// Validate replicas
	if i.Body.Replicas != nil {
		if *i.Body.Replicas < 0 || *i.Body.Replicas > 10 {
			return nil, huma.Error400BadRequest("replicas must be between 0 and 10", nil)
		}
	}

	// Check that ports are not the same
	httpPort := proxy.HttpPort
	httpsPort := proxy.HttpsPort
	if i.Body.HttpPort != nil {
		httpPort = *i.Body.HttpPort
	}
	if i.Body.HttpsPort != nil {
		httpsPort = *i.Body.HttpsPort
	}
	if httpPort == httpsPort {
		return nil, huma.Error400BadRequest("http_port and https_port can't be the same", nil)
	}

	err = proxy.Update(q, i.Body.HttpPort, i.Body.HttpsPort, i.Body.Replicas)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, proxy)
	if err != nil {
		return nil, err
	}

	ret := models.IngressProxyFromDB(*proxy)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *IngressProxyServer) restDeleteIngressProxy(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	proxy, err := dmodel.GetIngressProxyById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ingresses, err := dmodel.ListBoxIngressesForProxy(q, proxy.ID)
	if err != nil {
		return nil, err
	}
	if len(ingresses) != 0 {
		return nil, huma.Error400BadRequest("can't delete ingress proxies with active ingresses")
	}

	err = dmodel.SoftDeleteWithConstraintsByIds[*dmodel.IngressProxy](q, &proxy.WorkspaceID, proxy.ID)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, proxy)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
