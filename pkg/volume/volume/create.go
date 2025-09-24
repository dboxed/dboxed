package volume

import (
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

func Create(opts CreateOptions) error {
	if _, err := os.Stat(opts.ImagePath); err == nil && !opts.Force {
		return fmt.Errorf("image '%s' already exists, we won't overwrite it", opts.ImagePath)
	}

	if !slices.Contains(AllowedFsTypes, opts.FsType) {
		return fmt.Errorf("invalid fs-type, must be one of %s", strings.Join(AllowedFsTypes, ", "))
	}

	vgName := opts.VgName
	if vgName == "" {
		vgName = uuid.NewString()
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

	err = lvm.PVCreate(loDev.Path())
	if err != nil {
		return err
	}

	err = lvm.VGCreate(vgName, []string{loDev.Path()}, opts.LvmTags)
	if err != nil {
		return err
	}
	defer func() {
		_ = DeactivateVolume(vgName)
	}()

	err = lvm.PVAddTags(loDev.Path(), opts.LvmTags)
	if err != nil {
		return err
	}

	var fsTags []string
	fsTags = append(fsTags, opts.LvmTags...)
	fsTags = append(fsTags, "filesystem")
	err = lvm.LVCreate(vgName, volName, opts.FsSize, fsTags)
	if err != nil {
		return err
	}

	fsDev, err := buildDevPath(vgName, volName, false)
	if err != nil {
		return err
	}
	err = util.RunCommand(fmt.Sprintf("mkfs.%s", opts.FsType), fsDev)
	if err != nil {
		return err
	}

	return nil
}
