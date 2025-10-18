package dmodel

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeSnapshot struct {
	OwnedByWorkspace

	VolumeProviderID querier.NullForJoin[int64]  `db:"volume_provider_id"`
	VolumedID        querier.NullForJoin[int64]  `db:"volume_id"`
	LockID           querier.NullForJoin[string] `db:"lock_id"`

	Rustic *VolumeSnapshotRustic `join:"true"`
}

type VolumeSnapshotRustic struct {
	ID querier.NullForJoin[int64] `db:"id"`

	SnapshotId       querier.NullForJoin[string]    `db:"snapshot_id"`
	SnapshotTime     querier.NullForJoin[time.Time] `db:"snapshot_time"`
	ParentSnapshotId *string                        `db:"parent_snapshot_id"`

	Hostname querier.NullForJoin[string] `db:"hostname"`

	FilesNew            querier.NullForJoin[int] `db:"files_new"`
	FilesChanged        querier.NullForJoin[int] `db:"files_changed"`
	FilesUnmodified     querier.NullForJoin[int] `db:"files_unmodified"`
	TotalFilesProcessed querier.NullForJoin[int] `db:"total_files_processed"`
	TotalBytesProcessed querier.NullForJoin[int] `db:"total_bytes_processed"`

	DirsNew               querier.NullForJoin[int] `db:"dirs_new"`
	DirsChanged           querier.NullForJoin[int] `db:"dirs_changed"`
	DirsUnmodified        querier.NullForJoin[int] `db:"dirs_unmodified"`
	TotalDirsProcessed    querier.NullForJoin[int] `db:"total_dirs_processed"`
	TotalDirsizeProcessed querier.NullForJoin[int] `db:"total_dirsize_processed"`
	DataBlobs             querier.NullForJoin[int] `db:"data_blobs"`
	TreeBlobs             querier.NullForJoin[int] `db:"tree_blobs"`
	DataAdded             querier.NullForJoin[int] `db:"data_added"`
	DataAddedPacked       querier.NullForJoin[int] `db:"data_added_packed"`
	DataAddedFiles        querier.NullForJoin[int] `db:"data_added_files"`
	DataAddedFilesPacked  querier.NullForJoin[int] `db:"data_added_files_packed"`
	DataAddedTrees        querier.NullForJoin[int] `db:"data_added_trees"`
	DataAddedTreesPacked  querier.NullForJoin[int] `db:"data_added_trees_packed"`

	BackupStart    querier.NullForJoin[time.Time] `db:"backup_start"`
	BackupEnd      querier.NullForJoin[time.Time] `db:"backup_end"`
	BackupDuration querier.NullForJoin[float32]   `db:"backup_duration"`
	TotalDuration  querier.NullForJoin[float32]   `db:"total_duration"`
}

func (v *VolumeSnapshot) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *VolumeSnapshotRustic) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func GetVolumeSnapshotById(q *querier.Querier, workspaceId *int64, volumeId *int64, id int64, skipDeleted bool) (*VolumeSnapshot, error) {
	return querier.GetOne[VolumeSnapshot](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"volume_id":    querier.OmitIfNull(volumeId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func GetVolumeSnapshotBySnapshotId(q *querier.Querier, snapshotId string, skipDeleted bool) (*VolumeSnapshot, error) {
	return querier.GetOne[VolumeSnapshot](q, map[string]any{
		"snapshot_id": snapshotId,
		"deleted_at":  querier.ExcludeNonNull(skipDeleted),
	})
}

func ListVolumeSnapshotsForProvider(q *querier.Querier, workspaceId *int64, providerId int64, skipDeleted bool) ([]VolumeSnapshot, error) {
	return querier.GetMany[VolumeSnapshot](q, map[string]any{
		"workspace_id":       querier.OmitIfNull(workspaceId),
		"volume_provider_id": providerId,
		"deleted_at":         querier.ExcludeNonNull(skipDeleted),
	}, nil)
}

func ListVolumeSnapshotsForVolume(q *querier.Querier, workspaceId *int64, volumeId int64, skipDeleted bool) ([]VolumeSnapshot, error) {
	return querier.GetManySorted[VolumeSnapshot](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"volume_id":    volumeId,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	}, &querier.SortAndPage{
		Sort: []querier.SortField{
			{Field: "created_at", Direction: querier.SortOrderDesc},
		},
	})
}
