package netbird

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/netbirdio/netbird/shared/management/http/api"
)

func (r *Reconciler) queryNetbirdPeers(ctx context.Context) error {
	l, err := r.netbirdClient.SetupKeys.List(ctx)
	if err != nil {
		return err
	}
	r.setupKeysById = map[string]*api.SetupKey{}
	for _, sk := range l {
		r.setupKeysById[sk.Id] = &sk
	}
	r.usedSetupKeys = map[string]struct{}{}
	return nil
}

func (r *Reconciler) Cleanup(ctx context.Context) error {
	groups, err := r.groupIds(r.desiredGroups(ctx), true)
	if err != nil {
		return err
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
			err = r.netbirdClient.SetupKeys.Delete(ctx, sk.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) ReconcileBox(ctx context.Context, box *dmodel.Box) error {
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
				return err
			}
		}
	}

	if box.Netbird.SetupKeyID == nil {
		autoGroups, err := r.groupIds(r.desiredGroups(ctx), true)
		if err != nil {
			return err
		}

		log.InfoContext(ctx, "creating netbird setup key")
		sk, err := r.netbirdClient.SetupKeys.Create(ctx, api.CreateSetupKeyRequest{
			Ephemeral:  util.Ptr(true),
			Name:       fmt.Sprintf("%s-%s-%s", config.InstanceName, r.n.Name, box.Name),
			Type:       "reusable",
			AutoGroups: autoGroups,
		})
		if err != nil {
			return err
		}
		err = box.Netbird.UpdateSetupKey(q, &sk.Key, &sk.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
