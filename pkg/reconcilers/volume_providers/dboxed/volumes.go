package dboxed

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func (r *Reconciler) ReconcileVolume(ctx context.Context, v *dmodel.Volume) error {
	q := querier.GetQuerier(ctx)

	log := r.log

	if v.Dboxed.Status.VolumeID != nil {
		log = log.With(slog.Any("dboxedVolumeId", *v.Dboxed.Status.VolumeID))
	}

	if v.DeletedAt.Valid {
		return r.reconcileDeleteVolume(ctx, log, v)
	}

	err := dmodel.AddFinalizer(q, v, "dboxed-volume")
	if err != nil {
		return err
	}

	if v.Dboxed.Status.VolumeID == nil {
		err := r.createDboxedVolume(ctx, log, v)
		if err != nil {
			return err
		}
	}

	err = r.updateDboxedVolumeStatus(ctx, log, v)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileDeleteVolume(ctx context.Context, log *slog.Logger, v *dmodel.Volume) error {
	q := querier.GetQuerier(ctx)

	if v.Dboxed.Status.VolumeID == nil {
		err := dmodel.RemoveFinalizer(q, v, "dboxed-volume")
		if err != nil {
			return err
		}
		return nil
	}

	log.InfoContext(ctx, "deleting dboxed volume")
	//err := r.client.DeleteVolume(ctx, r.vp.Dboxed.RepositoryId.V, *v.Dboxed.Status.VolumeID)
	//if err != nil {
	//	return err
	//}
	//err = v.Dboxed.Status.UpdateVolumeID(q, nil)
	//if err != nil {
	//	return err
	//}
	//
	//err = soft_delete.RemoveFinalizer(q, v, "dboxed-volume")
	//if err != nil {
	//	return err
	//}

	return nil
}

func (r *Reconciler) createDboxedVolume(ctx context.Context, log *slog.Logger, v *dmodel.Volume) error {
	//q := querier.GetQuerier(ctx)
	//
	//name := fmt.Sprintf("dboxed-%s", v.Uuid)
	//
	//log.InfoContext(ctx, "creating dboxed volume")
	//dv, err := r.client.CreateVolume(ctx, r.vp.Dboxed.RepositoryId.V, vmodels.CreateVolume{
	//	Name:   name,
	//	FsSize: v.Dboxed.FsSize.V,
	//	FsType: v.Dboxed.FsType.V,
	//})
	//if err != nil {
	//	return err
	//}
	//log.InfoContext(ctx, "dboxed volume created", slog.Any("dboxedVolumeId", dv.ID))
	//err = v.Dboxed.Status.UpdateVolumeID(q, &dv.ID)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (r *Reconciler) updateDboxedVolumeStatus(ctx context.Context, log *slog.Logger, v *dmodel.Volume) error {
	//q := querier.GetQuerier(ctx)
	//
	//if v.Dboxed.Status.VolumeID == nil {
	//	return nil
	//}
	//
	//dv, err := r.client.GetVolumeById(ctx, r.vp.Dboxed.RepositoryId.V, *v.Dboxed.Status.VolumeID)
	//if err != nil {
	//	return err
	//}
	//
	//if v.Dboxed.Status.FsSize == nil || v.Dboxed.Status.FsType == nil || dv.FsSize != *v.Dboxed.Status.FsSize || dv.FsType != *v.Dboxed.Status.FsType {
	//	log.InfoContext(ctx, "updating dboxed volume status", slog.Any("fsSize", dv.FsSize), slog.Any("fsType", dv.FsType))
	//	err = v.Dboxed.Status.UpdateInfo(q, &dv.FsSize, &dv.FsType)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}
