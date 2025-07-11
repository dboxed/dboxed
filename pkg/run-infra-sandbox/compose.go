package run_infra_sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/types"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func (rn *RunInfraSandbox) runComposeUp(ctx context.Context) error {
	projectPath := filepath.Join(types.UnboxedDataDir, "compose")
	configPath := filepath.Join(projectPath, "compose.yaml")

	err := os.MkdirAll(projectPath, 0700)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(rn.conf.BoxSpec.Compose)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, b, 0600)
	if err != nil {
		return err
	}

	return nil
}
