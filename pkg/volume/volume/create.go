package volume

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/fallocate"
	"github.com/dboxed/dboxed/pkg/volume/lvm"
	"github.com/google/uuid"
)

type CreateOptions struct {
	LockId    string
	ImagePath string
	ImageSize int64
	FsSize    int64
	FsType    string
	Force     bool

	VgName string

	LvmTags []string
}

func Create(ctx context.Context, opts CreateOptions) error {
	if _, err := os.Stat(opts.ImagePath); err == nil && !opts.Force {
		return fmt.Errorf("image '%s' already exists, we won't overwrite it", opts.ImagePath)
	}

	if !slices.Contains(AllowedFsTypes, opts.FsType) {
		return fmt.Errorf("invalid fs-type, must be one of %s", strings.Join(AllowedFsTypes, ", "))
	}

	vgName := opts.VgName
	if vgName == "" {
		u, err := uuid.NewV7()
		if err != nil {
			return err
		}
		vgName = u.String()
	}

	volName := "filesystem"

	f, err := os.Create(opts.ImagePath)
	if err != nil {
		return err
	}
	err = fallocate.Fallocate(f, 0, opts.ImageSize)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(opts.ImagePath)
		return err
	}
	_ = f.Close()

	loDev, loDevHandle, err := AttachLoopDev(opts.ImagePath, opts.LockId)
	if err != nil {
		return err
	}
	defer loDevHandle.Close()

	err = lvm.PVCreate(ctx, loDev.Path())
	if err != nil {
		return err
	}

	err = lvm.VGCreate(ctx, vgName, []string{loDev.Path()}, opts.LvmTags)
	if err != nil {
		return err
	}
	defer func() {
		_ = DeactivateVolume(ctx, vgName)
	}()

	err = lvm.PVAddTags(ctx, loDev.Path(), opts.LvmTags)
	if err != nil {
		return err
	}

	var fsTags []string
	fsTags = append(fsTags, opts.LvmTags...)
	fsTags = append(fsTags, "filesystem")
	err = lvm.LVCreate(ctx, vgName, volName, opts.FsSize, fsTags)
	if err != nil {
		return err
	}

	fsDev, err := lvm.BuildDevPath(vgName, volName, false)
	if err != nil {
		return err
	}
	err = util.RunCommand(ctx, fmt.Sprintf("mkfs.%s", opts.FsType), fsDev)
	if err != nil {
		return err
	}

	return nil
}
