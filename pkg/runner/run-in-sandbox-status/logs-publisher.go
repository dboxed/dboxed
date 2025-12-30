package run_in_sandbox_status

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/logs/multitail"
)

type LogsPublisher struct {
	Client *baseclient.Client
	BoxId  string

	mt *multitail.MultiTail
}

func (lp *LogsPublisher) Stop(cancelAfter *time.Duration) {
	if lp == nil {
		return
	}
	if lp.mt != nil {
		lp.mt.StopAndWait(cancelAfter)
	}
}

func (lp *LogsPublisher) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "initializing logs publishing to dboxed api")

	tta, err := logs.NewTailToApi(ctx, lp.Client, filepath.Join(consts.LogsDir, consts.LogsTailDbFilename))
	if err != nil {
		return err
	}

	lp.mt = tta.MultiTail

	err = lp.publishDboxedLogsDir(consts.LogsDir)
	if err != nil {
		return err
	}

	err = lp.publishDockerContainerLogsDir("/var/lib/docker/containers")
	if err != nil {
		return err
	}
	return nil
}

func (lp *LogsPublisher) publishDboxedLogsDir(dir string) error {
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
			BoxId:    &lp.BoxId,
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

func (lp *LogsPublisher) publishDockerContainerLogsDir(containersDir string) error {
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
		BoxId:    &lp.BoxId,
		FileName: filepath.Join("containers", config.Name, config.ID),
		Format:   "docker-logs",
		Metadata: map[string]any{
			"container": config,
		},
	}, nil
}
