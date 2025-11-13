package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type IngressProxy struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	BoxID     string                  `json:"boxId"`
	Name      string                  `json:"name"`
	ProxyType global.IngressProxyType `json:"proxyType"`
	HttpPort  int                     `json:"httpPort"`
	HttpsPort int                     `json:"httpsPort"`
}

type CreateIngressProxy struct {
	Name      string                  `json:"name"`
	ProxyType global.IngressProxyType `json:"proxyType"`
	Network   string                  `json:"network"`
	HttpPort  int                     `json:"httpPort"`
	HttpsPort int                     `json:"httpsPort"`
}

type UpdateIngressProxy struct {
	HttpPort  *int `json:"httpPort,omitempty"`
	HttpsPort *int `json:"httpsPort,omitempty"`
}

func IngressProxyFromDB(p dmodel.IngressProxy) *IngressProxy {
	return &IngressProxy{
		ID:            p.ID,
		Workspace:     p.WorkspaceID,
		CreatedAt:     p.CreatedAt,
		Status:        p.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: p.ReconcileStatus.ReconcileStatusDetails.V,
		BoxID:         p.BoxID,
		Name:          p.Name,
		ProxyType:     global.IngressProxyType(p.ProxyType),
		HttpPort:      p.HttpPort,
		HttpsPort:     p.HttpsPort,
	}
}

type BoxIngress struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	ProxyID string `json:"proxyId"`
	BoxID   string `json:"boxId"`

	Description *string `json:"description,omitempty"`

	Hostname   string `json:"hostname"`
	PathPrefix string `json:"pathPrefix"`
	Port       int    `json:"port"`
}

type CreateBoxIngress struct {
	ProxyID     string  `json:"proxyId"`
	Description *string `json:"description,omitempty"`
	Hostname    string  `json:"hostname"`
	PathPrefix  string  `json:"pathPrefix"`
	Port        int     `json:"port"`
}

type UpdateBoxIngress struct {
	Description *string `json:"description,omitempty"`
	Hostname    *string `json:"hostname,omitempty"`
	PathPrefix  *string `json:"pathPrefix,omitempty"`
	Port        *int    `json:"port,omitempty"`
}

func BoxIngressFromDB(bi dmodel.BoxIngress) *BoxIngress {
	return &BoxIngress{
		ID:          bi.ID,
		CreatedAt:   bi.CreatedAt,
		ProxyID:     bi.ProxyID,
		BoxID:       bi.BoxID,
		Description: bi.Description,
		Hostname:    bi.Hostname,
		PathPrefix:  bi.PathPrefix,
		Port:        bi.Port,
	}
}
