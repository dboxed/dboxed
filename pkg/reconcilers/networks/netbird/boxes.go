package netbird

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/netbirdio/netbird/shared/management/http/api"
)

func (r *Reconciler) querySetupKeys(ctx context.Context) base.ReconcileResult {
	l, err := r.netbirdClient.SetupKeys.List(ctx)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to list Netbird setup keys: %s", err.Error())
	}
	r.setupKeysById = map[string]*api.SetupKey{}
	for _, sk := range l {
		r.setupKeysById[sk.Id] = &sk
	}
	r.usedSetupKeys = map[string]struct{}{}
	return base.ReconcileResult{}
}

func (r *Reconciler) queryPeers(ctx context.Context) base.ReconcileResult {
	l, err := r.netbirdClient.Peers.List(ctx)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to list Netbird peers: %s", err.Error())
	}
	r.peersById = map[string]*api.Peer{}
	r.peersByName = map[string]*api.Peer{}
	for _, peer := range l {
		r.peersById[peer.Id] = &peer
		r.peersByName[peer.Name] = &peer
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) Cleanup(ctx context.Context) base.ReconcileResult {
	groups, result := r.groupIds(r.desiredGroups(ctx), true)
	if result.Error != nil {
		return result
	}
	for _, sk := range r.setupKeysById {
		allFound := true
		for _, g := range groups {
			if !slices.Contains(sk.AutoGroups, g) {
				allFound = false
				break
			}
		}
		if !allFound {
			continue
		}
		if _, ok := r.usedSetupKeys[sk.Id]; !ok {
			r.log.InfoContext(ctx, "deleting unused netbird setup key", slog.Any("setupKeyId", sk.Id))
			err := r.netbirdClient.SetupKeys.Delete(ctx, sk.Id)
			if err != nil {
				return base.ErrorWithMessage(err, "failed to delete Netbird setup key %s: %s", sk.Id, err.Error())
			}
		}
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileBox(ctx context.Context, log *slog.Logger, box *dmodel.Box) {
	result := r.doReconcileBox(ctx, box)
	base.SetReconcileResult(ctx, log, box.Netbird, result)
}

func (r *Reconciler) doReconcileBox(ctx context.Context, box *dmodel.Box) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	config := config.GetConfig(ctx)

	if box.Netbird.SetupKeyID != nil {
		r.usedSetupKeys[*box.Netbird.SetupKeyID] = struct{}{}
	}

	log := r.log.With(slog.Any("machined", box.ID))
	if box.Netbird.SetupKeyID != nil {
		log = log.With(slog.Any("netbirdSetupKeyId", *box.Netbird.SetupKeyID))
	}

	if box.Netbird.SetupKeyID != nil {
		_, ok := r.setupKeysById[*box.Netbird.SetupKeyID]
		if !ok {
			log.InfoContext(ctx, "netbird setup key vanished, setting to null to force re-creation")
			err := box.Netbird.UpdateSetupKey(q, nil, nil)
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	if box.Netbird.SetupKeyID == nil {
		autoGroups, result := r.groupIds(r.desiredGroups(ctx), true)
		if result.Error != nil {
			return result
		}

		name := fmt.Sprintf("%s-%s-%s", config.InstanceName, r.n.Name, box.Name)
		log.InfoContext(ctx, "creating netbird setup key", "name", name)
		sk, err := r.netbirdClient.SetupKeys.Create(ctx, api.CreateSetupKeyRequest{
			AllowExtraDnsLabels: util.Ptr(true),
			Ephemeral:           util.Ptr(true),
			Name:                name,
			Type:                "reusable",
			AutoGroups:          autoGroups,
		})
		if err != nil {
			return base.ErrorWithMessage(err, "failed to create Netbird setup key %s: %s", name, err.Error())
		}
		err = box.Netbird.UpdateSetupKey(q, &sk.Key, &sk.Id)
		if err != nil {
			return base.InternalError(err)
		}
	}

	return base.ReconcileResult{}
}
