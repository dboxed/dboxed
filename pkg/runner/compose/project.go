package compose

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/util"
)

type ComposeHelper struct {
	BaseDir      string
	NameOverride *string
	Project      *ctypes.Project
}

func (rn *ComposeHelper) writeComposeFile() (string, error) {
	b, err := rn.Project.MarshalYAML()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(rn.BaseDir, rn.projectName())
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}

	p := filepath.Join(dir, "docker-compose.yaml")

	err = util.AtomicWriteFile(p, b, 0600)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (rn *ComposeHelper) projectName() string {
	if rn.NameOverride != nil {
		return *rn.NameOverride
	}
	return rn.Project.Name
}

func (rn *ComposeHelper) RunPull(ctx context.Context) error {
	dir, err := rn.writeComposeFile()
	if err != nil {
		return err
	}

	err = RunComposeCli(ctx, nil, dir, rn.projectName(), nil, "pull")
	if err != nil {
		return err
	}
	return nil
}

func (rn *ComposeHelper) RunBuild(ctx context.Context) error {
	dir, err := rn.writeComposeFile()
	if err != nil {
		return err
	}

	err = RunComposeCli(ctx, nil, dir, rn.projectName(), nil, "build")
	if err != nil {
		return err
	}
	return nil
}

func (rn *ComposeHelper) RunUp(ctx context.Context) error {
	dir, err := rn.writeComposeFile()
	if err != nil {
		return err
	}

	err = RunComposeCli(ctx, nil, dir, rn.projectName(), nil, "up", "-d", "--remove-orphans", "--pull=never")
	if err != nil {
		return err
	}
	return nil
}

func (rn *ComposeHelper) RunExec(ctx context.Context, serviceName string, interactive bool, args ...string) error {
	dir, err := rn.writeComposeFile()
	if err != nil {
		return err
	}

	args2 := []string{
		"exec",
	}
	if interactive {
		args2 = append(args2, "-i")
	}
	args2 = append(args2, serviceName)
	args2 = append(args2, args...)

	err = RunComposeCli(ctx, nil, dir, rn.projectName(), nil, args2...)
	if err != nil {
		return err
	}
	return nil
}

func RunComposeDown(ctx context.Context, name string, removeVolumes bool, ignoreComposeErrors bool) error {
	args := []string{
		"down", "--remove-orphans",
	}
	if removeVolumes {
		args = append(args, "-v")
	}
	err := RunComposeCli(ctx, nil, "", name, nil, args...)
	if err != nil {
		if ignoreComposeErrors {
			slog.ErrorContext(ctx, "error while calling docker compose", slog.Any("error", err))
			return nil
		} else {
			return err
		}
	}
	return nil

}
