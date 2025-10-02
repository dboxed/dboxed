package querier

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ReadWriteDB struct {
	writeDB *sqlx.DB
	readDB  *sqlx.DB
}

func OpenReadWriteDB(connUrl string, enableSqliteFKs bool) (*ReadWriteDB, error) {
	purl, err := url.Parse(connUrl)
	if err != nil {
		return nil, err
	}

	var writeDb *sqlx.DB
	var readDb *sqlx.DB
	if purl.Scheme == "sqlite3" {
		q := purl.Query()
		if enableSqliteFKs {
			if !q.Has("_foreign_keys") {
				q.Set("_foreign_keys", "on")
			}
		}

		q.Set("_journal_mode", "WAL")
		q.Set("_txlock", "immediate")
		q.Set("_busy_timeout", "5000")
		q.Set("_synchronous", "NORMAL")
		q.Set("_cache_size", "1000000000")

		purl.RawQuery = q.Encode()

		dbfile := strings.Replace(purl.String(), "sqlite3://", "", 1)

		readDb, err = sqlx.Open("sqlite3", dbfile)
		if err != nil {
			return nil, err
		}
		writeDb, err = sqlx.Open("sqlite3", dbfile)
		if err != nil {
			return nil, err
		}
		writeDb.SetMaxOpenConns(1)
		readDb.SetMaxOpenConns(max(4, runtime.NumCPU()))
	} else if purl.Scheme == "postgresql" {
		writeDb, err = sqlx.Open("pgx", purl.String())
		if err != nil {
			return nil, err
		}
		readDb = writeDb
		readDb.SetMaxOpenConns(max(4, runtime.NumCPU()))
	} else {
		return nil, fmt.Errorf("unsupported db url: %s", connUrl)
	}

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
