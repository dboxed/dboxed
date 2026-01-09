//go:build docker_volume_plugin

package cli

import (
	docker_volume_plugin "github.com/dboxed/dboxed/cmd/dboxed/commands/docker-volume-plugin"
)

type cliOnlyDockerVolumePlugin struct {
	DockerVolumePlugin docker_volume_plugin.DockerVolumePluginCommands `cmd:"" help:"Sub commands to control docker volume plugin"`
}
