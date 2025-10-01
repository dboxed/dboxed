package dmodel

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Volume struct {
	OwnedByWorkspace
	ReconcileStatus

	Uuid string `db:"uuid"`
	Name string `db:"name"`

	VolumeProviderID   int64  `db:"volume_provider_id"`
	VolumeProviderType string `db:"volume_provider_type"`

	LockId      *string    `db:"lock_id"`
	LockTime    *time.Time `db:"lock_time"`
	LockBoxUuid *string    `db:"lock_box_uuid"`

	LatestSnapshotId *int64 `db:"latest_snapshot_id"`

	Rustic *VolumeRustic `join:"true"`
}

type VolumeRustic struct {
	ID querier.NullForJoin[int64] `db:"id"`

	FsSize querier.NullForJoin[int64]  `db:"fs_size"`
	FsType querier.NullForJoin[string] `db:"fs_type"`

	Status *VolumeRusticStatus `join:"true"`
}

type VolumeRusticStatus struct {
	ID querier.NullForJoin[int64] `db:"id"`
}

type VolumeWithAttachment struct {
	Volume

	Attachment *BoxVolumeAttachment `join:"true" join_left_field:"id" join_right_table:"box_volume_attachment" join_right_field:"volume_id"`
}

func (x *VolumeWithAttachment) GetTableName() string {
	return "volume"
}

type BoxVolumeAttachment struct {
	BoxId    querier.NullForJoin[int64] `db:"box_id"`
	VolumeId querier.NullForJoin[int64] `db:"volume_id"`

	RootUid  querier.NullForJoin[int64]  `db:"root_uid"`
	RootGid  querier.NullForJoin[int64]  `db:"root_gid"`
	RootMode querier.NullForJoin[string] `db:"root_mode"`
}

type BoxVolumeAttachmentWithVolume struct {
	BoxVolumeAttachment
	Volume Volume `join:"true" join_left_field:"volume_id"`
}

func (v *BoxVolumeAttachment) GetTableName() string {
	return "box_volume_attachment"
}

func (v *Volume) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *VolumeRustic) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *VolumeRusticStatus) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *BoxVolumeAttachment) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func GetVolumeById(q *querier.Querier, workspaceId *int64, id int64, skipDeleted bool) (*VolumeWithAttachment, error) {
	return querier.GetOne[VolumeWithAttachment](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func GetVolumeByName(q *querier.Querier, workspaceId int64, name string, skipDeleted bool) (*VolumeWithAttachment, error) {
	return querier.GetOne[VolumeWithAttachment](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumesForWorkspace(q *querier.Querier, workspaceId int64, skipDeleted bool) ([]VolumeWithAttachment, error) {
	return querier.GetMany[VolumeWithAttachment](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumesForVolumeProvider(q *querier.Querier, volumeProviderId int64, skipDeleted bool) ([]VolumeWithAttachment, error) {
	return querier.GetMany[VolumeWithAttachment](q, map[string]any{
		"volume_provider_id": volumeProviderId,
		"deleted_at":         querier.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxVolumeAttachments(q *querier.Querier, boxId int64) ([]BoxVolumeAttachmentWithVolume, error) {
	return querier.GetMany[BoxVolumeAttachmentWithVolume](q, map[string]any{
		"box_id": boxId,
	})
}

func GetBoxVolumeAttachment(q *querier.Querier, boxId int64, volumeId int64) (*BoxVolumeAttachmentWithVolume, error) {
	return querier.GetOne[BoxVolumeAttachmentWithVolume](q, map[string]any{
		"box_id":    boxId,
		"volume_id": volumeId,
	})
}

func (v *Volume) UpdateLock(q *querier.Querier, newLockId *string, newLockTime *time.Time, boxUuid *string) error {
	oldLockId := v.LockId
	oldLockTime := v.LockTime
	v.LockId = newLockId
	v.LockTime = nil
	v.LockTime = newLockTime
	v.LockBoxUuid = boxUuid
	return querier.UpdateOneByFields[Volume](q, map[string]any{
		"id":        v.ID,
		"lock_id":   oldLockId,
		"lock_time": oldLockTime,
	}, map[string]any{
		"lock_id":       v.LockId,
		"lock_time":     v.LockTime,
		"lock_box_uuid": v.LockBoxUuid,
	})
}

func (v *Volume) UpdateLatestSnapshot(q *querier.Querier, snapshotId *int64) error {
	oldSnapshotId := v.LatestSnapshotId
	v.LatestSnapshotId = snapshotId
	return querier.UpdateOneByFields[Volume](q, map[string]any{
		"id":                 v.ID,
		"latest_snapshot_id": oldSnapshotId,
	}, map[string]any{
		"latest_snapshot_id": snapshotId,
	})
}

func (v *BoxVolumeAttachment) Update(q *querier.Querier, rootUid *int64, rootGid *int64, rootMode *string) error {
	var fields []string
	if rootUid != nil {
		v.RootUid = querier.N(*rootUid)
		fields = append(fields, "root_uid")
	}
	if rootGid != nil {
		v.RootGid = querier.N(*rootGid)
		fields = append(fields, "root_gid")
	}
	if rootMode != nil {
		v.RootMode = querier.N(*rootMode)
		fields = append(fields, "root_mode")
	}
	return querier.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id":    v.BoxId,
		"volume_id": v.VolumeId,
	}, v, fields...)
}
