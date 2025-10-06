package logs

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
)

type LogsPublisher struct {
	api *TailToApi
}

func (lp *LogsPublisher) Stop() {
	if lp.api != nil {
		lp.api.MultiTail.StopAndWait()
	}
}

func (lp *LogsPublisher) Start(ctx context.Context, c *baseclient.Client, boxId int64, tailDbFile string) error {
	slog.InfoContext(ctx, "initializing logs publishing to dboxed api")

	tta, err := NewTailToApi(ctx, c, tailDbFile, boxId)
	if err != nil {
		return err
	}

	lp.api = tta
	return nil
}

func (lp *LogsPublisher) PublishDboxedLogsDir(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		fileName := filepath.Base(path)
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

	if lp.api != nil {
		return lp.api.MultiTail.WatchDir(dir, "*.log", 0, buildMetadata)
	}
	return nil
}

// PublishMultilogLogsDir publishes logs from s6-log output
func (lp *LogsPublisher) PublishMultilogLogsDir(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		serviceName := filepath.Base(filepath.Dir(path))
		logFormatBytes, _ := os.ReadFile(filepath.Join(filepath.Dir(path), "log-format"))
		logFormat := strings.TrimSpace(string(logFormatBytes))
		if logFormat == "" {
			logFormat = "raw"
		}
		return boxspec.LogMetadata{
			FileName: serviceName,
			Format:   logFormat,
			Metadata: map[string]any{
				"service-name": serviceName,
			},
		}, nil
	}

	if lp.api != nil {
		return lp.api.MultiTail.WatchDir(dir, "*/current", 1, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) PublishDockerContainerLogsDir(containersDir string) error {
	err := os.MkdirAll(containersDir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		return lp.buildDockerContainerLogMetadata(containersDir, path)
	}

	if lp.api != nil {
		return lp.api.MultiTail.WatchDir(containersDir, "*/*.log", 1, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) buildDockerContainerLogMetadata(dockerDataDir string, logPath string) (boxspec.LogMetadata, error) {
	relPath, err := filepath.Rel(dockerDataDir, logPath)
	if err != nil {
		return boxspec.LogMetadata{}, err
	}
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
		FileName: relPath,
		Format:   "docker-logs",
		Metadata: map[string]any{
			"container": config,
		},
	}, nil
}
