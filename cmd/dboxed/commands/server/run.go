package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"reflect"
	"runtime"
	"slices"
	"sync"

	"github.com/dboxed/dboxed/pkg/reconcilers/boxes"
	"github.com/dboxed/dboxed/pkg/reconcilers/load_balancers"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers"
	"github.com/dboxed/dboxed/pkg/reconcilers/machines"
	"github.com/dboxed/dboxed/pkg/reconcilers/networks"
	"github.com/dboxed/dboxed/pkg/reconcilers/s3buckets"
	"github.com/dboxed/dboxed/pkg/reconcilers/volume_providers"
	"github.com/dboxed/dboxed/pkg/reconcilers/workspaces"
	config2 "github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/migration/migrator"
	"github.com/dboxed/dboxed/pkg/server/db/migration/postgres"
	"github.com/dboxed/dboxed/pkg/server/db/migration/sqlite"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/server"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type initRunFunc func(ctx context.Context, config config2.Config) (runFunc, error)
type runFunc func(ctx context.Context) error

type RunCmd struct {
	Config string `help:"Config file" type:"existingfile"`

	All         RunAllCmd         `cmd:"" help:"run the api server and all reconcilers"`
	Api         RunApiCmd         `cmd:"" help:"run the api server"`
	Reconcilers RunReconcilersCmd `cmd:"" help:"run all reconcilers"`

	loadedConfig config2.Config
}

var apiFuncs = []initRunFunc{
	runApiServer,
}

var reconcilerFuncs = []initRunFunc{
	runReconcilerWorkspaces,
	runReconcilerMachineProviders,
	runReconcilerNetworks,
	runReconcilerS3Buckets,
	runReconcilerVolumeProviders,
	runReconcilerLoadBalancers,
	runReconcilerBoxes,
	runReconcilerMachines,
}

var allFuncs = slices.Concat(
	apiFuncs,
	reconcilerFuncs,
)

type RunAllCmd struct {
}

type RunApiCmd struct {
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

func (cmd *RunReconcilersCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()

	return runMultiple(ctx, runCmd.loadedConfig, false,
		reconcilerFuncs...,
	)
}

func runMultiple(ctx context.Context, config config2.Config, allowMigrate bool, runs ...initRunFunc) error {
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

	for _, initRun := range runs {
		fnName := runtime.FuncForPC(reflect.ValueOf(initRun).Pointer()).Name()

		slog.InfoContext(ctx, fmt.Sprintf("starting %s", fnName))

		runFn, err := initRun(ctx, config)
		if err != nil {
			return fmt.Errorf("error in %s: %w", fnName, err)
		}

		go func() {
			err := runFn(ctx)
			if err != nil {
				slog.ErrorContext(ctx, fnName, slog.Any("error", err))

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

func openDB(ctx context.Context, config config2.Config, enableSqliteFKs bool) (*querier.ReadWriteDB, error) {
	return querier.OpenReadWriteDB(config.DB.Url, enableSqliteFKs)
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

func initDB(ctx context.Context, config config2.Config, allowMigrate bool) (*querier.ReadWriteDB, error) {
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

func runApiServer(ctx context.Context, config config2.Config) (runFunc, error) {
	s, err := server.NewDboxedServer(ctx, config)
	if err != nil {
		return nil, err
	}

	err = s.InitGin()
	if err != nil {
		return nil, err
	}

	err = s.InitHuma()
	if err != nil {
		return nil, err
	}

	err = s.InitApi(ctx)
	if err != nil {
		return nil, err
	}

	return s.ListenAndServe, nil
}

func runReconcilerWorkspaces(ctx context.Context, config config2.Config) (runFunc, error) {
	r := workspaces.NewWorkspacesReconciler()
	return r.Run, nil
}

func runReconcilerMachineProviders(ctx context.Context, config config2.Config) (runFunc, error) {
	r := machine_providers.NewMachineProvidersReconciler()
	return r.Run, nil
}

func runReconcilerNetworks(ctx context.Context, config config2.Config) (runFunc, error) {
	r := networks.NewNetworksReconciler()
	return r.Run, nil
}

func runReconcilerS3Buckets(ctx context.Context, config config2.Config) (runFunc, error) {
	r := s3buckets.NewS3BucketsReconciler()
	return r.Run, nil
}

func runReconcilerVolumeProviders(ctx context.Context, config config2.Config) (runFunc, error) {
	r := volume_providers.NewVolumeProvidersReconciler()
	return r.Run, nil
}

func runReconcilerBoxes(ctx context.Context, config config2.Config) (runFunc, error) {
	r := boxes.NewBoxesReconciler()
	return r.Run, nil
}

func runReconcilerLoadBalancers(ctx context.Context, config config2.Config) (runFunc, error) {
	r := load_balancers.NewLoadBalancersReconciler()
	return r.Run, nil
}

func runReconcilerMachines(ctx context.Context, config config2.Config) (runFunc, error) {
	r := machines.NewMachinesReconciler()
	return r.Run, nil
}
