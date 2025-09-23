package netbird

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/config"
	api2 "github.com/netbirdio/netbird/shared/management/http/api"
)

func (r *Reconciler) workspaceGroup(ctx context.Context) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-workspace-%d", config.InstanceName, r.n.WorkspaceID)
}

func (r *Reconciler) networkGroup(ctx context.Context) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-network-%d", config.InstanceName, r.n.ID)
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
	ret = append(ret, fmt.Sprintf("%s-network-%d", config.InstanceName, r.n.ID))
	return ret
}

func (r *Reconciler) groupIds(groupNames []string, failOnMissing bool) ([]string, error) {
	var ret []string
	for _, g := range groupNames {
		g2, ok := r.nbGroupsByName[g]
		if ok {
			ret = append(ret, g2.Id)
		} else {
			if failOnMissing {
				return nil, fmt.Errorf("netbird group with name %s not found", g)
			}
		}
	}
	return ret, nil
}

func (r *Reconciler) queryNetbirdGroups(ctx context.Context) error {
	existingGroups, err := r.netbirdClient.Groups.List(ctx)
	if err != nil {
		return err
	}
	r.nbGroupsByName = map[string]*api2.Group{}
	for _, g := range existingGroups {
		r.nbGroupsByName[g.Name] = &g
	}
	return nil
}

func (r *Reconciler) queryNetbirdPolicies(ctx context.Context) error {
	existingPolicies, err := r.netbirdClient.Policies.List(ctx)
	if err != nil {
		return err
	}
	r.nbPoliciesByName = map[string]*api2.Policy{}
	for _, p := range existingPolicies {
		r.nbPoliciesByName[p.Name] = &p
	}
	return nil
}

func (r *Reconciler) reconcileNetbirdGroups(ctx context.Context) error {
	for _, groupName := range r.desiredGroups(ctx) {
		g := r.nbGroupsByName[groupName]
		if g == nil {
			r.log.InfoContext(ctx, "creating netbird group", slog.Any("name", groupName))

			g, err := r.netbirdClient.Groups.Create(ctx, api2.GroupRequest{
				Name: groupName,
			})
			if err != nil {
				return err
			}
			r.nbGroupsByName[groupName] = g
		}
	}

	return nil
}
