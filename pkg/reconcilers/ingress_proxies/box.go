package ingress_proxies

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func (r *reconciler) reconcileBoxPortForwards(ctx context.Context, proxy *dmodel.IngressProxy, box *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingPortForwards, err := dmodel.ListBoxPortForwards(q, box.ID)
	if err != nil {
		return base.InternalError(err)
	}

	addOrUpdatePortForward := func(desc string, protocol string, hostPort int, sandboxPort int) error {
		var existing *dmodel.BoxPortForward
		for _, pf := range existingPortForwards {
			if pf.Description != nil && *pf.Description == desc {
				existing = &pf
				break
			}
		}
		log := log.With("description", desc, "hostPort", hostPort, "sandboxPort", sandboxPort)
		if existing == nil {
			pf := &dmodel.BoxPortForward{
				BoxID:         box.ID,
				Description:   &desc,
				Protocol:      protocol,
				HostPortFirst: hostPort,
				HostPortLast:  hostPort,
				SandboxPort:   sandboxPort,
			}
			log.Info("creating port-forward for ingress-proxy")
			return pf.Create(q)
		}
		if hostPort != existing.HostPortFirst || hostPort != existing.HostPortLast || sandboxPort != existing.SandboxPort {
			log.Info("updating port-forward for ingress-proxy")
			return existing.Update(q, &desc, nil, &hostPort, &hostPort, &sandboxPort)
		}
		return nil
	}

	err = addOrUpdatePortForward("http-tcp", "tcp", proxy.HttpPort, 80)
	if err != nil {
		return base.InternalError(err)
	}
	err = addOrUpdatePortForward("https-tcp", "tcp", proxy.HttpsPort, 443)
	if err != nil {
		return base.InternalError(err)
	}
	err = addOrUpdatePortForward("https-udp", "udp", proxy.HttpsPort, 443)
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}
