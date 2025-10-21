package logs

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/logs/multitail"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type LogsPublisher struct {
	mt *multitail.MultiTail
}

func (lp *LogsPublisher) Stop(cancel bool) {
	if lp.mt != nil {
		lp.mt.StopAndWait(cancel)
	}
}

func (lp *LogsPublisher) Start(ctx context.Context, mt *multitail.MultiTail) error {
	slog.InfoContext(ctx, "initializing logs publishing to dboxed api")
	lp.mt = mt
	return nil
}

func (lp *LogsPublisher) PublishDboxedLogsDir(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		fileName := filepath.Join("dboxed", filepath.Base(path))
		format := "slog-json"
		if strings.HasSuffix(fileName, ".stdout.log") {
			format = "raw"
		}
		return boxspec.LogMetadata{
			FileName: fileName,
			Format:   format,
			Metadata: map[string]any{},
		}, nil
	}

	if lp.mt != nil {
		return lp.mt.WatchDir(dir, "*.log", 0, buildMetadata)
	}
	return nil
}

// PublishS6Logs publishes logs from s6-log output
func (lp *LogsPublisher) PublishS6Logs(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		logDir := filepath.Dir(path)
		serviceName := filepath.Base(logDir)
		logFormatBytes, _ := os.ReadFile(filepath.Join(logDir, "log-format"))
		logFormat := strings.TrimSpace(string(logFormatBytes))
		if logFormat == "" {
			logFormat = "raw"
		}
		return boxspec.LogMetadata{
			FileName: filepath.Join("dboxed", serviceName),
			Format:   logFormat,
			Metadata: map[string]any{
				"service-name": serviceName,
			},
		}, nil
	}

	if lp.mt != nil {
		return lp.mt.WatchDir(dir, "*/current", 1, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) PublishVolumeServiceLogs(volumesDir string) error {
	err := os.MkdirAll(volumesDir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		logDir := filepath.Dir(path)
		volumeDir := filepath.Dir(logDir)

		var fileName string
		metadata := map[string]any{}
		volumeState, err := volume_serve.LoadVolumeState(volumeDir)
		if err == nil {
			metadata["volume"] = map[string]any{
				"name":       volumeState.Volume.Name,
				"id":         volumeState.Volume.ID,
				"uuid":       volumeState.Volume.Uuid,
				"mount-name": volumeState.MountName,
			}
			fileName = filepath.Join("volumes", volumeState.Volume.Name)
		} else {
			fileName = filepath.Join("volumes", filepath.Base(volumeDir))
		}
		return boxspec.LogMetadata{
			FileName: fileName,
			Format:   "slog-json",
			Metadata: metadata,
		}, nil
	}

	if lp.mt != nil {
		return lp.mt.WatchDir(volumesDir, "*/logs/current", 2, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) PublishDockerContainerLogsDir(containersDir string) error {
	err := os.MkdirAll(containersDir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		return lp.buildDockerContainerLogMetadata(path)
	}

	if lp.mt != nil {
		return lp.mt.WatchDir(containersDir, "*/*.log", 1, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) buildDockerContainerLogMetadata(logPath string) (boxspec.LogMetadata, error) {
	configPath := filepath.Join(filepath.Dir(logPath), "config.v2.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return boxspec.LogMetadata{}, err
	}

	var config dockercli.DockerContainerConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return boxspec.LogMetadata{}, err
	}

	return boxspec.LogMetadata{
		FileName: filepath.Join("containers", config.Name, config.ID),
		Format:   "docker-logs",
		Metadata: map[string]any{
			"container": config,
		},
	}, nil
}
