package dmodel

import (
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Volume struct {
	OwnedByWorkspace
	SoftDeleteFields
	ReconcileStatus

	Name string `db:"name"`

	VolumeProviderID   string             `db:"volume_provider_id"`
	VolumeProviderType VolumeProviderType `db:"volume_provider_type"`

	MountId *string `db:"mount_id"`

	LatestSnapshotId *string `db:"latest_snapshot_id"`

	Rustic *VolumeRustic `join:"true"`
}

type VolumeRustic struct {
	ID querier.NullForJoin[string] `db:"id"`

	FsSize querier.NullForJoin[int64]  `db:"fs_size"`
	FsType querier.NullForJoin[string] `db:"fs_type"`

	Status *VolumeRusticStatus `join:"true"`
}

type VolumeRusticStatus struct {
	ID querier.NullForJoin[string] `db:"id"`
}

type VolumeWithJoins struct {
	Volume

	Attachment  *BoxVolumeAttachment `join:"true" join_left_field:"id" join_right_table:"box_volume_attachment" join_right_field:"volume_id"`
	MountStatus *VolumeMountStatus   `join:"true" db:"mount_status" join_left_field:"mount_id" join_right_table:"volume_mount_status" join_right_field:"mount_id"`
}

func (x *VolumeWithJoins) GetTableName() string {
	return "volume"
}

type BoxVolumeAttachment struct {
	BoxId    querier.NullForJoin[string] `db:"box_id"`
	VolumeId querier.NullForJoin[string] `db:"volume_id"`

	RootUid  querier.NullForJoin[int64]  `db:"root_uid"`
	RootGid  querier.NullForJoin[int64]  `db:"root_gid"`
	RootMode querier.NullForJoin[string] `db:"root_mode"`
}

type BoxVolumeAttachmentWithJoins struct {
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

func GetVolumeById(q *querier.Querier, workspaceId *string, id string, skipDeleted bool) (*Volume, error) {
	return querier.GetOne[Volume](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func GetVolumeWithDetailsById(q *querier.Querier, workspaceId *string, id string, skipDeleted bool) (*VolumeWithJoins, error) {
	return querier.GetOne[VolumeWithJoins](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func GetVolumeByName(q *querier.Querier, workspaceId string, name string, skipDeleted bool) (*VolumeWithJoins, error) {
	return querier.GetOne[VolumeWithJoins](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumesByMountBoxId(q *querier.Querier, workspaceId *string, boxId string, skipDeleted bool) ([]VolumeWithJoins, error) {
	return querier.GetMany[VolumeWithJoins](q, map[string]any{
		"workspace_id":        querier.OmitIfNull(workspaceId),
		"mount_status.box_id": boxId,
		"deleted_at":          querier.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListVolumesForWorkspace(q *querier.Querier, workspaceId string, skipDeleted bool) ([]VolumeWithJoins, error) {
	return querier.GetMany[VolumeWithJoins](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListVolumesForVolumeProvider(q *querier.Querier, volumeProviderId string, skipDeleted bool) ([]VolumeWithJoins, error) {
	return querier.GetMany[VolumeWithJoins](q, map[string]any{
		"volume_provider_id": volumeProviderId,
		"deleted_at":         querier.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListBoxVolumeAttachments(q *querier.Querier, boxId string) ([]BoxVolumeAttachmentWithJoins, error) {
	return querier.GetMany[BoxVolumeAttachmentWithJoins](q, map[string]any{
		"box_id": boxId,
	}, nil)
}

func GetBoxVolumeAttachment(q *querier.Querier, boxId string, volumeId string) (*BoxVolumeAttachmentWithJoins, error) {
	return querier.GetOne[BoxVolumeAttachmentWithJoins](q, map[string]any{
		"box_id":    boxId,
		"volume_id": volumeId,
	})
}

func (v *Volume) UpdateMountId(q *querier.Querier, newMountId *string) error {
	oldMountId := v.MountId
	v.MountId = newMountId
	return querier.UpdateOneByFields[Volume](q, map[string]any{
		"id":       v.ID,
		"mount_id": oldMountId,
	}, map[string]any{
		"mount_id": v.MountId,
	})
}

func (v *Volume) ForceReleaseMount(q *querier.Querier) error {
	v.MountId = nil
	return querier.UpdateOneFromStruct[Volume](q, v,
		"mount_id",
	)
}

func (v *Volume) UpdateLatestSnapshot(q *querier.Querier, snapshotId *string) error {
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
	if rootUid != nil && *rootUid != v.RootUid.V {
		v.RootUid = querier.N(*rootUid)
		fields = append(fields, "root_uid")
	}
	if rootGid != nil && *rootGid != v.RootGid.V {
		v.RootGid = querier.N(*rootGid)
		fields = append(fields, "root_gid")
	}
	if rootMode != nil && *rootMode != v.RootMode.V {
		v.RootMode = querier.N(*rootMode)
		fields = append(fields, "root_mode")
	}
	if len(fields) == 0 {
		return nil
	}
	return querier.UpdateOneByFieldsFromStruct(q, map[string]any{
		"box_id":    v.BoxId,
		"volume_id": v.VolumeId,
	}, v, fields...)
}
