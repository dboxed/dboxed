package netbird

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/util"
	api2 "github.com/netbirdio/netbird/shared/management/http/api"
)

func (r *Reconciler) policyName(ctx context.Context) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-network-%s", config.InstanceName, r.n.ID)
}

func (r *Reconciler) reconcileNetbirdPolicies(ctx context.Context, delete bool) base.ReconcileResult {
	groupName := r.networkGroup(ctx)
	policyName := r.policyName(ctx)

	ep, ok := r.nbPoliciesByName[policyName]
	if ok {
		if delete {
			r.log.InfoContext(ctx, "deleting netbird policy", slog.Any("policyName", policyName))
			err := r.netbirdClient.Policies.Delete(ctx, *ep.Id)
			if err != nil {
				return base.ErrorFromMessage("failed to delete Netbird policy %s: %s", *ep.Id, err.Error())
			}
		}
		return base.ReconcileResult{}
	}

	if delete {
		return base.ReconcileResult{}
	}

	g, ok := r.nbGroupsByName[groupName]
	if !ok {
		return base.ErrorFromMessage("group %s not found", groupName)
	}
	groupIds := []string{g.Id}

	r.log.InfoContext(ctx, "creating netbird policy", slog.Any("policyName", policyName))
	ep, err := r.netbirdClient.Policies.Create(ctx, api2.PostApiPoliciesJSONRequestBody{
		Name:        policyName,
		Description: util.Ptr(fmt.Sprintf("dboxed policy to allow access between %s peers", policyName)),
		Enabled:     true,
		Rules: []api2.PolicyRuleUpdate{
			{
				Enabled:       true,
				Action:        api2.PolicyRuleUpdateActionAccept,
				Bidirectional: true,
				Destinations:  &groupIds,
				Sources:       &groupIds,
				Protocol:      api2.PolicyRuleUpdateProtocolAll,
			},
		},
	})
	if err != nil {
		return base.ErrorFromMessage("failed to create Netbird poolicy %s: %s", policyName, err.Error())
	}
	r.nbPoliciesByName[policyName] = ep

	return base.ReconcileResult{}
}
