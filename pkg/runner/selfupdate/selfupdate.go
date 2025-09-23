package selfupdate

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
	util2 "github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/sys/unix"
)

func SelfUpdateIfNeeded(ctx context.Context, binaryUrl, binaryHash string, workDir string) error {
	if binaryUrl == "" {
		return nil
	}

	if os.Getenv("DBOXED_SELFUPDATED") == "true" {
		slog.Info("skipping selfupdate as we're already running an updated binary")
		return nil
	}

	if binaryHash != "" {
		selfPath, err := os.Executable()
		if err != nil {
			return err
		}
		selfBytes, err := os.ReadFile(selfPath)
		if err != nil {
			return err
		}
		selfHash := util.Sha256Sum(selfBytes)

		if selfHash == binaryHash {
			return nil
		}
	}

	slog.Info("updating self")

	dir := filepath.Join(workDir, "selfupdate")
	pth, err := util2.DownloadFile(ctx, binaryUrl, binaryHash, dir, util2.CompressionGzip)
	if err != nil {
		return err
	}
	err = os.Chmod(pth, 0777)
	if err != nil {
		return err
	}

	slog.Info("exec into selfupdated binary")

	env := os.Environ()
	env = append(env, "DBOXED_SELFUPDATED=true")
	err = unix.Exec(pth, os.Args, env)
	if err != nil {
		return err
	}

	return nil
}
