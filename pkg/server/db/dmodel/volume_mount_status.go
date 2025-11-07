package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeMountStatus struct {
	VolumeId querier2.NullForJoin[string] `db:"volume_id"`
	MountId  querier2.NullForJoin[string] `db:"mount_id"`
	BoxId    *string                      `db:"box_id"`

	MountTime     querier2.NullForJoin[time.Time] `db:"mount_time"`
	ReleaseTime   *time.Time                      `db:"release_time"`
	ForceReleased querier2.NullForJoin[bool]      `db:"force_released"`

	StatusTime querier2.NullForJoin[time.Time] `db:"status_time"`

	VolumeTotalSize *int64 `db:"volume_total_size"`
	VolumeFreeSize  *int64 `db:"volume_free_size"`

	LastFinishedSnapshotId *string `db:"last_finished_snapshot_id"`

	SnapshotStartTime *time.Time `db:"snapshot_start_time"`
	SnapshotEndTime   *time.Time `db:"snapshot_end_time"`
}

func (v *VolumeMountStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetVolumeMountStatusById(q *querier2.Querier, volumeId string, mountId string) (*VolumeMountStatus, error) {
	return querier2.GetOne[VolumeMountStatus](q, map[string]any{
		"volume_id": volumeId,
		"mount_id":  mountId,
	})
}

func (v *VolumeMountStatus) Update(q *querier2.Querier) error {
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"volume_id": v.VolumeId.V,
		"mount_id":  v.MountId.V,
	}, v,
		"release_time",
		"force_released",
		"status_time",
		"volume_total_size",
		"volume_free_size",
		"last_snapshot_id",
		"snapshot_start_time",
		"snapshot_end_time",
	)
}

func (v *VolumeMountStatus) UpdateReleaseInfo(q *querier2.Querier, releaseTime *time.Time, forceReleased bool) error {
	oldReleaseTime := v.ReleaseTime
	oldForceReleased := v.ForceReleased.V
	v.ReleaseTime = releaseTime
	v.ForceReleased = querier2.N(forceReleased)
	v.StatusTime = querier2.N(time.Now())
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"volume_id":      v.VolumeId.V,
		"mount_id":       v.MountId.V,
		"release_time":   oldReleaseTime,
		"force_released": oldForceReleased,
	}, v,
		"release_time",
		"force_released",
		"status_time",
	)
}

func (v *VolumeMountStatus) UpdateMountInfo(q *querier2.Querier, totalSize int64, freeSize int64, lastFinishedSnapshotId *string, snapshotStartTime *time.Time, snapshotEndTime *time.Time) error {
	v.VolumeTotalSize = &totalSize
	v.VolumeFreeSize = &freeSize
	v.LastFinishedSnapshotId = lastFinishedSnapshotId
	v.SnapshotStartTime = snapshotStartTime
	v.SnapshotEndTime = snapshotEndTime
	v.StatusTime = querier2.N(time.Now())
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"volume_id": v.VolumeId.V,
		"mount_id":  v.MountId.V,
	}, v,
		"volume_total_size",
		"volume_free_size",
		"last_finished_snapshot_id",
		"snapshot_start_time",
		"snapshot_end_time",
		"status_time",
	)
}
