package logs

import (
	"context"
	"encoding/json"
	"github.com/dboxed/dboxed/pkg/logs/multitail"
	"github.com/dboxed/dboxed/pkg/types"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type LogsPublisher struct {
	nats *TailToNats
}

func (lp *LogsPublisher) Start(ctx context.Context, boxSpec types.BoxSpec, tailDbFile string) error {

	if boxSpec.Logs == nil {
		return nil
	}

	if boxSpec.Logs.Nats != nil {
		return lp.startLogsPublishingNats(ctx, boxSpec, tailDbFile)
	}
	return nil
}

func (lp *LogsPublisher) Stop() {
	if lp.nats != nil {
		lp.nats.MultiTail.StopAndWait()
	}
}

func (lp *LogsPublisher) startLogsPublishingNats(ctx context.Context, boxSpec types.BoxSpec, tailDbFile string) error {
	slog.InfoContext(ctx, "initializing logs publishing to nats",
		slog.Any("natsUrl", boxSpec.Logs.Nats.Url),
		slog.Any("metadataKVStore", boxSpec.Logs.Nats.MetadataKVStore),
		slog.Any("logStream", boxSpec.Logs.Nats.LogStream),
		slog.Any("logId", boxSpec.Logs.Nats.LogId),
	)

	ttn, err := NewTailToNats(ctx, boxSpec.Logs.Nats.Url, boxSpec.Logs.Nats.NKeySeed, tailDbFile, boxSpec.Logs.Nats.MetadataKVStore, boxSpec.Logs.Nats.LogStream, boxSpec.Logs.Nats.LogId)
	if err != nil {
		return err
	}

	lp.nats = ttn
	return nil
}

func (lp *LogsPublisher) PublishDboxedLogsDir(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (multitail.LogMetadata, error) {
		fileName := filepath.Base(path)
		format := "slog-json"
		if strings.HasSuffix(fileName, ".stdout.log") {
			format = "raw"
		}
		return multitail.LogMetadata{
			FileName: fileName,
			Format:   format,
			Metadata: map[string]any{},
		}, nil
	}

	if lp.nats != nil {
		return lp.nats.MultiTail.WatchDir(dir, "*.log", 0, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) PublishDockerLogsDir(dockerDataDir string) error {
	err := os.MkdirAll(dockerDataDir, 0700)
	if err != nil {
		return err
	}

	watchDir := filepath.Join(dockerDataDir, "containers")
	err = os.MkdirAll(watchDir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (multitail.LogMetadata, error) {
		return lp.buildDockerContainerLogMetadata(dockerDataDir, path)
	}

	if lp.nats != nil {
		return lp.nats.MultiTail.WatchDir(watchDir, "*/*.log", 1, buildMetadata)
	}
	return nil
}

func (lp *LogsPublisher) buildDockerContainerLogMetadata(dockerDataDir string, logPath string) (multitail.LogMetadata, error) {
	relPath, err := filepath.Rel(dockerDataDir, logPath)
	if err != nil {
		return multitail.LogMetadata{}, err
	}
	configPath := filepath.Join(filepath.Dir(logPath), "config.v2.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return multitail.LogMetadata{}, err
	}

	var config types.DockerContainerConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return multitail.LogMetadata{}, err
	}

	return multitail.LogMetadata{
		FileName: relPath,
		Format:   "docker-logs",
		Metadata: map[string]any{
			"container": config,
		},
	}, nil
}
