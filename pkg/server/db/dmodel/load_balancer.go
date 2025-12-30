package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type LoadBalancer struct {
	OwnedByWorkspace
	SoftDeleteFields
	ReconcileStatus

	Name             string `db:"name"`
	LoadBalancerType string `db:"load_balancer_type"`
	NetworkId        string `db:"network_id"`

	HttpPort  int `db:"http_port"`
	HttpsPort int `db:"https_port"`

	Replicas int `db:"replicas"`
}

type LoadBalancerBox struct {
	LoadBalancerId string `db:"load_balancer_id"`
	BoxId          string `db:"box_id"`
}

type LoadBalancerService struct {
	ID string `db:"id" uuid:"true"`
	Times

	LoadBalancerId string `db:"load_balancer_id"`
	BoxID          string `db:"box_id"`

	Description *string `db:"description"`

	Hostname   string `db:"hostname"`
	PathPrefix string `db:"path_prefix"`
	Port       int    `db:"port"`
}

func (v LoadBalancerService) GetId() string {
	return v.ID
}

func (v *LoadBalancer) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *LoadBalancer) Update(q *querier2.Querier, httpPort *int, httpsPort *int, replicas *int) error {
	var fields []string
	if httpPort != nil {
		v.HttpPort = *httpPort
		fields = append(fields, "http_port")
	}
	if httpsPort != nil {
		v.HttpsPort = *httpsPort
		fields = append(fields, "https_port")
	}
	if replicas != nil {
		v.Replicas = *replicas
		fields = append(fields, "replicas")
	}
	if len(fields) == 0 {
		return nil
	}
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"workspace_id": v.WorkspaceID,
		"id":           v.ID,
	}, v, fields...)
}

func CheckLoadBalancerById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) error {
	return querier2.CheckOne[LoadBalancer](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetLoadBalancerById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*LoadBalancer, error) {
	return querier2.GetOne[LoadBalancer](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetLoadBalancerByName(q *querier2.Querier, workspaceId *string, name string, skipDeleted bool) (*LoadBalancer, error) {
	return querier2.GetOne[LoadBalancer](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"name":         name,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListLoadBalancersForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]LoadBalancer, error) {
	return querier2.GetMany[LoadBalancer](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func (v *LoadBalancerService) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListLoadBalancerServices(q *querier2.Querier, boxId string) ([]LoadBalancerService, error) {
	return querier2.GetMany[LoadBalancerService](q, map[string]any{
		"box_id": boxId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("id", querier2.SortOrderAsc),
	})
}

func ListLoadBalancerServicesForLoadBalancer(q *querier2.Querier, loadBalancerId string) ([]LoadBalancerService, error) {
	return querier2.GetMany[LoadBalancerService](q, map[string]any{
		"load_balancer_id": loadBalancerId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("id", querier2.SortOrderAsc),
	})
}

func GetLoadBalancerService(q *querier2.Querier, boxId string, id string) (*LoadBalancerService, error) {
	return querier2.GetOne[LoadBalancerService](q, map[string]any{
		"box_id": boxId,
		"id":     id,
	})
}

func (v *LoadBalancerService) Update(q *querier2.Querier, description *string, hostname *string, pathPrefix *string, port *int) error {
	var fields []string
	if description != nil {
		v.Description = description
		fields = append(fields, "description")
	}
	if hostname != nil {
		v.Hostname = *hostname
		fields = append(fields, "hostname")
	}
	if pathPrefix != nil {
		v.PathPrefix = *pathPrefix
		fields = append(fields, "path_prefix")
	}
	if port != nil {
		v.Port = *port
		fields = append(fields, "port")
	}
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id": v.BoxID,
		"id":     v.ID,
	}, v, fields...)
}

func (v *LoadBalancerBox) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListLoadBalancerBoxesForLoadBalancer(q *querier2.Querier, loadBalancerId string) ([]LoadBalancerBox, error) {
	return querier2.GetMany[LoadBalancerBox](q, map[string]any{
		"load_balancer_id": loadBalancerId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("box_id", querier2.SortOrderAsc),
	})
}
