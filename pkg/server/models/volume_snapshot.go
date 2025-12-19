package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type VolumeSnapshot struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	VolumeID string `json:"volumeId"`
	MountId  string `json:"mountId"`

	Restic *VolumeSnapshotRestic `json:"restic,omitempty"`
}

type VolumeSnapshotRestic struct {
	SnapshotId       string    `json:"snapshotId"`
	SnapshotTime     time.Time `json:"snapshotTime"`
	ParentSnapshotId *string   `json:"parentSnapshotId,omitempty"`

	Hostname string `json:"hostname"`

	BackupStart time.Time `json:"backupStart"`
	BackupEnd   time.Time `json:"backupEnd"`

	// statistics from the backup json output
	FilesNew            int   `json:"filesNew"`
	FilesChanged        int   `json:"filesChanged"`
	FilesUnmodified     int   `json:"filesUnmodified"`
	DirsNew             int   `json:"dirsNew"`
	DirsChanged         int   `json:"dirsChanged"`
	DirsUnmodified      int   `json:"dirsUnmodified"`
	DataBlobs           int   `json:"dataBlobs"`
	TreeBlobs           int   `json:"treeBlobs"`
	DataAdded           int64 `json:"dataAdded"`
	DataAddedPacked     int64 `json:"dataAddedPacked"`
	TotalFilesProcessed int   `json:"totalFilesProcessed"`
	TotalBytesProcessed int64 `json:"totalBytesProcessed"`
}

type CreateVolumeSnapshot struct {
	MountId string `json:"mountId"`

	Restic *VolumeSnapshotRestic `json:"restic,omitempty"`
}

func VolumeSnapshotFromDB(v dmodel.VolumeSnapshot) VolumeSnapshot {
	ret := VolumeSnapshot{
		ID:        v.ID,
		Workspace: v.WorkspaceID,
		CreatedAt: v.CreatedAt,
		VolumeID:  v.VolumedID.V,
		MountId:   v.MountID.V,
	}

	if v.Restic != nil && v.Restic.ID.Valid {
		ret.Restic = &VolumeSnapshotRestic{
			SnapshotId:       v.Restic.SnapshotId.V,
			SnapshotTime:     v.Restic.SnapshotTime.V,
			ParentSnapshotId: v.Restic.ParentSnapshotId,
			Hostname:         v.Restic.Hostname.V,

			BackupStart: v.Restic.BackupStart.V,
			BackupEnd:   v.Restic.BackupEnd.V,

			FilesNew:            v.Restic.FilesNew.V,
			FilesChanged:        v.Restic.FilesChanged.V,
			FilesUnmodified:     v.Restic.FilesUnmodified.V,
			DirsNew:             v.Restic.DirsNew.V,
			DirsChanged:         v.Restic.DirsChanged.V,
			DirsUnmodified:      v.Restic.DirsUnmodified.V,
			DataBlobs:           v.Restic.DataBlobs.V,
			TreeBlobs:           v.Restic.TreeBlobs.V,
			DataAdded:           v.Restic.DataAdded.V,
			DataAddedPacked:     v.Restic.DataAddedPacked.V,
			TotalFilesProcessed: v.Restic.TotalFilesProcessed.V,
			TotalBytesProcessed: v.Restic.TotalBytesProcessed.V,
		}
	}

	return ret
}
