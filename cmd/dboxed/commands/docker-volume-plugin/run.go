package docker_volume_plugin

import (
	"context"
	"log"

	plugin "github.com/dboxed/dboxed/pkg/docker-volume-plugin"
	plugin_config "github.com/dboxed/dboxed/pkg/docker-volume-plugin/config"
	volume_helper "github.com/docker/go-plugins-helpers/volume"
	// "github.com/dboxed/dboxed/pkg/docker-volume-plugin/server"
	// plugin_config "github.com/dboxed/dboxed/pkg/server/config"
)

type initRunFunc func(ctx context.Context, config plugin_config.Config) (runFunc, error)
type runFunc func(ctx context.Context) error

type RunCmd struct {
	Config       string       `help:"Config file" type:"existingfile"`
	Plugin       RunPluginCmd `cmd:"" help:"run the docker volume plugin"`
	loadedConfig plugin_config.Config
}

// var pluginFuncs = []initRunFunc{
// 	runPlugin,
// }

type RunPluginCmd struct {
}

func (cmd *RunCmd) AfterApply() error {
	// config, err := plugin_config.LoadConfig(cmd.Config)
	// if err != nil {
	// 	return err
	// }
	// cmd.loadedConfig = *config
	return nil
}

func (cmd *RunPluginCmd) Run(runCmd *RunCmd) error {
	// ctx := context.Background()

	driver := &plugin.Driver{}
	h := volume_helper.NewHandler(driver)
	log.Print("Starting plugin ...")

	//TODO customize GID?
	if err := h.ServeUnix("dboxed", 0); err != nil {
		log.Fatalf("plugin serve error: %v", err)
		return err
	}
	return nil

	// return runMultiple(ctx,
	// 	runCmd.loadedConfig,
	// 	true,
	// 	pluginFuncs...,
	// )
}

// func runPlugin(ctx context.Context, config plugin_config.Config,
// ) (runFunc, error) {

// docker_volume_plugin.main()

// s, err := server.NewPluginServer(ctx, config)
// if err != nil {
// 	return nil, err
// }

// err = s.InitGin()
// if err != nil {
// 	return nil, err
// }

// err = s.InitHuma()
// if err != nil {
// 	return nil, err
// }

// err = s.InitApi(ctx)
// if err != nil {
// 	return nil, err
// }

// return s.ListenAndServe, nil
// }

// func runMultiple(ctx context.Context,
// 	config plugin_config.Config,
// 	allowMigrate bool,
// 	runs ...initRunFunc) error {
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	var firstErr error
// 	var m sync.Mutex

//TODO: Use allowMigrate for saved configs of already created volumes?

// db, err := initDB(ctx, config, allowMigrate)
// if err != nil {
// 	return err
// }

// ctx = context.WithValue(ctx, "config", &config)
// ctx = context.WithValue(ctx, "db", db)

// 	for _, initRun := range runs {
// 		fnName := runtime.FuncForPC(reflect.ValueOf(initRun).Pointer()).Name()

// 		slog.InfoContext(ctx, fmt.Sprintf("starting %s", fnName))

// 		runFn, err := initRun(ctx, config)

// 		if err != nil {
// 			return fmt.Errorf("error in %s: %w", fnName, err)
// 		}

// 		go func() {
// 			err := runFn(ctx)
// 			if err != nil {
// 				slog.ErrorContext(ctx, fnName, slog.Any("error", err))

// 				m.Lock()
// 				if firstErr == nil {
// 					firstErr = err
// 				}
// 				m.Unlock()
// 			}
// 			cancel()
// 		}()
// 	}

// 	<-ctx.Done()
// 	return firstErr
// }
