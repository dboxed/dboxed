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
	huma.Post(workspacesGroup, "/volumes", s.restCreateVolume)
	huma.Get(workspacesGroup, "/volumes", s.restListVolumes)
	huma.Get(workspacesGroup, "/volumes/{id}", s.restGetVolume)
	huma.Get(workspacesGroup, "/volumes/by-name/{volumeName}", s.restGetVolumeByName)
	huma.Delete(workspacesGroup, "/volumes/{id}", s.restDeleteVolume)

	huma.Post(workspacesGroup, "/volumes/{id}/lock", s.restLockVolume)
	huma.Post(workspacesGroup, "/volumes/{id}/release", s.restReleaseVolume)

	huma.Get(workspacesGroup, "/volumes/{id}/snapshots", s.restListSnapshots)
	huma.Get(workspacesGroup, "/volumes/{id}/snapshots/{snapshotId}", s.restGetSnapshot)
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

	ret := models.VolumeFromDB(*v, nil)

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

	switch dmodel.VolumeProviderType(vp.Type) {
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
		mm := models.VolumeFromDB(r.Volume, r.Attachment)
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

	m := models.VolumeFromDB(v.Volume, v.Attachment)
	return huma_utils.NewJsonBody(m), nil
}

type VolumeName struct {
	VolumeName string `path:"volumeName"`
}

func (s *VolumeServer) restGetVolumeByName(c context.Context, i *VolumeName) (*huma_utils.JsonBody[models.Volume], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	v, err := dmodel.GetVolumeByName(q, w.ID, i.VolumeName, true)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment)
	return huma_utils.NewJsonBody(m), nil
}

func (s *VolumeServer) restDeleteVolume(ctx context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Volume](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
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

	m := models.VolumeFromDB(v.Volume, v.Attachment)
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

	log := slog.With(slog.Any("volId", v.ID))

	if v.LockId == nil || *v.LockId != i.Body.LockId {
		return nil, huma.Error404NotFound("lock id does not match")
	}

	log.Info("releasing volume", slog.Any("lockId", i.Body.LockId))

	err = v.UpdateLock(q, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	m := models.VolumeFromDB(v.Volume, v.Attachment)
	return huma_utils.NewJsonBody(m), nil
}
