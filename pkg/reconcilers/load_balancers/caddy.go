package load_balancers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/load_balancers/files"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func (r *reconciler) reconcileLoadBalancerCaddy(ctx context.Context, lb *dmodel.LoadBalancer, box *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	composeFile, result := r.buildCaddyCompose(ctx, lb, box, log)
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

func (r *reconciler) buildCaddyCompose(ctx context.Context, lb *dmodel.LoadBalancer, proxyBox *dmodel.Box, log *slog.Logger) (string, base.ReconcileResult) {
	cf, result := r.buildCaddyfile(ctx, lb, log)
	if result.Error != nil {
		return "", result
	}
	cfj, err := json.Marshal(cf)
	if err != nil {
		return "", base.InternalError(err)
	}

	ret, err := files.GetCaddyComposeFile("2.10", string(cfj))
	if err != nil {
		return "", base.InternalError(err)
	}

	return ret, base.ReconcileResult{}
}

func (r *reconciler) buildCaddyfile(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) (string, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)

	cf := "#caddyfile\n"

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
