//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/klauspost/compress/gzip"
	"github.com/moby/go-archive"
)

func (rn *Sandbox) pullInfraImage(ctx context.Context) error {
	dst := rn.GetSandboxRoot()
	manifestPath := rn.getInfraImageConfig()
	err := rn.pullImage(ctx, rn.InfraImage, manifestPath, dst)
	if err != nil {
		return fmt.Errorf("failed to pull infra-image: %w", err)
	}
	return nil
}

func (rn *Sandbox) pullImage(ctx context.Context, imageRef string, configPath, rootfs string) error {
	imageCacheDir := filepath.Join(rn.HostWorkDir, "image-cache")

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return err
	}

	err = os.MkdirAll(imageCacheDir, 0755)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "retrieving image metadata", slog.Any("imageRef", ref.String()))

	platform := v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           "linux",
	}
	image, err := remote.Image(ref, remote.WithContext(ctx), remote.WithPlatform(platform))
	if err != nil {
		return err
	}

	digest, err := image.Digest()
	if err != nil {
		return err
	}

	imageConfig, err := image.RawConfigFile()
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(imageCacheDir, digest.Hex+".tar.gz")
	if _, err := os.Stat(cacheFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		r := mutate.Extract(image)
		defer r.Close()

		slog.InfoContext(ctx, "pulling image", slog.Any("imageRef", ref.String()))

		w, err := os.Create(cacheFile + ".tmp")
		if err != nil {
			return err
		}
		defer w.Close()

		w2 := gzip.NewWriter(w)
		if err != nil {
			return err
		}
		defer w2.Close()

		_, err = io.Copy(w2, r)
		if err != nil {
			return err
		}
		err = w2.Close()
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		err = r.Close()
		if err != nil {
			return err
		}
		err = os.Rename(cacheFile+".tmp", cacheFile)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, imageConfig, 0600)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "extracting image", slog.Any("imageRef", ref.String()))

	r, err := os.Open(cacheFile)
	if err != nil {
		return err
	}
	defer r.Close()

	r2, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer r2.Close()

	err = archive.Untar(r2, rootfs, &archive.TarOptions{})
	if err != nil {
		return err
	}
	return nil
}
