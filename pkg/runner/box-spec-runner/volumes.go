package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
	"sigs.k8s.io/yaml"
)

type volumeInterface interface {
	WorkDirBase(vol boxspec.BoxVolumeSpec) string
	IsReadOnly(vol boxspec.BoxVolumeSpec) bool

	Create(ctx context.Context, vol boxspec.BoxVolumeSpec) error
	Delete(ctx context.Context, vol boxspec.BoxVolumeSpec) error

	CheckRecreateNeeded(oldVol boxspec.BoxVolumeSpec, newVol boxspec.BoxVolumeSpec) bool
}

func (rn *BoxSpecRunner) buildVolumeInterface(vol boxspec.BoxVolumeSpec) (volumeInterface, error) {
	if vol.FileBundle != nil {
		return &volumeInterfaceFileBundle{}, nil
	} else if vol.Dboxed != nil {
		return &volumeInterfaceDboxed{rn: rn}, nil
	} else {
		return nil, fmt.Errorf("unknown volume type for %s", vol.Name)
	}
}

func getVolumeWorkDir(i volumeInterface, vol boxspec.BoxVolumeSpec) string {
	return filepath.Join(consts.VolumesDir, i.WorkDirBase(vol))
}

func getVolumeMountDir(i volumeInterface, vol boxspec.BoxVolumeSpec) string {
	return filepath.Join(getVolumeWorkDir(i, vol), "mount")
}

func (rn *BoxSpecRunner) buildVolumeSpecPath(vi volumeInterface, vol boxspec.BoxVolumeSpec) string {
	volumeSpecFile := filepath.Join(consts.VolumesDir, fmt.Sprintf("volume-spec-%s.yaml", vi.WorkDirBase(vol)))
	return volumeSpecFile
}

func (rn *BoxSpecRunner) readOldVolumeSpecs() ([]boxspec.BoxVolumeSpec, error) {
	des, err := os.ReadDir(consts.VolumesDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var oldVolumes []boxspec.BoxVolumeSpec
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		if m, _ := filepath.Match("volume-spec-*.yaml", de.Name()); !m {
			continue
		}

		volumeSpecFile := filepath.Join(consts.VolumesDir, de.Name())
		volSpec, err := util.UnmarshalYamlFile[boxspec.BoxVolumeSpec](volumeSpecFile)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if volSpec != nil {
			oldVolumes = append(oldVolumes, *volSpec)
		}
	}
	return oldVolumes, nil
}

func (rn *BoxSpecRunner) writeVolumeSpec(vi volumeInterface, vol boxspec.BoxVolumeSpec) error {
	b, err := yaml.Marshal(vol)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(rn.buildVolumeSpecPath(vi, vol), b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) deleteVolumeSpec(vi volumeInterface, vol boxspec.BoxVolumeSpec) error {
	err := os.Remove(rn.buildVolumeSpecPath(vi, vol))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) reconcileVolumes(ctx context.Context, newVolumes []boxspec.BoxVolumeSpec, allowDownService bool) error {
	oldVolumesByName := map[string]*boxspec.BoxVolumeSpec{}
	newVolumeByName := map[string]*boxspec.BoxVolumeSpec{}

	oldVolumes, err := rn.readOldVolumeSpecs()
	if err != nil {
		return err
	}

	for _, v := range oldVolumes {
		_, err := rn.buildVolumeInterface(v)
		if err != nil {
			return err
		}
		oldVolumesByName[v.Name] = &v
	}
	for _, v := range newVolumes {
		_, err := rn.buildVolumeInterface(v)
		if err != nil {
			return err
		}
		newVolumeByName[v.Name] = &v
	}

	needDown := false
	var deleteVolumes []boxspec.BoxVolumeSpec
	var createVolumes []boxspec.BoxVolumeSpec

	for _, oldVolume := range oldVolumesByName {
		if _, ok := newVolumeByName[oldVolume.Name]; !ok {
			slog.InfoContext(ctx, "need to down services due to volume being deleted", slog.Any("volumeName", oldVolume.Name))
			needDown = true
			deleteVolumes = append(deleteVolumes, *oldVolume)
		}
	}
	for _, newVolume := range newVolumeByName {
		vi, err := rn.buildVolumeInterface(*newVolume)
		if err != nil {
			return err
		}
		if oldVolume, ok := oldVolumesByName[newVolume.Name]; ok {
			oldVi, err := rn.buildVolumeInterface(*oldVolume)
			if err != nil {
				return err
			}
			changed := false
			if reflect.TypeOf(vi) != reflect.TypeOf(oldVi) {
				slog.InfoContext(ctx, "need to down services due to volume type change", slog.Any("name", newVolume.Name))
				changed = true
			} else if vi.CheckRecreateNeeded(*oldVolume, *newVolume) {
				slog.InfoContext(ctx, "need to down services due to volume spec change", slog.Any("name", newVolume.Name))
				changed = true
			}
			if changed {
				needDown = true
				deleteVolumes = append(deleteVolumes, *oldVolume)
				createVolumes = append(createVolumes, *newVolume)
			}
		} else {
			createVolumes = append(createVolumes, *newVolume)
		}
	}
	if allowDownService && needDown {
		err := rn.runComposeDown(ctx)
		if err != nil {
			return err
		}
	}

	for _, v := range deleteVolumes {
		vi, err := rn.buildVolumeInterface(v)
		if err != nil {
			return err
		}
		err = vi.Delete(ctx, v)
		if err != nil {
			return err
		}
		err = rn.deleteVolumeSpec(vi, v)
		if err != nil {
			return err
		}
	}
	for _, v := range createVolumes {
		vi, err := rn.buildVolumeInterface(v)
		if err != nil {
			return err
		}
		err = vi.Create(ctx, v)
		if err != nil {
			return err
		}
		err = rn.writeVolumeSpec(vi, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func fixVolumePermissions(vol boxspec.BoxVolumeSpec, volumeDir string) error {
	err := os.Chown(volumeDir, int(vol.RootUid), int(vol.RootGid))
	if err != nil {
		return err
	}
	rootMode, err := parseMode(vol.RootMode)
	if err != nil {
		return fmt.Errorf("failed to parse root dir mode: %w", err)
	}
	if rootMode != 0 {
		err = os.Chmod(volumeDir, rootMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseMode(s string) (os.FileMode, error) {
	n, err := strconv.ParseInt(s, 8, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode %s: %w", s, err)
	}
	fm := os.FileMode(n)
	if fm & ^os.ModePerm != 0 {
		return 0, fmt.Errorf("invalid file mode %s", s)
	}
	return fm, nil
}
