package models

import (
	"encoding/json"
	"time"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type LogMetadataModel struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	boxspec.LogMetadata
}

type PostLogs struct {
	Metadata boxspec.LogMetadata `json:"metadata"`
	Lines    []boxspec.LogsLine  `json:"lines"`
}

type LogsError struct {
	Message string `json:"message"`
}

func LogMetadataFromDB(s dmodel.LogMetadata) (*LogMetadataModel, error) {
	var m map[string]any
	err := json.Unmarshal([]byte(s.Metadata), &m)
	if err != nil {
		return nil, err
	}
	return &LogMetadataModel{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,
		LogMetadata: boxspec.LogMetadata{
			FileName: s.FileName,
			Format:   s.Format,
			Metadata: m,
		},
	}, nil
}

func LogLineFromDB(s dmodel.LogLine) boxspec.LogsLine {
	return boxspec.LogsLine{
		Time: s.Time,
		Line: s.Line,
	}
}
