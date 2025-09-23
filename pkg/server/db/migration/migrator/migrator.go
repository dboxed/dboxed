package migrator

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *sqlx.DB, migrations map[string]fs.FS) error {
	fs, ok := migrations[db.DriverName()]
	if !ok {
		return fmt.Errorf("migrations for %s not found", db.DriverName())
	}

	goose.SetBaseFS(fs)
	err := goose.SetDialect(db.DriverName())
	if err != nil {
		return err
	}

	err = goose.UpContext(ctx, db.DB, ".")
	if err != nil {
		return err
	}
	return nil
}
