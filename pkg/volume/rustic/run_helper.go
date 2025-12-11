package rustic

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/pelletier/go-toml/v2"
)

func RunRusticCommandJson[T any](ctx context.Context, config RusticConfig, args []string) (*T, error) {
	stdout, err := RunRusticCommand(ctx, config, true, args)
	if err != nil {
		return nil, err
	}

	var ret T
	err = json.Unmarshal(stdout, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func RunRusticCommand(ctx context.Context, config RusticConfig, catchStdout bool, args []string) ([]byte, error) {
	var stdout []byte
	err := RunWithRusticConfig(config, func(configDir string) error {
		c := command_helper.CommandHelper{
			Command:     "rustic",
			Args:        args,
			Dir:         configDir,
			CatchStdout: catchStdout,
			LogCmd:      true,
		}
		err := c.Run(ctx)
		if err != nil {
			return err
		}
		stdout = c.Stdout
		return nil
	})
	if err != nil {
		return nil, err
	}

	return stdout, err
}

func RunWithRusticConfig(config RusticConfig, fn func(configDir string) error) error {
	configDir, err := BuildRusticConfigDir(config)
	if err != nil {
		return err
	}
	defer os.RemoveAll(configDir)

	return fn(configDir)
}

func BuildRusticConfigDir(config RusticConfig) (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	doRm := true
	defer func() {
		if doRm {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	configBytes, err := toml.Marshal(config)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filepath.Join(tmpDir, "rustic.toml"), configBytes, 0600)
	if err != nil {
		return "", err
	}

	doRm = false
	return tmpDir, nil
}
