package hetzner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers/userdata"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func (r *Reconciler) queryHetznerServers(ctx context.Context) base.ReconcileResult {
	hetznerServers, err := r.hcloudClient.Server.AllWithOpts(ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: fmt.Sprintf("%s=%s", cloud_utils.MachineProviderIdTagName, r.mp.ID),
		},
	})
	if err != nil {
		return base.ErrorWithMessage(err, "failed to query Hetzner servers: %s", err.Error())
	}
	r.hetznerServersById = map[int64]*hcloud.Server{}
	r.hetznerServersByName = map[string]*hcloud.Server{}
	for _, s := range hetznerServers {
		r.hetznerServersById[s.ID] = s
		r.hetznerServersByName[s.Name] = s
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileMachine(ctx context.Context, log *slog.Logger, m *dmodel.Machine) {
	result := r.doReconcileMachine(ctx, m)
	base.SetReconcileResult(ctx, log, m.Hetzner, result)
}

func (r *Reconciler) doReconcileMachine(ctx context.Context, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	log := r.log.With(slog.Any("machined", m.ID))
	if m.Hetzner.Status.ServerID != nil {
		log = log.With(slog.Any("hetznerServerId", *m.Hetzner.Status.ServerID))
	}

	if m.DeletedAt.Valid {
		return r.reconcileDeleteHetznerMachine(ctx, log, m)
	}

	err := dmodel.AddFinalizer(q, m, "hetzner-machine")
	if err != nil {
		return base.InternalError(err)
	}

	if m.Hetzner.Status.ServerID == nil {
		// check if it was actually created but we somehow failed to store the ID
		server, ok := r.hetznerServersByName[r.buildHetznerServerName(ctx, m)]
		if ok {
			err = m.Hetzner.Status.UpdateServerID(q, &server.ID)
			if err != nil {
				return base.InternalError(err)
			}
		} else {
			result := r.createHetznerServer(ctx, log, m)
			if result.Error != nil {
				return result
			}
		}
	}

	return r.updateHetznerServer(ctx, log, m)
}

func (r *Reconciler) reconcileDeleteHetznerMachine(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	if m.Hetzner.Status.ServerID == nil {
		err := dmodel.RemoveFinalizer(q, m, "hetzner-machine")
		if err != nil {
			return base.InternalError(err)
		}
		return base.ReconcileResult{}
	}
	server, ok := r.hetznerServersById[*m.Hetzner.Status.ServerID]
	if !ok {
		log.InfoContext(ctx, "hetzner server already vanished")
		err := dmodel.RemoveFinalizer(q, m, "hetzner-machine")
		if err != nil {
			return base.InternalError(err)
		}
	} else {
		log.InfoContext(ctx, "deleting hetzner server")
		_, _, err := r.hcloudClient.Server.DeleteWithResult(ctx, &hcloud.Server{
			ID: *m.Hetzner.Status.ServerID,
		})
		if err != nil && !hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			return base.ErrorWithMessage(err, "failed to delete Hetzner server: %s", err.Error())
		}

		delete(r.hetznerServersById, server.ID)
		delete(r.hetznerServersByName, server.Name)
	}

	var err error
	err = m.Hetzner.Status.UpdateServerID(q, nil)
	if err != nil {
		return base.InternalError(err)
	}

	err = dmodel.RemoveFinalizer(q, m, "hetzner-machine")
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) buildHetznerServerName(ctx context.Context, m *dmodel.Machine) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-%s-%s", config.InstanceName, m.Name, m.ID)
}

func (r *Reconciler) createHetznerServer(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	image := "ubuntu-24.04"

	box := m.Box

	ud := userdata.GetUserdata(
		box.DboxedVersion,
		"dummy",
	)

	log.InfoContext(ctx, "creating hetzner server")
	labels := cloud_utils.BuildCloudMachineTags(r.mp.Hetzner.ID.V, m)
	createOpts := hcloud.ServerCreateOpts{
		Name: r.buildHetznerServerName(ctx, m),
		ServerType: &hcloud.ServerType{
			Name: m.Hetzner.ServerType.V,
		},
		Image:            &hcloud.Image{Name: image},
		Location:         &hcloud.Location{Name: m.Hetzner.ServerLocation.V},
		StartAfterCreate: util.Ptr(true),
		Labels:           labels,
		Networks: []*hcloud.Network{
			{ID: *r.mp.Hetzner.Status.HetznerNetworkID},
		},
		PublicNet: &hcloud.ServerCreatePublicNet{
			EnableIPv4: true,
		},
		UserData: ud,
	}

	if r.sshKeyId != -1 {
		createOpts.SSHKeys = append(createOpts.SSHKeys, &hcloud.SSHKey{ID: r.sshKeyId})
	}
	createResult, _, err := r.hcloudClient.Server.Create(ctx, createOpts)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to create Hetzner server: %s", err.Error())
	}

	r.hetznerServersById[createResult.Server.ID] = createResult.Server
	r.hetznerServersByName[createResult.Server.Name] = createResult.Server

	err = m.Hetzner.Status.UpdateServerID(q, &createResult.Server.ID)
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) updateHetznerServer(ctx context.Context, log *slog.Logger, m *dmodel.Machine) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	_, ok := r.hetznerServersById[*m.Hetzner.Status.ServerID]
	if !ok {
		log.InfoContext(ctx, "hetzner server disappeared, removing server id and expecting it to be re-created")
		var err error
		err = m.Hetzner.Status.UpdateServerID(q, nil)
		if err != nil {
			return base.InternalError(err)
		}
		return base.ReconcileResult{}
	}

	return base.ReconcileResult{}
}
