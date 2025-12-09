package load_balancers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
)

type LoadBalancerServer struct {
	config config.Config
}

func New(config config.Config) *LoadBalancerServer {
	return &LoadBalancerServer{
		config: config,
	}
}

func (s *LoadBalancerServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	allowLoadBalancerTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowLoadBalancerToken, true)

	huma.Post(workspacesGroup, "/load-balancers", s.restCreateLoadBalancer)
	huma.Get(workspacesGroup, "/load-balancers", s.restListLoadBalancers)
	huma.Get(workspacesGroup, "/load-balancers/{id}", s.restGetLoadBalancer)
	huma.Patch(workspacesGroup, "/load-balancers/{id}", s.restUpdateLoadBalancer)
	huma.Delete(workspacesGroup, "/load-balancers/{id}", s.restDeleteLoadBalancer)

	huma.Put(workspacesGroup, "/load-balancers/{id}/certmagic/locks/*key", s.restPutCertmagicLock, allowLoadBalancerTokenModifier)
	huma.Delete(workspacesGroup, "/load-balancers/{id}/certmagic/locks/*key", s.restDeleteCertmagicLock, allowLoadBalancerTokenModifier)

	huma.Head(workspacesGroup, "/load-balancers/{id}/certmagic/objects/*key", s.restHeadCertmagicObject, allowLoadBalancerTokenModifier)
	huma.Get(workspacesGroup, "/load-balancers/{id}/certmagic/objects/*key", s.restGetCertmagicObject, allowLoadBalancerTokenModifier)
	huma.Put(workspacesGroup, "/load-balancers/{id}/certmagic/objects/*key", s.restPutCertmagicObject, allowLoadBalancerTokenModifier)
	huma.Delete(workspacesGroup, "/load-balancers/{id}/certmagic/objects/*key", s.restDeleteCertmagicObject, allowLoadBalancerTokenModifier)

	return nil
}

func (s *LoadBalancerServer) restCreateLoadBalancer(c context.Context, i *huma_utils.JsonBody[models.CreateLoadBalancer]) (*huma_utils.JsonBody[models.LoadBalancer], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), nil)
	}

	if i.Body.LoadBalancerType != "caddy" {
		return nil, huma.Error400BadRequest("invalid load_balancer_type, must be 'caddy'", nil)
	}

	slog.InfoContext(c, "creating load balancer", slog.Any("name", i.Body.Name))

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

	lb := &dmodel.LoadBalancer{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Name:             i.Body.Name,
		LoadBalancerType: string(i.Body.LoadBalancerType),
		NetworkId:        network.ID,
		HttpPort:         i.Body.HttpPort,
		HttpsPort:        i.Body.HttpsPort,
		Replicas:         i.Body.Replicas,
	}

	err = lb.Create(q)
	if err != nil {
		return nil, err
	}

	ret := models.LoadBalancerFromDB(*lb)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *LoadBalancerServer) restListLoadBalancers(c context.Context, i *struct{}) (*huma_utils.List[models.LoadBalancer], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	proxies, err := dmodel.ListLoadBalancersForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.LoadBalancer
	for _, p := range proxies {
		ret = append(ret, *models.LoadBalancerFromDB(p))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *LoadBalancerServer) restGetLoadBalancer(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.LoadBalancer], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret := models.LoadBalancerFromDB(*lb)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdateLoadBalancerInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateLoadBalancer]
}

func (s *LoadBalancerServer) restUpdateLoadBalancer(c context.Context, i *restUpdateLoadBalancerInput) (*huma_utils.JsonBody[models.LoadBalancer], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
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
	httpPort := lb.HttpPort
	httpsPort := lb.HttpsPort
	if i.Body.HttpPort != nil {
		httpPort = *i.Body.HttpPort
	}
	if i.Body.HttpsPort != nil {
		httpsPort = *i.Body.HttpsPort
	}
	if httpPort == httpsPort {
		return nil, huma.Error400BadRequest("http_port and https_port can't be the same", nil)
	}

	err = lb.Update(q, i.Body.HttpPort, i.Body.HttpsPort, i.Body.Replicas)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, lb)
	if err != nil {
		return nil, err
	}

	ret := models.LoadBalancerFromDB(*lb)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *LoadBalancerServer) restDeleteLoadBalancer(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	lbServices, err := dmodel.ListLoadBalancerServicesForLoadBalancer(q, lb.ID)
	if err != nil {
		return nil, err
	}
	if len(lbServices) != 0 {
		return nil, huma.Error400BadRequest("can't delete load balancers with active services")
	}

	err = dmodel.SoftDeleteWithConstraintsByIds[*dmodel.LoadBalancer](q, &lb.WorkspaceID, lb.ID)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
