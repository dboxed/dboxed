package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Ptr[T any](v T) *T {
	return &v
}

func SleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

func AtomicWriteFile(path string, b []byte, perm os.FileMode) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(b)
	if err != nil {
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	err = os.Chmod(tmpFile.Name(), perm)
	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), path)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	return nil
}
