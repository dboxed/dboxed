package ingress_proxies

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
)

func (r *reconciler) reconcileBoxReplicas(ctx context.Context, proxy *dmodel.IngressProxy, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	network, err := dmodel.GetNetworkById(q, nil, proxy.NetworkId, true)
	if err != nil {
		return base.InternalError(err)
	}

	existingReplicas, err := dmodel.ListIngressProxyBoxesForProxy(q, proxy.ID)
	if err != nil {
		return base.InternalError(err)
	}

	for len(existingReplicas) > proxy.Replicas {
		result := r.softDeleteReplica(ctx, proxy, existingReplicas[len(existingReplicas)-1].BoxId, log)
		if result.Error != nil {
			return result
		}
		existingReplicas = existingReplicas[:len(existingReplicas)-1]
	}
	for len(existingReplicas) < proxy.Replicas {
		result := base.Transaction(ctx, func(ctx context.Context) base.ReconcileResult {
			q := querier.GetQuerier(ctx)
			boxName := fmt.Sprintf("ingress-proxy-%s-%d", proxy.Name, len(existingReplicas)+1)

			log.InfoContext(ctx, "creating box for ingress proxy")
			box, inputErr, err := boxes_utils.CreateBox(ctx, proxy.WorkspaceID, models.CreateBox{
				Name:    boxName,
				Network: &network.ID,
			}, global.BoxTypeIngressProxy)
			if err != nil {
				return base.InternalError(err)
			}
			if inputErr != "" {
				return base.ErrorFromMessage(inputErr)
			}
			newReplica := dmodel.IngressProxyBox{
				IngressProxyId: proxy.ID,
				BoxId:          box.ID,
			}
			err = newReplica.Create(q)
			if err != nil {
				return base.InternalError(err)
			}
			existingReplicas = append(existingReplicas, newReplica)
			return base.ReconcileResult{}
		})
		if result.Error != nil {
			return result
		}
	}

	for _, replica := range existingReplicas {
		box, err := dmodel.GetBoxById(q, nil, replica.BoxId, true)
		if err != nil {
			return base.InternalError(err)
		}
		result := r.reconcileBoxReplica(ctx, proxy, box, log)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileBoxReplica(ctx context.Context, proxy *dmodel.IngressProxy, box *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	log = log.With("boxId", box.ID, "boxName", box.Name)

	result := r.reconcileBoxPortForwards(ctx, proxy, box, log)
	if result.Error != nil {
		return result
	}

	switch global.IngressProxyType(proxy.ProxyType) {
	case global.IngressProxyCaddy:
		result = r.reconcileProxyCaddy(ctx, proxy, box, log)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

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
