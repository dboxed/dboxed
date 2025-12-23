package util

import (
	"context"
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
	switch runtime.GOOS {
	case "linux":
		if h != "" {
			return filepath.Join(h, ".cache", "dboxed")
		}
	case "darwin":
		if h != "" {
			return filepath.Join(h, "Library", "Caches", "dboxed")
		}
	case "windows":
		break
	}
	return filepath.Join(os.TempDir(), "dboxed-cache")
}
