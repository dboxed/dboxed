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
	"github.com/netbirdio/netbird/shared/management/client/rest"
	"github.com/netbirdio/netbird/shared/management/http/api"
)

type Reconciler struct {
	log *slog.Logger

	n *dmodel.Network

	nbGroupsByName   map[string]*api.Group
	nbPoliciesByName map[string]*api.Policy
	setupKeysById    map[string]*api.SetupKey
	usedSetupKeys    map[string]struct{}

	netbirdClient *rest.Client
}

func (r *Reconciler) queryNetbirdResources(ctx context.Context) base.ReconcileResult {
	result := r.queryNetbirdGroups(ctx)
	if result.Error != nil {
		return result
	}
	result = r.queryNetbirdPolicies(ctx)
	if result.Error != nil {
		return result
	}
	result = r.queryNetbirdPeers(ctx)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileNetwork(ctx context.Context, log *slog.Logger, n *dmodel.Network) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	r.log = log
	r.n = n

	r.buildNetbirdApiClient()

	if n.GetDeletedAt() != nil {
		return r.reconcileDeleteNetwork(ctx)
	}

	result := r.queryNetbirdResources(ctx)
	if result.Error != nil {
		return result
	}

	err := dmodel.AddFinalizer(q, n, "netbird-cleanup")
	if err != nil {
		return base.InternalError(err)
	}

	result = r.reconcileNetbirdGroups(ctx)
	if result.Error != nil {
		return result
	}

	result = r.reconcileNetbirdPolicies(ctx, false)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) reconcileDeleteNetwork(ctx context.Context) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	if !r.n.HasFinalizer("netbird-cleanup") {
		return base.ReconcileResult{}
	}

	result := r.queryNetbirdResources(ctx)
	if result.Error != nil {
		return result
	}

	config := config.GetConfig(ctx)
	groupIds, result := r.groupIds(r.groupsToDelete(ctx), false)
	if result.Error != nil {
		return result
	}

	networkGroupId, ok := r.nbGroupsByName[fmt.Sprintf("%s-network-%d", config.InstanceName, r.n.ID)]
	if ok {
		peers, err := r.netbirdClient.Peers.List(ctx)
		if err != nil {
			return base.ErrorFromMessage("failed to list Netbird peers: %s", err.Error())
		}
		for _, p := range peers {
			if !slices.ContainsFunc(p.Groups, func(g api.GroupMinimum) bool {
				return g.Id == networkGroupId.Id
			}) {
				continue
			}
			slog.InfoContext(ctx, "deleting netbird peer",
				slog.Any("peerId", p.Id),
				slog.Any("peerName", p.Name),
			)
			err = r.netbirdClient.Peers.Delete(ctx, p.Id)
			if err != nil {
				return base.ErrorFromMessage("failed to delete Netbird peer %s: %s", p.Id, err.Error())
			}
		}

		setupKeys, err := r.netbirdClient.SetupKeys.List(ctx)
		if err != nil {
			return base.ErrorFromMessage("failed to list Netbird setup keys: %s", err.Error())
		}
		for _, sk := range setupKeys {
			if !slices.Contains(sk.AutoGroups, networkGroupId.Id) {
				continue
			}
			slog.InfoContext(ctx, "deleting netbird setup key",
				slog.Any("keyId", sk.Id),
				slog.Any("keyName", sk.Name),
			)
			err = r.netbirdClient.SetupKeys.Delete(ctx, sk.Id)
			if err != nil {
				return base.ErrorFromMessage("failed to delete Netbird setup key %s: %s", sk.Id, err.Error())
			}
		}
	}

	result = r.reconcileNetbirdPolicies(ctx, true)
	if result.Error != nil {
		return result
	}

	for _, id := range groupIds {
		err := r.netbirdClient.Groups.Delete(ctx, id)
		if err != nil {
			return base.ErrorFromMessage("failed to delete Netbird group %s: %s", id, err.Error())
		}
	}

	err := dmodel.RemoveFinalizer(q, r.n, "netbird-cleanup")
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) buildNetbirdApiClient() {
	r.netbirdClient = rest.NewWithBearerToken(r.n.Netbird.ApiUrl.V, r.n.Netbird.ApiAccessToken.V)
}
