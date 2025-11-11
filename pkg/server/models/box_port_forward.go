package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type BoxPortForward struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	BoxID       string    `json:"boxId"`
	Description *string   `json:"description,omitempty"`

	Protocol      string `json:"protocol"`
	HostPortFirst int    `json:"hostPortFirst"`
	HostPortLast  int    `json:"hostPortLast"`
	SandboxPort   int    `json:"sandboxPort"`
}

type CreateBoxPortForward struct {
	Description *string `json:"description,omitempty"`

	Protocol      string `json:"protocol"`
	HostPortFirst int    `json:"hostPortFirst"`
	HostPortLast  int    `json:"hostPortLast"`
	SandboxPort   int    `json:"sandboxPort"`
}

type UpdateBoxPortForward struct {
	Description *string `json:"description,omitempty"`

	Protocol      *string `json:"protocol,omitempty"`
	HostPortFirst *int    `json:"hostPortFirst,omitempty"`
	HostPortLast  *int    `json:"hostPortLast,omitempty"`
	SandboxPort   *int    `json:"sandboxPort,omitempty"`
}

func BoxPortForwardFromDB(pf dmodel.BoxPortForward) *BoxPortForward {
	return &BoxPortForward{
		ID:            pf.ID.V,
		CreatedAt:     pf.CreatedAt,
		BoxID:         pf.BoxID,
		Description:   pf.Description,
		Protocol:      pf.Protocol,
		HostPortFirst: pf.HostPortFirst,
		HostPortLast:  pf.HostPortLast,
		SandboxPort:   pf.SandboxPort,
	}
}
