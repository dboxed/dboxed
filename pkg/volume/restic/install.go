package restic

import (
	"context"
	"fmt"
	"runtime"

	"github.com/dboxed/dboxed/pkg/util/download_helper"
)

const defaultResticVersion = "0.18.1"

func DownloadResticBinary(ctx context.Context) (string, error) {
	version := defaultResticVersion
	return DownloadResticBinaryVersion(ctx, version)
}

func DownloadResticBinaryVersion(ctx context.Context, version string) (string, error) {
	url := fmt.Sprintf("https://github.com/restic/restic/releases/download/v%s/restic_%s_%s_%s.bz2", version, version, runtime.GOOS, runtime.GOARCH)
	dir, err := download_helper.DownloadBinary(ctx, download_helper.Opts{
		Url:         url,
		DownloadKey: fmt.Sprintf("restic-%s", version),
		BinName:     "restic",
		Bz2:         true,
	})

	if err != nil {
		return "", err
	}

	return dir, nil
}
