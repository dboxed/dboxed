package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/dboxed/dboxed/pkg/nats_conn_pool"
	"github.com/dboxed/dboxed/pkg/nats_services"
	"github.com/dboxed/dboxed/pkg/reconcilers/boxes"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers"
	"github.com/dboxed/dboxed/pkg/reconcilers/machines"
	"github.com/dboxed/dboxed/pkg/reconcilers/networks"
	"github.com/dboxed/dboxed/pkg/reconcilers/volume_providers"
	"github.com/dboxed/dboxed/pkg/reconcilers/workspaces"
	config2 "github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/migration/migrator"
	"github.com/dboxed/dboxed/pkg/server/db/migration/postgres"
	"github.com/dboxed/dboxed/pkg/server/db/migration/sqlite"
	"github.com/dboxed/dboxed/pkg/server/server"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type runFunc func(ctx context.Context, config config2.Config) error

type RunCmd struct {
	Config string `help:"Config file" type:"existingfile"`

	All          RunAllCmd          `cmd:"" help:"run the api server and all reconcilers"`
	Api          RunApiCmd          `cmd:"" help:"run the api server"`
	NatsServices RunNatsServicesCmd `cmd:"" help:"run the nats services"`
	Reconcilers  RunReconcilersCmd  `cmd:"" help:"run all reconcilers"`

	loadedConfig config2.Config
}

var apiFuncs = []runFunc{
	runApiServer,
}

var natsServicesFuncs = []runFunc{
	runNatsAuthCallout,
	runNatsServices,
}

var reconcilerFuncs = []runFunc{
	runReconcilerWorkspaces,
	runReconcilerMachineProviders,
	runReconcilerNetworks,
	runReconcilerVolumeProviders,
	runReconcilerBoxes,
	runReconcilerMachines,
}

var allFuncs = slices.Concat(
	apiFuncs,
	natsServicesFuncs,
	reconcilerFuncs,
)

type RunAllCmd struct {
}

type RunApiCmd struct {
}

type RunNatsServicesCmd struct {
}

type RunReconcilersCmd struct {
}

func (cmd *RunCmd) AfterApply() error {
	config, err := config2.LoadConfig(cmd.Config)
	if err != nil {
		return err
	}
	cmd.loadedConfig = *config
	return nil
}

func (cmd *RunApiCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()
	return runMultiple(ctx, runCmd.loadedConfig, true,
		apiFuncs...,
	)
}

func (cmd *RunAllCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()

	return runMultiple(ctx, runCmd.loadedConfig, true,
		allFuncs...,
	)
}

func (cmd *RunNatsServicesCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()

	return runMultiple(ctx, runCmd.loadedConfig, false,
		natsServicesFuncs...,
	)
}

func (cmd *RunReconcilersCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()

	return runMultiple(ctx, runCmd.loadedConfig, false,
		reconcilerFuncs...,
	)
}

func runMultiple(ctx context.Context, config config2.Config, allowMigrate bool, runs ...runFunc) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var firstErr error
	var m sync.Mutex

	db, err := initDB(ctx, config, allowMigrate)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "config", &config)
	ctx = context.WithValue(ctx, "db", db)

	natsConnPool := nats_conn_pool.NewNatsConnectionPool(ctx)
	ctx = context.WithValue(ctx, "nats-conn-pool", natsConnPool)

	for _, run := range runs {
		go func() {
			err := run(ctx, config)
			if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("error in %s", reflect.ValueOf(run).Type().Name()), slog.Any("error", err))

				m.Lock()
				if firstErr == nil {
					firstErr = err
				}
				m.Unlock()
			}
			cancel()
		}()
	}

	<-ctx.Done()
	return firstErr
}

func openDB(ctx context.Context, config config2.Config, enableSqliteFKs bool) (*sqlx.DB, error) {
	purl, err := url.Parse(config.DB.Url)
	if err != nil {
		return nil, err
	}

	var sqlxDb *sqlx.DB
	if purl.Scheme == "sqlite3" {
		q := purl.Query()
		if enableSqliteFKs {
			if !q.Has("_foreign_keys") {
				q.Set("_foreign_keys", "on")
				purl.RawQuery = q.Encode()
			}
		}
		dbfile := strings.Replace(purl.String(), "sqlite3://", "", 1)

		sqlxDb, err = sqlx.Open("sqlite3", dbfile)
		if err != nil {
			return nil, err
		}
	} else if purl.Scheme == "postgresql" {
		sqlxDb, err = sqlx.Open("pgx", purl.String())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported db url: %s", config.DB.Url)
	}

	sqlxDb.SetMaxIdleConns(8)
	sqlxDb.SetMaxOpenConns(16)

	return sqlxDb, nil
}

func migrateDB(ctx context.Context, config config2.Config) error {
	db, err := openDB(ctx, config, false)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	err = migrator.Migrate(ctx, db, map[string]fs.FS{
		"pgx":     postgres.E,
		"sqlite3": sqlite.E,
	})
	if err != nil {
		return err
	}
	return nil
}

func initDB(ctx context.Context, config config2.Config, allowMigrate bool) (*sqlx.DB, error) {
	slog.InfoContext(ctx, "initializing database")

	if allowMigrate && config.DB.Migrate {
		slog.InfoContext(ctx, "migrating database")
		err := migrateDB(ctx, config)
		if err != nil {
			return nil, err
		}
	}

	db, err := openDB(ctx, config, true)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runApiServer(ctx context.Context, config config2.Config) error {
	s, err := server.NewDboxedServer(ctx, config)
	if err != nil {
		return err
	}

	err = s.InitGin()
	if err != nil {
		return err
	}

	err = s.InitHuma()
	if err != nil {
		return err
	}

	err = s.InitApi(ctx)
	if err != nil {
		return err
	}

	return s.ListenAndServe(ctx)
}

func runReconcilerWorkspaces(ctx context.Context, config config2.Config) error {
	r := workspaces.NewWorkspacesReconciler(config)
	return r.Run(ctx)
}

func runReconcilerMachineProviders(ctx context.Context, config config2.Config) error {
	r := machine_providers.NewMachineProvidersReconciler(config)
	return r.Run(ctx)
}

func runReconcilerNetworks(ctx context.Context, config config2.Config) error {
	r := networks.NewNetworksReconciler(config)
	return r.Run(ctx)
}

func runReconcilerVolumeProviders(ctx context.Context, config config2.Config) error {
	r := volume_providers.NewVolumeProvidersReconciler(config)
	return r.Run(ctx)
}

func runReconcilerBoxes(ctx context.Context, config config2.Config) error {
	r := boxes.NewBoxesReconciler(config)
	return r.Run(ctx)
}

func runReconcilerMachines(ctx context.Context, config config2.Config) error {
	r := machines.NewMachinesReconciler(config)
	return r.Run(ctx)
}

func runNatsAuthCallout(ctx context.Context, config config2.Config) error {
	r, err := nats_services.NewAuthCalloutService(config)
	if err != nil {
		return err
	}
	return r.Run(ctx)
}

func runNatsServices(ctx context.Context, config config2.Config) error {
	r, err := nats_services.NewDboxedServices(ctx, config)
	if err != nil {
		return err
	}
	return r.Run(ctx)
}
