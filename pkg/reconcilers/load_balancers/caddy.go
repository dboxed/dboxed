package load_balancers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/load_balancers/files"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (r *reconciler) reconcileLoadBalancerCaddy(ctx context.Context, lb *dmodel.LoadBalancer, box *dmodel.Box, token *models.Token, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	composeFile, result := r.buildCaddyCompose(ctx, lb, box, token, log)
	if result.Error != nil {
		return result
	}

	existingComposeProject, err := dmodel.GetBoxComposeProjectByName(q, box.ID, "caddy")
	if err != nil {
		if !querier.IsSqlNotFoundError(err) {
			return base.InternalError(err)
		}
	}

	if existingComposeProject == nil {
		bcp := &dmodel.BoxComposeProject{
			BoxID:          box.ID,
			Name:           "caddy",
			ComposeProject: composeFile,
		}
		err = bcp.Create(q)
		if err != nil {
			return base.InternalError(err)
		}
	} else {
		if composeFile != existingComposeProject.ComposeProject {
			err = existingComposeProject.UpdateComposeProject(q, composeFile)
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) buildCaddyCompose(ctx context.Context, lb *dmodel.LoadBalancer, proxyBox *dmodel.Box, token *models.Token, log *slog.Logger) (string, base.ReconcileResult) {
	cf, result := r.buildCaddyfile(ctx, lb, token, log)
	if result.Error != nil {
		return "", result
	}

	ret, err := files.GetCaddyComposeFile("latest", cf)
	if err != nil {
		return "", base.InternalError(err)
	}

	return ret, base.ReconcileResult{}
}

func (r *reconciler) buildCaddyfile(ctx context.Context, lb *dmodel.LoadBalancer, token *models.Token, log *slog.Logger) (string, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)
	cfg := config.GetConfig(ctx)

	cf, err := files.GetCaddyfile(cfg.Server.BaseUrl, *token.Token, lb.WorkspaceID, lb.ID)
	if err != nil {
		return "", base.InternalError(err)
	}

	lbServices, err := dmodel.ListLoadBalancerServicesForLoadBalancer(q, lb.ID)
	if err != nil {
		return "", base.InternalError(err)
	}

	boxes := map[string]*dmodel.Box{}

	for _, bi := range lbServices {
		box, ok := boxes[bi.BoxID]
		if !ok {
			box, err = dmodel.GetBoxById(q, nil, bi.BoxID, true)
			if err != nil {
				return "", base.InternalError(err)
			}
			boxes[bi.BoxID] = box
		}

		matcher := bi.PathPrefix
		if !strings.HasSuffix(matcher, "/") {
			matcher += "/"
		}
		matcher += "*"
		boxFqdn := fmt.Sprintf("%s.dboxed", box.Name)

		cf += fmt.Sprintf("%s {\n", bi.Hostname)
		cf += fmt.Sprintf("  reverse_proxy %s %s:%d\n", matcher, boxFqdn, bi.Port)
		cf += fmt.Sprintf("}\n\n")
	}

	return cf, base.ReconcileResult{}
}
