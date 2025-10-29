package hetzner

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type Reconciler struct {
	log *slog.Logger

	mp *dmodel.MachineProvider

	sshKeyId int64

	hetznerServersById   map[int64]*hcloud.Server
	hetznerServersByName map[string]*hcloud.Server

	hcloudClient *hcloud.Client
}

func (r *Reconciler) reconcileCommon(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error {
	r.log = log
	r.mp = mp

	r.log = slog.With(slog.Any("id", r.mp.ID), slog.Any("name", r.mp.Name))
	r.log = slog.With(slog.Any("hetznerNetworkName", r.mp.Hetzner.HetznerNetworkName))
	if r.mp.Hetzner.Status.HetznerNetworkID != nil {
		r.log = slog.With(slog.Any("hetznerNetworkId", *r.mp.Hetzner.Status.HetznerNetworkID))
	}

	r.hcloudClient = hcloud.NewClient(hcloud.WithToken(r.mp.Hetzner.HcloudToken.V))

	return nil
}
func (r *Reconciler) ReconcileMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) base.ReconcileResult {
	err := r.reconcileCommon(ctx, log, mp)
	if err != nil {
		return base.ReconcileResult{Error: err}
	}

	result := r.reconcileSshKey(ctx)
	if result.Error != nil {
		return result
	}

	result = r.reconcileHetznerNetwork(ctx)
	if result.Error != nil {
		return result
	}

	result = r.queryHetznerServers(ctx)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}
