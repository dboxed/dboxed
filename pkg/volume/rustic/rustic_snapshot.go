package rustic

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/models"
)

type Snapshot struct {
	Id             string          `json:"id"`
	Parent         *string         `json:"parent,omitempty"`
	Time           time.Time       `json:"time"`
	ProgramVersion string          `json:"program_version"`
	Tree           string          `json:"tree"`
	Paths          []string        `json:"paths"`
	Hostname       string          `json:"hostname"`
	Username       string          `json:"username"`
	Uid            int             `json:"uid"`
	Gid            int             `json:"gid"`
	Tags           []string        `json:"tags"`
	Original       string          `json:"original"`
	Summary        SnapshotSummary `json:"summary"`
}

type SnapshotSummary struct {
	FilesNew              int       `json:"files_new"`
	FilesChanged          int       `json:"files_changed"`
	FilesUnmodified       int       `json:"files_unmodified"`
	TotalFilesProcessed   int       `json:"total_files_processed"`
	TotalBytesProcessed   int       `json:"total_bytes_processed"`
	DirsNew               int       `json:"dirs_new"`
	DirsChanged           int       `json:"dirs_changed"`
	DirsUnmodified        int       `json:"dirs_unmodified"`
	TotalDirsProcessed    int       `json:"total_dirs_processed"`
	TotalDirsizeProcessed int       `json:"total_dirsize_processed"`
	DataBlobs             int       `json:"data_blobs"`
	TreeBlobs             int       `json:"tree_blobs"`
	DataAdded             int       `json:"data_added"`
	DataAddedPacked       int       `json:"data_added_packed"`
	DataAddedFiles        int       `json:"data_added_files"`
	DataAddedFilesPacked  int       `json:"data_added_files_packed"`
	DataAddedTrees        int       `json:"data_added_trees"`
	DataAddedTreesPacked  int       `json:"data_added_trees_packed"`
	Command               string    `json:"command"`
	BackupStart           time.Time `json:"backup_start"`
	BackupEnd             time.Time `json:"backup_end"`
	BackupDuration        float64   `json:"backup_duration"`
	TotalDuration         float64   `json:"total_duration"`
}

func (rs *Snapshot) ToApi() models.VolumeSnapshotRustic {
	return models.VolumeSnapshotRustic{
		SnapshotId:       rs.Id,
		SnapshotTime:     rs.Time,
		ParentSnapshotId: rs.Parent,
		Hostname:         rs.Hostname,

		FilesNew:              rs.Summary.FilesNew,
		FilesChanged:          rs.Summary.FilesChanged,
		FilesUnmodified:       rs.Summary.FilesUnmodified,
		TotalFilesProcessed:   rs.Summary.TotalFilesProcessed,
		TotalBytesProcessed:   rs.Summary.TotalBytesProcessed,
		DirsNew:               rs.Summary.DirsNew,
		DirsChanged:           rs.Summary.DirsChanged,
		DirsUnmodified:        rs.Summary.DirsUnmodified,
		TotalDirsProcessed:    rs.Summary.TotalDirsProcessed,
		TotalDirsizeProcessed: rs.Summary.TotalDirsizeProcessed,
		DataBlobs:             rs.Summary.DataBlobs,
		TreeBlobs:             rs.Summary.TreeBlobs,
		DataAdded:             rs.Summary.DataAdded,
		DataAddedPacked:       rs.Summary.DataAddedPacked,
		DataAddedFiles:        rs.Summary.DataAddedFiles,
		DataAddedFilesPacked:  rs.Summary.DataAddedFilesPacked,
		DataAddedTrees:        rs.Summary.DataAddedTrees,
		DataAddedTreesPacked:  rs.Summary.DataAddedTreesPacked,
		BackupStart:           rs.Summary.BackupStart,
		BackupEnd:             rs.Summary.BackupEnd,
		BackupDuration:        float32(rs.Summary.BackupDuration),
		TotalDuration:         float32(rs.Summary.TotalDuration),
	}
}
