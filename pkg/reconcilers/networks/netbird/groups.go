package netbird

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	api2 "github.com/netbirdio/netbird/shared/management/http/api"
)

func (r *Reconciler) workspaceGroup(ctx context.Context) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-workspace-%s", config.InstanceName, r.n.WorkspaceID)
}

func (r *Reconciler) networkGroup(ctx context.Context) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-network-%s", config.InstanceName, r.n.ID)
}

func (r *Reconciler) desiredGroups(ctx context.Context) []string {
	var ret []string
	ret = append(ret, r.workspaceGroup(ctx))
	ret = append(ret, r.networkGroup(ctx))
	return ret
}

func (r *Reconciler) groupsToDelete(ctx context.Context) []string {
	config := config.GetConfig(ctx)

	var ret []string
	ret = append(ret, fmt.Sprintf("%s-network-%s", config.InstanceName, r.n.ID))
	return ret
}

func (r *Reconciler) groupIds(groupNames []string, failOnMissing bool) ([]string, base.ReconcileResult) {
	var ret []string
	for _, g := range groupNames {
		g2, ok := r.nbGroupsByName[g]
		if ok {
			ret = append(ret, g2.Id)
		} else {
			if failOnMissing {
				return nil, base.ErrorFromMessage("netbird group with name %s not found", g)
			}
		}
	}
	return ret, base.ReconcileResult{}
}

func (r *Reconciler) queryNetbirdGroups(ctx context.Context) base.ReconcileResult {
	existingGroups, err := r.netbirdClient.Groups.List(ctx)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to list Netbird groups: %s", err.Error())
	}
	r.nbGroupsByName = map[string]*api2.Group{}
	for _, g := range existingGroups {
		r.nbGroupsByName[g.Name] = &g
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) queryNetbirdPolicies(ctx context.Context) base.ReconcileResult {
	existingPolicies, err := r.netbirdClient.Policies.List(ctx)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to list Netbird policies: %s", err.Error())
	}
	r.nbPoliciesByName = map[string]*api2.Policy{}
	for _, p := range existingPolicies {
		r.nbPoliciesByName[p.Name] = &p
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) reconcileNetbirdGroups(ctx context.Context) base.ReconcileResult {
	for _, groupName := range r.desiredGroups(ctx) {
		g := r.nbGroupsByName[groupName]
		if g == nil {
			r.log.InfoContext(ctx, "creating netbird group", slog.Any("name", groupName))

			g, err := r.netbirdClient.Groups.Create(ctx, api2.GroupRequest{
				Name: groupName,
			})
			if err != nil {
				return base.ErrorWithMessage(err, "failed to create Netbird group %s: %s", groupName, err.Error())
			}
			r.nbGroupsByName[groupName] = g
		}
	}

	return base.ReconcileResult{}
}
