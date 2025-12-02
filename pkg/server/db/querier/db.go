package querier

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"runtime"

	"github.com/jmoiron/sqlx"
)

type ReadWriteDB struct {
	writeDB *sqlx.DB
	readDB  *sqlx.DB
}

func OpenReadWriteDB(connUrl string) (*ReadWriteDB, error) {
	purl, err := url.Parse(connUrl)
	if err != nil {
		return nil, err
	}

	var writeDb *sqlx.DB
	var readDb *sqlx.DB
	if purl.Scheme != "postgresql" {
		return nil, fmt.Errorf("unsupported db url: %s", connUrl)
	}

	writeDb, err = sqlx.Open("pgx", purl.String())
	if err != nil {
		return nil, err
	}
	readDb = writeDb
	readDb.SetMaxOpenConns(max(4, runtime.NumCPU()))

	db := &ReadWriteDB{
		writeDB: writeDb,
		readDB:  readDb,
	}

	return db, nil
}

func (db *ReadWriteDB) DriverName() string {
	return db.readDB.DriverName()
}

func (db *ReadWriteDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.readDB.QueryContext(ctx, query, args...)
}

func (db *ReadWriteDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.readDB.QueryxContext(ctx, query, args...)
}

func (db *ReadWriteDB) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return db.readDB.QueryRowxContext(ctx, query, args...)
}

func (db *ReadWriteDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.writeDB.ExecContext(ctx, query, args...)
}

func (db *ReadWriteDB) Transaction(ctx context.Context, fn func(tx *sqlx.Tx) (bool, error)) error {
	tx, err := db.writeDB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	doRollback := true
	defer func() {
		if doRollback {
			_ = tx.Rollback()
		}
	}()

	commit, err := fn(tx)
	if err != nil {
		return err
	}
	doRollback = false
	if commit {
		err = tx.Commit()
		if err != nil {
			return err
		}
	} else {
		err = tx.Rollback()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *ReadWriteDB) Close() error {
	err := db.writeDB.Close()
	if err != nil {
		return err
	}
	if db.readDB != db.writeDB {
		err = db.readDB.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *ReadWriteDB) GetWriteDB() *sqlx.DB {
	return db.writeDB
}
