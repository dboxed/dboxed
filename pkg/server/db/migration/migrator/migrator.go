package migrator

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *querier.ReadWriteDB, migrations map[string]fs.FS) error {
	fs, ok := migrations[db.DriverName()]
	if !ok {
		return fmt.Errorf("migrations for %s not found", db.DriverName())
	}

	goose.SetBaseFS(fs)
	err := goose.SetDialect(db.DriverName())
	if err != nil {
		return err
	}

	err = goose.UpContext(ctx, db.GetWriteDB().DB, ".")
	if err != nil {
		return err
	}
	return nil
}
