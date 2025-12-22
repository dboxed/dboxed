package docker_volume_plugin

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	// config2 "github.com/dboxed/dboxed/pkg/server/config"
)

type initRunFunc func(ctx context.Context,

// config config2.Config
) (runFunc, error)
type runFunc func(ctx context.Context) error

type RunCmd struct {
	Config string       `help:"Config file" type:"existingfile"`
	Plugin RunPluginCmd `cmd:"" help:"run the docker volume plugin"`
	// TODO use custom config instead of borrowing from server
	// loadedConfig config2.Config
}

var pluginFuncs = []initRunFunc{
	runPlugin,
}

type RunPluginCmd struct {
}

func (cmd *RunCmd) AfterApply() error {
	// config, err := config2.LoadConfig(cmd.Config)
	// if err != nil {
	// 	return err
	// }
	// cmd.loadedConfig = *config
	return nil
}

func (cmd *RunPluginCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()
	return runMultiple(ctx,
		// runCmd.loadedConfig,
		true,
		pluginFuncs...,
	)
}

func runPlugin(ctx context.Context,

// config config2.Config
) (runFunc, error) {
	return nil, nil
}

func runMultiple(ctx context.Context,
	// config config2.Config,
	allowMigrate bool,
	runs ...initRunFunc) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var firstErr error
	var m sync.Mutex

	// db, err := initDB(ctx, config, allowMigrate)
	// if err != nil {
	// 	return err
	// }

	// ctx = context.WithValue(ctx, "config", &config)
	// ctx = context.WithValue(ctx, "db", db)

	for _, initRun := range runs {
		fnName := runtime.FuncForPC(reflect.ValueOf(initRun).Pointer()).Name()

		slog.InfoContext(ctx, fmt.Sprintf("starting %s", fnName))

		runFn, err := initRun(ctx) // config

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
