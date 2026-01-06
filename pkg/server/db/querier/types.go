package querier

type HasId interface {
	GetId() string
}

type HasTableName interface {
	GetTableName() string
}

type HasOmits interface {
	GetOmittedColumns() []string
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

func SortBySingleField(fieldName string, dir SortOrder) []SortField {
	return []SortField{
		{
			Field:     fieldName,
			Direction: dir,
		},
	}
}
