package querier

type HasId interface {
	GetId() int64
}

type HasTableName interface {
	GetTableName() string
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc           = "DESC"
)

type SortField struct {
	Field     string
	Direction SortOrder
}
