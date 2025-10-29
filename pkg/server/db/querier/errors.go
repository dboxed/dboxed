package querier

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mattn/go-sqlite3"
)

type SqlNotFoundError struct {
	TableName string
	Err       error
}

func (e SqlNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.TableName)
}

func IsSqlNotFoundError(err error) bool {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return true
	}

	var err2 *SqlNotFoundError
	if errors.As(err, &err2) {
		return true
	}
	return false
}

func IsSqlConstraintViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		if int(sqliteErr.Code) == int(sqlite3.ErrConstraint) {
			return true
		}
		return true
	}

	return false
}
