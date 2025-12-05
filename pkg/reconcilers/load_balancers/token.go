package load_balancers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/tokens"
	"github.com/dboxed/dboxed/pkg/util"
)

func (r *reconciler) buildTokenName(lb *dmodel.LoadBalancer) string {
	return tokens.InternalTokenNamePrefix + "lb_" + strings.ReplaceAll(lb.Name, "-", "_")
}

func (r *reconciler) reconcileToken(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) (*models.Token, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)

	tokenName := r.buildTokenName(lb)

	token, err := dmodel.GetTokenByName(q, lb.WorkspaceID, tokenName)
	if err == nil {
		return util.Ptr(models.TokenFromDB(*token, true)), base.ReconcileResult{}
	} else {
		if !querier.IsSqlNotFoundError(err) {
			return nil, base.InternalError(err)
		}
	}

	log.InfoContext(ctx, "creating token for load balancer")

	mtoken, err := tokens.CreateToken(ctx, lb.WorkspaceID, models.CreateToken{
		Name:           r.buildTokenName(lb),
		Type:           dmodel.TokenTypeLoadBalancer,
		LoadBalancerId: &lb.ID,
	}, true, true)
	if err != nil {
		return nil, base.InternalError(err)
	}
	return mtoken, base.ReconcileResult{}
}
