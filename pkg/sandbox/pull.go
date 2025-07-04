package sandbox

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/moby/go-archive"
	"log/slog"
	"os"
	"path/filepath"
)

func (rn *Sandbox) pullImages(ctx context.Context) error {
	err := rn.forAllContainers(func(c *types.ContainerSpec) error {
		manifestPath := rn.getContainerImageConfig(c.Name)
		dst := rn.getContainerRoot(c.Name)
		err := rn.pullImage(ctx, c.Image, manifestPath, dst)
		if err != nil {
			return fmt.Errorf("failed to pull image for container %s: %w", c.Name, err)
		}
		return nil
	})
	if err != nil {
		return err
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

	image, err := remote.Image(ref, remote.WithContext(ctx))
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
