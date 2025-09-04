package box_spec_runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed-volume/pkg/client"
	"github.com/dboxed/dboxed-volume/pkg/volume_serve"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) createDboxedVolume(ctx context.Context, vol types.BoxVolumeSpec) error {
	workDir := rn.getVolumeWorkDirOnHost(vol)
	mountDir := rn.getVolumeMountDirOnHost(vol)

	imagePath := filepath.Join(workDir, "dboxed-volume-image")
	snapshotMountPath := filepath.Join(workDir, "dboxed-volume-snapshot")

	err := os.MkdirAll(snapshotMountPath, 0700)
	if err != nil {
		return err
	}

	c, err := client.New(vol.Dboxed.ApiUrl, &vol.Dboxed.Token)
	if err != nil {
		return err
	}

	prevLockId, err := rn.readDboxedVolumeLockId(vol)
	if err != nil {
		return err
	}

	backupInterval, err := time.ParseDuration(vol.Dboxed.BackupInterval)
	if err != nil {
		return err
	}

	vs := volume_serve.VolumeServe{
		Client:            c,
		RepositoryId:      vol.Dboxed.RepositoryId,
		VolumeId:          vol.Dboxed.VolumeId,
		PrevLockId:        prevLockId,
		Image:             imagePath,
		Mount:             mountDir,
		SnapshotMount:     snapshotMountPath,
		BackupInterval:    backupInterval,
		WebdavProxyListen: "127.0.0.1:0",
	}
	err = vs.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) readDboxedVolumeLockId(vol types.BoxVolumeSpec) (*string, error) {
	p := filepath.Join(rn.getVolumeWorkDirOnHost(vol), "dboxed-volume-lock-id")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	s := strings.TrimSpace(string(b))
	return &s, nil
}

func (rn *BoxSpecRunner) writeDboxedVolumeLockId(vol types.BoxVolumeSpec, lockId string) error {
	p := filepath.Join(rn.getVolumeWorkDirOnHost(vol), "dboxed-volume-lock-id")
	err := os.MkdirAll(filepath.Dir(p), 0700)
	if err != nil {
		return err
	}
	return util.AtomicWriteFile(p, []byte(lockId), 0644)
}
