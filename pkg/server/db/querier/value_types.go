package querier

import (
	"database/sql"
	"database/sql/driver"
)

// we use this to indicate that it can't be null actually but we have to make it nullable
// because it's used in joins
type NullForJoin[T any] sql.Null[T]

func (n *NullForJoin[T]) Scan(value any) error {
	return (*sql.Null[T])(n).Scan(value)
}
func (n NullForJoin[T]) Value() (driver.Value, error) {
	return (*sql.Null[T])(&n).Value()
}

func N[T any](v T) NullForJoin[T] {
	return NullForJoin[T]{
		V:     v,
		Valid: true,
	}
}

type IsOmitIfNull interface {
	isOmitIfNullValid() bool
}

type OmitIfNullT[T any] sql.Null[T]

func (n OmitIfNullT[T]) Value() (driver.Value, error) {
	return driver.Value(n.V), nil
}

func (n OmitIfNullT[T]) isOmitIfNullValid() bool {
	return n.Valid
}

func OmitIfNull[T any](v *T) OmitIfNullT[T] {
	var ret OmitIfNullT[T]
	if v != nil {
		ret.V = *v
		ret.Valid = true
	}
	return ret
}

type RawSqlT struct {
	SQL string
}

func RawSql(sql string) RawSqlT {
	return RawSqlT{
		SQL: sql,
	}
}

func ExcludeNonNull(b bool) any {
	if b {
		return RawSql("is null")
	} else {
		return OmitIfNull[any](nil)
	}
}
