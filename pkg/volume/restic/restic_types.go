package restic

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/models"
)

// Snapshot is the state of a resource at one point in time.
type Snapshot struct {
	Id string `json:"id"`

	Time     time.Time `json:"time"`
	Parent   *string   `json:"parent,omitempty"`
	Tree     *string   `json:"tree"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname,omitempty"`
	Username string    `json:"username,omitempty"`
	UID      uint32    `json:"uid,omitempty"`
	GID      uint32    `json:"gid,omitempty"`
	Excludes []string  `json:"excludes,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Original *string   `json:"original,omitempty"`

	ProgramVersion string           `json:"program_version,omitempty"`
	Summary        *SnapshotSummary `json:"summary,omitempty"`
}

type SnapshotSummary struct {
	BackupStart time.Time `json:"backup_start"`
	BackupEnd   time.Time `json:"backup_end"`

	// statistics from the backup json output
	FilesNew            uint   `json:"files_new"`
	FilesChanged        uint   `json:"files_changed"`
	FilesUnmodified     uint   `json:"files_unmodified"`
	DirsNew             uint   `json:"dirs_new"`
	DirsChanged         uint   `json:"dirs_changed"`
	DirsUnmodified      uint   `json:"dirs_unmodified"`
	DataBlobs           int    `json:"data_blobs"`
	TreeBlobs           int    `json:"tree_blobs"`
	DataAdded           uint64 `json:"data_added"`
	DataAddedPacked     uint64 `json:"data_added_packed"`
	TotalFilesProcessed uint   `json:"total_files_processed"`
	TotalBytesProcessed uint64 `json:"total_bytes_processed"`
}

func (rs *Snapshot) ToApi() *models.VolumeSnapshotRestic {
	return &models.VolumeSnapshotRestic{
		SnapshotId:       rs.Id,
		SnapshotTime:     rs.Time,
		ParentSnapshotId: rs.Parent,
		Hostname:         rs.Hostname,

		BackupStart: rs.Summary.BackupStart,
		BackupEnd:   rs.Summary.BackupEnd,

		FilesNew:            int(rs.Summary.FilesNew),
		FilesChanged:        int(rs.Summary.FilesChanged),
		FilesUnmodified:     int(rs.Summary.FilesUnmodified),
		DirsNew:             int(rs.Summary.DirsNew),
		DirsChanged:         int(rs.Summary.DirsChanged),
		DirsUnmodified:      int(rs.Summary.DirsUnmodified),
		DataBlobs:           rs.Summary.DataBlobs,
		TreeBlobs:           rs.Summary.TreeBlobs,
		DataAdded:           int64(rs.Summary.DataAdded),
		DataAddedPacked:     int64(rs.Summary.DataAddedPacked),
		TotalFilesProcessed: int(rs.Summary.TotalFilesProcessed),
		TotalBytesProcessed: int64(rs.Summary.TotalBytesProcessed),
	}
}
