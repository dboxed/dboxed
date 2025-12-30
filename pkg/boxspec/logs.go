package boxspec

import "time"

type LogMetadata struct {
	BoxId     *string `json:"boxId,omitempty"`
	MachineId *string `json:"machineId,omitempty"`

	FileName string         `json:"fileName"`
	Format   string         `json:"format"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type LogsBatch struct {
	Lines []LogsLine `json:"lines"`
	Seq   int64      `json:"seq,omitempty"`
}

type LogsLine struct {
	Line string    `json:"line"`
	Time time.Time `json:"time"`
}
