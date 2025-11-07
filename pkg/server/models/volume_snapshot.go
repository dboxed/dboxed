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

	Rustic *VolumeSnapshotRustic `json:"rustic,omitempty"`
}

type VolumeSnapshotRustic struct {
	SnapshotId       string    `json:"snapshotId"`
	SnapshotTime     time.Time `json:"snapshotTime"`
	ParentSnapshotId *string   `json:"parentSnapshotId,omitempty"`

	Hostname string `json:"hostname"`

	FilesNew            int `json:"filesNew"`
	FilesChanged        int `json:"filesChanged"`
	FilesUnmodified     int `json:"filesUnmodified"`
	TotalFilesProcessed int `json:"totalFilesProcessed"`
	TotalBytesProcessed int `json:"totalBytesProcessed"`

	DirsNew               int `json:"dirsNew"`
	DirsChanged           int `json:"dirsChanged"`
	DirsUnmodified        int `json:"dirsUnmodified"`
	TotalDirsProcessed    int `json:"totalDirsProcessed"`
	TotalDirsizeProcessed int `json:"totalDirsizeProcessed"`
	DataBlobs             int `json:"dataBlobs"`
	TreeBlobs             int `json:"treeBlobs"`
	DataAdded             int `json:"dataAdded"`
	DataAddedPacked       int `json:"dataAddedPacked"`
	DataAddedFiles        int `json:"dataAddedFiles"`
	DataAddedFilesPacked  int `json:"dataAddedFilesPacked"`
	DataAddedTrees        int `json:"dataAddedTrees"`
	DataAddedTreesPacked  int `json:"dataAddedTreesPacked"`

	BackupStart    time.Time `json:"backupStart"`
	BackupEnd      time.Time `json:"backupEnd"`
	BackupDuration float32   `json:"backupDuration"`
	TotalDuration  float32   `json:"totalDuration"`
}

type CreateVolumeSnapshot struct {
	MountId string `json:"mountId"`

	Rustic *VolumeSnapshotRustic `json:"rustic,omitempty"`
}

func VolumeSnapshotFromDB(v dmodel.VolumeSnapshot) VolumeSnapshot {
	ret := VolumeSnapshot{
		ID:        v.ID,
		Workspace: v.WorkspaceID,
		CreatedAt: v.CreatedAt,
		VolumeID:  v.VolumedID.V,
		MountId:   v.MountID.V,
	}

	if v.Rustic != nil && v.Rustic.ID.Valid {
		ret.Rustic = &VolumeSnapshotRustic{
			SnapshotId:            v.Rustic.SnapshotId.V,
			SnapshotTime:          v.Rustic.SnapshotTime.V,
			ParentSnapshotId:      v.Rustic.ParentSnapshotId,
			Hostname:              v.Rustic.Hostname.V,
			FilesNew:              v.Rustic.FilesNew.V,
			FilesChanged:          v.Rustic.FilesChanged.V,
			FilesUnmodified:       v.Rustic.FilesUnmodified.V,
			TotalFilesProcessed:   v.Rustic.TotalFilesProcessed.V,
			TotalBytesProcessed:   v.Rustic.TotalBytesProcessed.V,
			DirsNew:               v.Rustic.DirsNew.V,
			DirsChanged:           v.Rustic.DirsChanged.V,
			DirsUnmodified:        v.Rustic.DirsUnmodified.V,
			TotalDirsProcessed:    v.Rustic.TotalDirsProcessed.V,
			TotalDirsizeProcessed: v.Rustic.TotalDirsizeProcessed.V,
			DataBlobs:             v.Rustic.DataBlobs.V,
			TreeBlobs:             v.Rustic.TreeBlobs.V,
			DataAdded:             v.Rustic.DataAdded.V,
			DataAddedPacked:       v.Rustic.DataAddedPacked.V,
			DataAddedFiles:        v.Rustic.DataAddedFiles.V,
			DataAddedFilesPacked:  v.Rustic.DataAddedFilesPacked.V,
			DataAddedTrees:        v.Rustic.DataAddedTrees.V,
			DataAddedTreesPacked:  v.Rustic.DataAddedTreesPacked.V,
			BackupStart:           v.Rustic.BackupStart.V,
			BackupEnd:             v.Rustic.BackupEnd.V,
			BackupDuration:        v.Rustic.BackupDuration.V,
			TotalDuration:         v.Rustic.TotalDuration.V,
		}
	}

	return ret
}
