package docker_volume_plugin

import (
	"context"
	"log/slog"

	plugin "github.com/dboxed/dboxed/pkg/docker-volume-plugin"
	plugin_config "github.com/dboxed/dboxed/pkg/docker-volume-plugin/config"
	volume_helper "github.com/docker/go-plugins-helpers/volume"
)

type initRunFunc func(ctx context.Context, config plugin_config.Config) (runFunc, error)
type runFunc func(ctx context.Context) error

type RunCmd struct {
	Config       string       `help:"Config file" type:"existingfile"`
	Plugin       RunPluginCmd `cmd:"" help:"run the docker volume plugin"`
	loadedConfig plugin_config.Config
}

type RunPluginCmd struct {
}

func (cmd *RunPluginCmd) Run(runCmd *RunCmd) error {
	ctx := context.Background()

	slog.InfoContext(ctx, "Starting plugin ...")

	driver := &plugin.Driver{}
	h := volume_helper.NewHandler(driver)

	//TODO customize GID to not always be root?
	if err := h.ServeUnix("dboxed", 0); err != nil {
		slog.ErrorContext(ctx, "plugin serve error", slog.Any("error", err))
		return err
	}
	return nil
}
