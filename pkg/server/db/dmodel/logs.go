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

	TotalLineBytes int64      `db:"total_line_bytes" omitOnConflictUpdate:"true"`
	LastLogTime    *time.Time `db:"last_log_time" omitOnConflictUpdate:"true"`
}

type LogLine struct {
	ID          int64 `db:"id" omitCreate:"true"`
	WorkspaceID int64 `db:"workspace_id"`
	LogID       int64 `db:"log_id"`

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
	}, nil)
}

func GetLogMetadataById(q *querier2.Querier, workspaceId *int64, logId int64, skipDeleted bool) (*LogMetadata, error) {
	return querier2.GetOne[LogMetadata](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           logId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListLogLinesSinceTime(q *querier2.Querier, logId int64, since time.Time, limit *int64) ([]LogLine, error) {
	var timeExpr string
	switch q.DB.DriverName() {
	case "pgx":
		timeExpr = fmt.Sprintf(">= timestamptz '%s'", since.UTC().Format(time.RFC3339))
	case "sqlite3":
		timeExpr = fmt.Sprintf(">= datetime('%s', 'utc')", since.UTC().Format(time.RFC3339))
	default:
		return nil, fmt.Errorf("unsupported db driver")
	}

	return querier2.GetManySorted[LogLine](q, map[string]any{
		"log_id": logId,
		"time":   querier2.RawSql(timeExpr),
	}, &querier2.SortAndPage{
		Sort: []querier2.SortField{
			{
				Field:     "id",
				Direction: querier2.SortOrderAsc,
			},
		},
		Limit: limit,
	})
}

func ListLogLinesSinceSeq(q *querier2.Querier, logId int64, seq int64, limit *int64) ([]LogLine, error) {
	return querier2.GetManySorted[LogLine](q, map[string]any{
		"log_id": logId,
		"id":     querier2.RawSql(fmt.Sprintf(">= %d", seq)),
	}, &querier2.SortAndPage{
		Sort: []querier2.SortField{
			{
				Field:     "id",
				Direction: querier2.SortOrderAsc,
			},
		},
		Limit: limit,
	})
}

type LogLineBytes struct {
	ID        int64 `db:"id"`
	LogID     int64 `db:"log_id"`
	LineBytes int64 `db:"log_line_bytes"`
}

func QueryLogLineBytes(q *querier2.Querier, workspaceId int64, limit int64) ([]LogLineBytes, error) {
	query := fmt.Sprintf("select id, log_id, octet_length(line) as log_line_bytes from log_line where workspace_id = :workspace_id order by id asc limit :limit")

	var ret []LogLineBytes
	err := q.SelectNamed(&ret, query, map[string]any{
		"workspace_id": workspaceId,
		"limit":        limit,
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type WorkspaceLogBytesUsage struct {
	WorkspaceID  int64 `db:"workspace_id"`
	SumLineBytes int64 `db:"sum_line_bytes"`
}

func (v *WorkspaceLogBytesUsage) GetTableName() string {
	return "log_metadata"
}

func QueryWorkspaceLogBytesUsage(q *querier2.Querier, workspaceId int64) (*WorkspaceLogBytesUsage, error) {
	query := fmt.Sprintf("select workspace_id, coalesce(sum(total_line_bytes), 0) as sum_line_bytes from log_metadata where workspace_id = :workspace_id group by workspace_id")

	var ret WorkspaceLogBytesUsage
	err := q.GetNamed(&ret, query, map[string]any{
		"workspace_id": workspaceId,
	})
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func DeleteLogLinesUntilId(q *querier2.Querier, workspaceId int64, untilLogLineId int64) (int64, error) {
	query := fmt.Sprintf("delete from log_line where workspace_id = :workspace_id and id <= :uid")

	qr, err := q.ExecNamed(query, map[string]any{
		"workspace_id": workspaceId,
		"uid":          untilLogLineId,
	})
	if err != nil {
		return 0, err
	}
	la, err := qr.RowsAffected()
	if err != nil {
		return 0, err
	}
	return la, nil
}

func AddLogMetadataTotalBytes(q *querier2.Querier, logId int64, add int64) error {
	query := fmt.Sprintf("update log_metadata set total_line_bytes = total_line_bytes + :add where id = :log_id")

	return q.ExecOneNamed(query, map[string]any{
		"log_id": logId,
		"add":    add,
	})
}

func (v *LogMetadata) UpdateLastLogTime(q *querier2.Querier, lastLogTime time.Time) error {
	v.LastLogTime = &lastLogTime
	return querier2.UpdateOneFromStruct(q, v, "last_log_time")
}
