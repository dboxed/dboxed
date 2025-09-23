package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Volume struct {
	OwnedByWorkspace
	ReconcileStatus

	Uuid string `db:"uuid"`
	Name string `db:"name"`

	VolumeProviderID   int64  `db:"volume_provider_id"`
	VolumeProviderType string `db:"volume_provider_type"`

	Dboxed *VolumeDboxed `join:"true"`
}

type VolumeWithAttachment struct {
	Volume

	Attachment *BoxVolumeAttachment `join:"true" join_left_field:"id" join_right_table:"box_volume_attachment" join_right_field:"volume_id"`
}

func (x *VolumeWithAttachment) GetTableName() string {
	return "volume"
}

type VolumeDboxed struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	FsSize querier2.NullForJoin[int64]  `db:"fs_size"`
	FsType querier2.NullForJoin[string] `db:"fs_type"`

	Status *VolumeDboxedStatus `join:"true"`
}

type VolumeDboxedStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	VolumeID *int64  `db:"volume_id"`
	FsSize   *int64  `db:"fs_size"`
	FsType   *string `db:"fs_type"`
}

type BoxVolumeAttachment struct {
	BoxId    querier2.NullForJoin[int64] `db:"box_id"`
	VolumeId querier2.NullForJoin[int64] `db:"volume_id"`

	RootUid  querier2.NullForJoin[int64]  `db:"root_uid"`
	RootGid  querier2.NullForJoin[int64]  `db:"root_gid"`
	RootMode querier2.NullForJoin[string] `db:"root_mode"`
}

type BoxVolumeAttachmentWithVolume struct {
	BoxVolumeAttachment
	Volume Volume `join:"true" join_left_field:"volume_id"`
}

func (v *BoxVolumeAttachment) GetTableName() string {
	return "box_volume_attachment"
}

func (v *Volume) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *VolumeDboxed) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *VolumeDboxedStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *BoxVolumeAttachment) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetVolumeById(q *querier2.Querier, workspaceId *int64, id int64, skipDeleted bool) (*VolumeWithAttachment, error) {
	return querier2.GetOne[VolumeWithAttachment](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumesForWorkspace(q *querier2.Querier, workspaceId int64, skipDeleted bool) ([]VolumeWithAttachment, error) {
	return querier2.GetMany[VolumeWithAttachment](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumesForVolumeProvider(q *querier2.Querier, volumeProviderId int64, skipDeleted bool) ([]VolumeWithAttachment, error) {
	return querier2.GetMany[VolumeWithAttachment](q, map[string]any{
		"volume_provider_id": volumeProviderId,
		"deleted_at":         querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxVolumeAttachments(q *querier2.Querier, boxId int64) ([]BoxVolumeAttachmentWithVolume, error) {
	return querier2.GetMany[BoxVolumeAttachmentWithVolume](q, map[string]any{
		"box_id": boxId,
	})
}

func (v *VolumeDboxedStatus) UpdateVolumeID(q *querier2.Querier, id *int64) error {
	v.VolumeID = id
	return querier2.UpdateOneFromStruct(q, v,
		"volume_id",
	)
}

func (v *VolumeDboxedStatus) UpdateInfo(q *querier2.Querier, fsSize *int64, fsType *string) error {
	v.FsSize = fsSize
	v.FsType = fsType
	return querier2.UpdateOneFromStruct(q, v,
		"fs_size",
		"fs_type",
	)
}

func GetBoxVolumeAttachment(q *querier2.Querier, boxId int64, volumeId int64) (*BoxVolumeAttachmentWithVolume, error) {
	return querier2.GetOne[BoxVolumeAttachmentWithVolume](q, map[string]any{
		"box_id":    boxId,
		"volume_id": volumeId,
	})
}

func (v *BoxVolumeAttachment) Update(q *querier2.Querier, rootUid *int64, rootGid *int64, rootMode *string) error {
	var fields []string
	if rootUid != nil {
		v.RootUid = querier2.N(*rootUid)
		fields = append(fields, "root_uid")
	}
	if rootGid != nil {
		v.RootGid = querier2.N(*rootGid)
		fields = append(fields, "root_gid")
	}
	if rootMode != nil {
		v.RootMode = querier2.N(*rootMode)
		fields = append(fields, "root_mode")
	}
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id":    v.BoxId,
		"volume_id": v.VolumeId,
	}, v, fields...)
}
