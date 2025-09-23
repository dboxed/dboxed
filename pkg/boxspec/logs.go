package boxspec

import "time"

type LogsBatch struct {
	Lines []LogsLine `json:"lines"`
	Seq   int64      `json:"seq,omitempty"`
}

type LogsLine struct {
	Line string    `json:"line"`
	Time time.Time `json:"time"`
}
