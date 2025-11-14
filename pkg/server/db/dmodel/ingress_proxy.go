package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type IngressProxy struct {
	OwnedByWorkspace
	ReconcileStatus

	Name      string `db:"name"`
	ProxyType string `db:"proxy_type"`
	NetworkId string `db:"network_id"`

	HttpPort  int `db:"http_port"`
	HttpsPort int `db:"https_port"`

	Replicas int `db:"replicas"`
}

type IngressProxyBox struct {
	IngressProxyId string `db:"ingress_proxy_id"`
	BoxId          string `db:"box_id"`
}

type BoxIngress struct {
	ID string `db:"id" uuid:"true"`
	Times

	ProxyID string `db:"proxy_id"`
	BoxID   string `db:"box_id"`

	Description *string `db:"description"`

	Hostname   string `db:"hostname"`
	PathPrefix string `db:"path_prefix"`
	Port       int    `db:"port"`
}

func (v BoxIngress) GetId() string {
	return v.ID
}

func (v *IngressProxy) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *IngressProxy) Update(q *querier2.Querier, httpPort *int, httpsPort *int, replicas *int) error {
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
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"workspace_id": v.WorkspaceID,
		"id":           v.ID,
	}, v, fields...)
}

func GetIngressProxyById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*IngressProxy, error) {
	return querier2.GetOne[IngressProxy](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListIngressProxiesForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]IngressProxy, error) {
	return querier2.GetMany[IngressProxy](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func (v *BoxIngress) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListBoxIngresses(q *querier2.Querier, boxId string) ([]BoxIngress, error) {
	return querier2.GetMany[BoxIngress](q, map[string]any{
		"box_id": boxId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("id", querier2.SortOrderAsc),
	})
}

func ListBoxIngressesForProxy(q *querier2.Querier, proxyId string) ([]BoxIngress, error) {
	return querier2.GetMany[BoxIngress](q, map[string]any{
		"proxy_id": proxyId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("id", querier2.SortOrderAsc),
	})
}

func GetBoxIngress(q *querier2.Querier, boxId string, id string) (*BoxIngress, error) {
	return querier2.GetOne[BoxIngress](q, map[string]any{
		"box_id": boxId,
		"id":     id,
	})
}

func (v *BoxIngress) Update(q *querier2.Querier, description *string, hostname *string, pathPrefix *string, port *int) error {
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

func (v *IngressProxyBox) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListIngressProxyBoxesForProxy(q *querier2.Querier, proxyId string) ([]IngressProxyBox, error) {
	return querier2.GetMany[IngressProxyBox](q, map[string]any{
		"ingress_proxy_id": proxyId,
	}, &querier2.SortAndPage{
		Sort: querier2.SortBySingleField("box_id", querier2.SortOrderAsc),
	})
}
