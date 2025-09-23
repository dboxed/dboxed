package dboxed

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Reconciler struct {
	log *slog.Logger

	vp *dmodel.VolumeProvider

	//client *client.Client
}

func (r *Reconciler) reconcileCommon(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider) error {
	r.log = log
	r.vp = vp

	err := r.buildDboxedVolumeClient(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider) error {
	err := r.reconcileCommon(ctx, log, vp)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) buildDboxedVolumeClient(ctx context.Context) error {
	//var err error
	//r.client, err = client.New(&r.vp.Dboxed.ApiUrl.V, &r.vp.Dboxed.Token.V)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (r *Reconciler) ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider) error {
	err := r.reconcileCommon(ctx, log, vp)
	if err != nil {
		return err
	}

	return nil
}
