package volumes

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
)

type VolumeServer struct {
}

func New(config config.Config) *VolumeServer {
	s := &VolumeServer{}
	return s
}

func (s *VolumeServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	allowBoxTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true)

	huma.Post(workspacesGroup, "/volumes", s.restCreateVolume)
	huma.Get(workspacesGroup, "/volumes", s.restListVolumes, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/volumes/{id}", s.restGetVolume, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/volumes/by-uuid/{uuid}", s.restGetVolumeByUuid, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/volumes/by-name/{name}", s.restGetVolumeByName, allowBoxTokenModifier)
	huma.Delete(workspacesGroup, "/volumes/{id}", s.restDeleteVolume)

	huma.Post(workspacesGroup, "/volumes/{id}/lock", s.restLockVolume, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/volumes/{id}/release", s.restReleaseVolume, allowBoxTokenModifier)

	huma.Post(workspacesGroup, "/volumes/{id}/snapshots", s.restCreateSnapshot, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/volumes/{id}/snapshots", s.restListSnapshots, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/volumes/{id}/snapshots/{snapshotId}", s.restGetSnapshot, allowBoxTokenModifier)
	huma.Delete(workspacesGroup, "/volumes/{id}/snapshots/{snapshotId}", s.restDeleteSnapshot)

	return nil
}

func (s *VolumeServer) restCreateVolume(ctx context.Context, i *huma_utils.JsonBody[models.CreateVolume]) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(ctx)

	v, inputErr, err := s.createVolume(ctx, i.Body)
	if err != nil {
		return nil, err
	}
	if inputErr != "" {
		return nil, huma.Error400BadRequest(inputErr)
	}

	ret := models.VolumeFromDB(*v, nil, nil)

	err = dmodel.AddChangeTracking(q, v)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(ret), nil
}

func (s *VolumeServer) createVolume(ctx context.Context, body models.CreateVolume) (*dmodel.Volume, string, error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err.Error(), nil
	}

	vp, err := dmodel.GetVolumeProviderById(q, &w.ID, body.VolumeProvider, true)
	if err != nil {
		return nil, "", err
	}

	v := &dmodel.Volume{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Uuid:               uuid.NewString(),
		Name:               body.Name,
		VolumeProviderType: vp.Type,
		VolumeProviderID:   vp.ID,
	}

	err = v.Create(q)
	if err != nil {
		return nil, "", err
	}

	switch vp.Type {
	case dmodel.VolumeProviderTypeRustic:
		if body.Rustic == nil {
			return nil, "missing rustic config", nil
		}
		err = s.createVolumeRustic(ctx, v, *body.Rustic)
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", huma.Error501NotImplemented("volume provider not implemented")
	}

	return v, "", nil
}

func (s *VolumeServer) createVolumeRustic(ctx context.Context, v *dmodel.Volume, body models.CreateVolumeRustic) error {
	q := querier.GetQuerier(ctx)

	if body.FsSize <= humanize.MiByte {
		return huma.Error400BadRequest("fsSize is too small")
	}
	if !slices.Contains(volume.AllowedFsTypes, body.FsType) {
		return huma.Error400BadRequest("unsupported or invalid fsType")
	}

	v.Rustic = &dmodel.VolumeRustic{
		ID:     querier.N(v.ID),
		FsSize: querier.N(body.FsSize),
		FsType: querier.N(body.FsType),
	}
	err := v.Rustic.Create(q)
	if err != nil {
		return err
	}
	return nil
}

func (s *VolumeServer) restListVolumes(ctx context.Context, i *struct{}) (*huma_utils.List[models.Volume], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	l, err := dmodel.ListVolumesForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.Volume
	for _, r := range l {
		err = s.checkBoxToken(ctx, &r.Volume, r.Attachment)
		if err != nil {
			continue
		}

		mm := models.VolumeFromDB(r.Volume, r.Attachment, nil)
		ret = append(ret, mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *VolumeServer) restGetVolume(ctx context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(ctx, &v.Volume, v.Attachment)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment, nil)
	return huma_utils.NewJsonBody(m), nil
}

type VolumeUuid struct {
	VolumeUuid string `path:"uuid"`
}

func (s *VolumeServer) restGetVolumeByUuid(c context.Context, i *VolumeUuid) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	v, err := dmodel.GetVolumeByUuid(q, &w.ID, i.VolumeUuid, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, &v.Volume, v.Attachment)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment, nil)
	return huma_utils.NewJsonBody(m), nil
}

type VolumeName struct {
	VolumeName string `path:"name"`
}

