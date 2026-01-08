package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"k8s.io/client-go/util/homedir"
)

func GetUserCacheDir(ctx context.Context) string {
	p := os.Getenv("DBOXED_CACHE_DIR")
	if p != "" {
		return p
	}
	p = os.Getenv("XDG_CACHE_HOME")
	if p != "" {
		return filepath.Join(p, "dboxed")
	}
	h := homedir.HomeDir()
	if h == "" || h == "/" || runtime.GOOS == "windows" {
		tmpDir := filepath.Join(os.TempDir(), "dboxed-cache")
		err := os.MkdirAll(tmpDir, 0700)
		if err != nil {
			panic(fmt.Sprintf("failed to create dboxed-cache dir in temp dir: %s", err.Error()))
		}
		return filepath.Join(tmpDir, "dboxed-cache")
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(h, ".cache", "dboxed")
	case "darwin":
		return filepath.Join(h, "Library", "Caches", "dboxed")
	default:
		panic("unsupported os in GetUserCacheDir")
	}
}
