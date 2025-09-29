package netbird

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
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

func (r *Reconciler) reconcileCommon(ctx context.Context, log *slog.Logger, n *dmodel.Network) error {
	r.log = log
	r.n = n

	err := r.buildNetbirdApiClient()
	if err != nil {
		return err
	}

	err = r.queryNetbirdGroups(ctx)
	if err != nil {
		return err
	}
	err = r.queryNetbirdPolicies(ctx)
	if err != nil {
		return err
	}
	err = r.queryNetbirdPeers(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, log *slog.Logger, n *dmodel.Network) error {
	err := r.reconcileCommon(ctx, log, n)
	if err != nil {
		return err
	}

	err = r.reconcileNetbirdGroups(ctx)
	if err != nil {
		return err
	}

	err = r.reconcileNetbirdPolicies(ctx, false)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ReconcileDelete(ctx context.Context, log *slog.Logger, n *dmodel.Network) error {
	err := r.reconcileCommon(ctx, log, n)
	if err != nil {
		return err
	}

	config := config.GetConfig(ctx)
	groupIds, err := r.groupIds(r.groupsToDelete(ctx), false)
	if err != nil {
		return err
	}

	networkGroupId, ok := r.nbGroupsByName[fmt.Sprintf("%s-network-%d", config.InstanceName, r.n.ID)]
	if ok {
		peers, err := r.netbirdClient.Peers.List(ctx)
		if err != nil {
			return err
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
				return err
			}
		}

		setupKeys, err := r.netbirdClient.SetupKeys.List(ctx)
		if err != nil {
			return err
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
				return err
			}
		}
	}

	err = r.reconcileNetbirdPolicies(ctx, true)
	if err != nil {
		return err
	}

	for _, id := range groupIds {
		err = r.netbirdClient.Groups.Delete(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) buildNetbirdApiClient() error {
	r.netbirdClient = rest.NewWithBearerToken(r.n.Netbird.ApiUrl.V, r.n.Netbird.ApiAccessToken.V)
	return nil
}