func (s *VolumeServer) restGetVolumeByName(c context.Context, i *VolumeName) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	v, err := dmodel.GetVolumeByName(q, w.ID, i.VolumeName, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, &v.Volume, v.Attachment)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment, nil)
	return huma_utils.NewJsonBody(m), nil
}

func (s *VolumeServer) restDeleteVolume(ctx context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if v.Attachment != nil && v.Attachment.BoxId.Valid {
		return nil, huma.Error400BadRequest("can not delete volume that is attached to a box")
	}

	// make sure we can delete all snapshots (we have a 'on delete restrict' constraint on it)
	err = v.UpdateLatestSnapshot(q, nil)
	if err != nil {
		return nil, err
	}

	extraDeleteVolume := func(q *querier.Querier) error {
		// simulate deletion of all snapshots
		_, err = querier.DeleteManyByFields[dmodel.VolumeSnapshot](q, map[string]any{
			"volume_id": v.Volume.ID,
		})
		return err
	}

	err = dmodel.SoftDeleteWithConstraintsByIdsExtra[*dmodel.Volume](q, &w.ID, i.Id, extraDeleteVolume)
	if err != nil {
		return nil, err
	}

	snapshots, err := dmodel.ListVolumeSnapshotsForVolume(q, &w.ID, v.ID, true)
	if err != nil {
		return nil, err
	}

	for _, s := range snapshots {
		err = dmodel.SoftDeleteWithConstraintsByIds[*dmodel.VolumeSnapshot](q, &w.ID, s.ID)
		if err != nil {
			return nil, err
		}
	}

	return &huma_utils.Empty{}, nil
}

type restLockVolume struct {
	huma_utils.IdByPath
	Body models.VolumeLockRequest
}

func (s *VolumeServer) restLockVolume(ctx context.Context, i *restLockVolume) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(ctx, &v.Volume, v.Attachment)
	if err != nil {
		return nil, err
	}

	log := slog.With(slog.Any("volId", v.ID))

	lockTimeout := time.Minute * 5
	allow := false
	lockId := ""
	if v.LockId == nil {
		allow = true
		lockId = uuid.NewString()
		log = log.With(slog.Any("newLockId", lockId))
		log.Info("locking volume")
	} else {
		if i.Body.PrevLockId == nil && v.LockTime.Add(lockTimeout).Before(time.Now()) {
			allow = true
			lockId = uuid.NewString()
			log = log.With(slog.Any("newLockId", lockId))
			log.Info("old lock expired, re-locking")
		} else if i.Body.PrevLockId != nil && *v.LockId == *i.Body.PrevLockId {
			allow = true
			lockId = *v.LockId
			log.Info("refreshing lock")
		}
	}
	if !allow {
		return nil, huma.Error409Conflict("volume is already locked")
	}

	err = v.UpdateLock(q, &lockId, util.Ptr(time.Now()), i.Body.BoxId)
	if err != nil {
		return nil, err
	}

	vp, err := dmodel.GetVolumeProviderById(q, &w.ID, v.VolumeProviderID, true)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment, vp)
	return huma_utils.NewJsonBody(m), nil
}

type restReleaseVolume struct {
	huma_utils.IdByPath
	Body models.VolumeReleaseRequest
}

func (s *VolumeServer) restReleaseVolume(ctx context.Context, i *restReleaseVolume) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(ctx, &v.Volume, v.Attachment)
	if err != nil {
		return nil, err
	}

	log := slog.With(slog.Any("volId", v.ID))

	if v.LockId == nil || *v.LockId != i.Body.LockId {
		return nil, huma.Error404NotFound("lock id does not match")
	}

	log.Info("releasing volume", slog.Any("lockId", i.Body.LockId))

	err = v.UpdateLock(q, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment, nil)
	return huma_utils.NewJsonBody(m), nil
}

func (s *VolumeServer) checkBoxToken(ctx context.Context, volume *dmodel.Volume, attachment *dmodel.BoxVolumeAttachment) error {
	q := querier.GetQuerier(ctx)
	token := auth.GetToken(ctx)

	if token == nil || token.BoxID != nil {
		// not a box token
		return nil
	}

	if attachment == nil || !attachment.BoxId.Valid {
		return huma.Error403Forbidden("access to volume not allowed")
	}

	box, err := dmodel.GetBoxById(q, &volume.WorkspaceID, *token.BoxID, true)
	if err != nil {
		return err
	}

	if box.WorkspaceID != volume.WorkspaceID || box.ID != attachment.BoxId.V {
		return huma.Error403Forbidden("access to volume not allowed")
	}
	return nil
}
