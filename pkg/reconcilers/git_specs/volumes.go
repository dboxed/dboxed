package git_specs

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/models/dboxed_specs"
	"github.com/dboxed/dboxed/pkg/server/resources/volumes"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dustin/go-humanize"
	"github.com/kluctl/kluctl/lib/git/types"
)

func (r *reconciler) reconcileSpecVolume(ctx context.Context, gs *dmodel.GitSpec, u types.GitUrl, name string, volume *dboxed_specs.Volume, e *dmodel.GitSpecMapping, log *slog.Logger) base.ReconcileResult {
	if e == nil {
		return r.createVolume(ctx, gs, u, name, volume, log)
	} else {
		var oldVolume *dboxed_specs.Volume
		err := json.Unmarshal([]byte(e.Spec), &oldVolume)
		if err != nil {
			return base.InternalError(err)
		}
		if volume.Recreate != oldVolume.Recreate {
			deleted, result := r.deleteObject(ctx, e, log)
			if result.ExitReconcile() {
				return result
			}
			if deleted {
				return r.createVolume(ctx, gs, u, name, volume, log)
			}
		} else {
			if !util.EqualsViaJson(volume, oldVolume) {
				return base.ErrorFromMessage("volume %s has been modified, which is not allowed", name)
			}
		}
		return base.ReconcileResult{}
	}
}

func (r *reconciler) createVolume(ctx context.Context, gs *dmodel.GitSpec, u types.GitUrl, name string, volume *dboxed_specs.Volume, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	vp, err := dmodel.GetVolumeProviderByName(q, gs.WorkspaceID, volume.Provider, true)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to retrieve volume provider with name '%s'", volume.Provider)
	}

	createArgs := models.CreateVolume{
		Name:           name,
		VolumeProvider: vp.ID,
	}
	switch vp.Type {
	case dmodel.VolumeProviderTypeRustic:
		if volume.Rustic == nil {
			return base.ErrorFromMessage("missing rustic config for volumes %s", name)
		}
		fsSize, err := humanize.ParseBytes(volume.Rustic.FsSize)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to parse fsSize: %s", err.Error())
		}
		createArgs.Rustic = &models.CreateVolumeRustic{
			FsSize: int64(fsSize),
			FsType: volume.Rustic.FsType,
		}
	default:
		return base.ErrorFromMessage("unknown volume provider type %s", vp.Type)
	}

	log.InfoContext(ctx, "creating volume", "volumeName", name)
	dbVolume, err := volumes.CreateVolume(ctx, gs.WorkspaceID, createArgs)
	if err != nil {
		return base.InternalError(err)
	}

	_, result := r.createMapping(ctx, gs.WorkspaceID, u.RepoKey(), volume.Recreate, "volume", dbVolume.ID, name, volume)
	if result.ExitReconcile() {
		return result
	}
	return base.ReconcileResult{}
}
