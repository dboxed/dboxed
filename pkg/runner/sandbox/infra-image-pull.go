package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/moby/go-archive"
)

func (rn *Sandbox) pullInfraImage(ctx context.Context) error {
	dst := rn.GetSandboxRoot()
	manifestPath := rn.getInfraImageConfig()
	err := rn.pullImage(ctx, rn.InfraImage, manifestPath, dst)
	if err != nil {
		return fmt.Errorf("failed to pull imfra image: %w", err)
	}
	return nil
}

func (rn *Sandbox) pullImage(ctx context.Context, imageRef string, configPath, rootfs string) error {
	imageCache := cache.NewFilesystemCache(filepath.Join(rn.HostWorkDir, "image-cache"))

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "pulling image", slog.Any("imageRef", ref.String()))

	platform := v1.Platform{
		Architecture: runtime.GOARCH,
		OS:           "linux",
	}
	image, err := remote.Image(ref, remote.WithContext(ctx), remote.WithPlatform(platform))
	if err != nil {
		return err
	}

	image = cache.Image(image, imageCache)

	imageConfig, err := image.RawConfigFile()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, imageConfig, 0600)
	if err != nil {
		return err
	}

	r := mutate.Extract(image)
	defer r.Close()

	err = archive.Untar(r, rootfs, &archive.TarOptions{})
	if err != nil {
		return err
	}
	return nil
}
