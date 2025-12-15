package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type LoadBalancer struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Name             string                  `json:"name"`
	LoadBalancerType dmodel.LoadBalancerType `json:"loadBalancerType"`
	Network          string                  `json:"network"`
	HttpPort         int                     `json:"httpPort"`
	HttpsPort        int                     `json:"httpsPort"`
	Replicas         int                     `json:"replicas"`
}

type CreateLoadBalancer struct {
	Name             string                  `json:"name"`
	LoadBalancerType dmodel.LoadBalancerType `json:"loadBalancerType"`
	Network          string                  `json:"network"`
	HttpPort         int                     `json:"httpPort"`
	HttpsPort        int                     `json:"httpsPort"`
	Replicas         int                     `json:"replicas,omitempty"`
}

type UpdateLoadBalancer struct {
	HttpPort  *int `json:"httpPort,omitempty"`
	HttpsPort *int `json:"httpsPort,omitempty"`
	Replicas  *int `json:"replicas,omitempty"`
}

func LoadBalancerFromDB(p dmodel.LoadBalancer) *LoadBalancer {
	return &LoadBalancer{
		ID:               p.ID,
		Workspace:        p.WorkspaceID,
		CreatedAt:        p.CreatedAt,
		Status:           p.ReconcileStatus.ReconcileStatus.V,
		StatusDetails:    p.ReconcileStatus.ReconcileStatusDetails.V,
		Name:             p.Name,
		LoadBalancerType: dmodel.LoadBalancerType(p.LoadBalancerType),
		Network:          p.NetworkId,
		HttpPort:         p.HttpPort,
		HttpsPort:        p.HttpsPort,
		Replicas:         p.Replicas,
	}
}

type LoadBalancerService struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	LoadBalancerID string `json:"loadBalancerId"`
	BoxID          string `json:"boxId"`

	Description *string `json:"description,omitempty"`

	Hostname   string `json:"hostname"`
	PathPrefix string `json:"pathPrefix"`
	Port       int    `json:"port"`
}

type CreateLoadBalancerService struct {
	LoadBalancerID string  `json:"loadBalancerId"`
	Description    *string `json:"description,omitempty"`
	Hostname       string  `json:"hostname"`
	PathPrefix     string  `json:"pathPrefix"`
	Port           int     `json:"port"`
}

type UpdateLoadBalancerService struct {
	Description *string `json:"description,omitempty"`
	Hostname    *string `json:"hostname,omitempty"`
	PathPrefix  *string `json:"pathPrefix,omitempty"`
	Port        *int    `json:"port,omitempty"`
}

func LoadBalancerServiceFromDB(bi dmodel.LoadBalancerService) *LoadBalancerService {
	return &LoadBalancerService{
		ID:             bi.ID,
		CreatedAt:      bi.CreatedAt,
		LoadBalancerID: bi.LoadBalancerId,
		BoxID:          bi.BoxID,
		Description:    bi.Description,
		Hostname:       bi.Hostname,
		PathPrefix:     bi.PathPrefix,
		Port:           bi.Port,
	}
}
