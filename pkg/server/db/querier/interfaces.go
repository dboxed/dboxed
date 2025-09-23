package querier

type HasId interface {
	GetId() int64
}

type HasTableName interface {
	GetTableName() string
}
