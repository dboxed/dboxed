package dmodel

import (
	"fmt"
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type LogMetadata struct {
	OwnedByWorkspace

	BoxID int64 `db:"box_id"`

	FileName string `db:"file_name"`
	Format   string `db:"format"`
	Metadata string `db:"metadata"`
}

type LogLine struct {
	ID    int64 `db:"id" omitCreate:"true"`
	LogID int64 `db:"log_id"`

	Time time.Time `db:"time"`
	Line string    `db:"line"`
}

func (v *LogMetadata) CreateOrUpdate(q *querier2.Querier) error {
	return querier2.CreateOrUpdate(q, v, "box_id, file_name")
}

func (v *LogLine) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func ListLogMetadataForBox(q *querier2.Querier, workspaceId *int64, boxId int64, skipDeleted bool) ([]LogMetadata, error) {
	return querier2.GetMany[LogMetadata](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"box_id":       boxId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetLogMetadataById(q *querier2.Querier, workspaceId *int64, logId int64, skipDeleted bool) (*LogMetadata, error) {
	return querier2.GetOne[LogMetadata](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           logId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListLogLinesSinceTime(q *querier2.Querier, logId int64, since time.Time) ([]LogLine, error) {
	return querier2.GetManySorted[LogLine](q, map[string]any{
		"log_id": logId,
		"time":   querier2.RawSql(fmt.Sprintf(">= datetime('%s', 'utc')", since.UTC().Format(time.RFC3339))),
	}, []querier2.SortField{
		{
			Field:     "id",
			Direction: querier2.SortOrderAsc,
		},
	})
}

func ListLogLinesSinceSeq(q *querier2.Querier, logId int64, seq int64) ([]LogLine, error) {
	return querier2.GetManySorted[LogLine](q, map[string]any{
		"log_id": logId,
		"id":     querier2.RawSql(fmt.Sprintf(">= %d", seq)),
	}, []querier2.SortField{
		{
			Field:     "id",
			Direction: querier2.SortOrderAsc,
		},
	})
}
