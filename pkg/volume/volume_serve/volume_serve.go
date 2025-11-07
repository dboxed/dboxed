package volume_serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/volume/mount"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dboxed/dboxed/pkg/volume/volume_backup"
	"github.com/dustin/go-humanize"
)

type VolumeServeOpts struct {
	Client *baseclient.Client

	MountName string

	VolumeId string
	BoxId    *string

	Dir            string
	BackupInterval time.Duration

	WebdavProxyListen string
}

type VolumeServe struct {
	opts VolumeServeOpts

	volume *models.Volume
	log    *slog.Logger

	LocalVolume *volume.Volume

	stop chan struct{}
}

func New(opts VolumeServeOpts) (*VolumeServe, error) {
	vs := &VolumeServe{
		opts: opts,
		stop: make(chan struct{}),
	}

	dir, err := normalizePath(opts.Dir)
	if err != nil {
		return nil, err
	}
	vs.opts.Dir = dir

	vs.log = slog.With(
		slog.Any("volumeId", opts.VolumeId),
		slog.Any("dir", opts.Dir),
	)
	if opts.BoxId != nil {
		vs.log = vs.log.With(slog.Any("boxId", *opts.BoxId))
	}

	return vs, nil
}

func (vs *VolumeServe) buildClient(ctx context.Context, s *VolumeState) (*baseclient.Client, error) {
	if vs.opts.Client != nil {
		return vs.opts.Client, nil
	}

	c, err := baseclient.New(nil, s.ClientAuth, false)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (vs *VolumeServe) Create(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if s != nil {
		return fmt.Errorf("%s already contains a potentially mounted volume", vs.opts.Dir)
	}

	err = os.MkdirAll(vs.GetMountDir(), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(vs.opts.Dir, "snapshot"), 0700)
	if err != nil {
		return err
	}

	err = vs.refreshVolumeMount(ctx)
	if err != nil {
		return err
	}
	vs.log = vs.log.With(slog.Any("mountId", *vs.volume.MountId))

	if vs.volume.Rustic == nil {
		return fmt.Errorf("only rustic is supported for now")
	}

	lvmTags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-%s", vs.volume.ID),
		fmt.Sprintf("dboxed-volume-mount-%s", *vs.volume.MountId),
	}

	image := filepath.Join(vs.opts.Dir, "image")

	imageSize := vs.volume.Rustic.FsSize * 2
	vs.log.Info("creating local volume image",
		slog.Any("path", image),
		slog.Any("imageSize", humanize.Bytes(uint64(imageSize))),
		slog.Any("fsSize", humanize.Bytes(uint64(vs.volume.Rustic.FsSize))),
		slog.Any("fsType", vs.volume.Rustic.FsType),
		slog.Any("lvmTags", lvmTags),
	)
	err = volume.Create(ctx, volume.CreateOptions{
		MountId:   *vs.volume.MountId,
		ImagePath: image,
		ImageSize: imageSize,
		FsSize:    vs.volume.Rustic.FsSize,
		FsType:    vs.volume.Rustic.FsType,
		LvmTags:   lvmTags,
	})
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Open(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}
	if s == nil || s.Volume == nil || s.Volume.MountId == nil {
		return fmt.Errorf("%s does not contain a mounted volume", vs.opts.Dir)
	}

	err = vs.refreshVolumeMount(ctx)
	if err != nil {
		return err
	}

	image := filepath.Join(vs.opts.Dir, "image")
	vs.log.Info("opening local volume image",
		slog.Any("path", image),
	)

	vs.LocalVolume, err = volume.Open(ctx, image, *vs.volume.MountId)
	if err != nil {
		return err
	}

	refMountDir := filepath.Join(vs.opts.Dir, "loop-ref")
	err = volume.WriteLoopRef(ctx, refMountDir, *vs.volume.MountId)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Deactivate(ctx context.Context) error {
	err := vs.LocalVolume.Deactivate(ctx)
	if err != nil {
		return err
	}

	refMountDir := filepath.Join(vs.opts.Dir, "loop-ref")
	err = volume.UnmountLoopRefs(ctx, refMountDir)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VolumeServe) GetMountDir() string {
	return filepath.Join(vs.opts.Dir, "mount")
}

func (vs *VolumeServe) MountDevice(ctx context.Context, readOnly bool) error {
	mountDir := vs.GetMountDir()

	mountInfo, err := mount.GetMountByMountpoint(mountDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	devPath, err := vs.LocalVolume.DevPath(true)
	if err != nil {
		return err
	}

	if mountInfo != nil {
		if mountInfo.Source != devPath {
			return fmt.Errorf("mount point %s is already mounted from source %s and type %s", mountDir, mountInfo.Source, mountInfo.FSType)
		}
		opts := strings.Split(mountInfo.Options, ",")
		if slices.Contains(opts, "ro") && !readOnly {
			return fmt.Errorf("mount point %s is already mounted but it is read-only", mountDir)
		}
	}

	vs.log.Info("mounting volume", slog.Any("mountPath", mountDir))
	err = vs.LocalVolume.Mount(ctx, mountDir, readOnly)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) MountVolumeViaApi(ctx context.Context) error {
	err := vs.refreshVolumeMount(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) ReleaseVolumeMountViaApi(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if s == nil || s.Volume == nil || s.Volume.MountId == nil {
		return fmt.Errorf("volume is not mounted")
	}

	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return err
	}

	c2 := &clients.VolumesClient{Client: c}

	req := models.VolumeReleaseRequest{
		MountId: *s.Volume.MountId,
	}

	newVolume, err := c2.VolumeReleaseMount(ctx, vs.opts.VolumeId, req)
	if err != nil {
		return err
	}

	vs.volume = newVolume

	s.Volume = newVolume

	err = vs.saveVolumeState(*s)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Run(ctx context.Context) error {
	return vs.periodicBackup(ctx)
}

func (vs *VolumeServe) Stop() {
	close(vs.stop)
}

func (vs *VolumeServe) refreshVolumeMount(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if s == nil {
		if vs.opts.Client == nil {
			return fmt.Errorf("no volume state loaded and missing client")
		}

		s = &VolumeState{
			ClientAuth: vs.opts.Client.GetClientAuth(true),
			MountName:  vs.opts.MountName,
		}
	}

	var prevMountId *string
	if s.Volume != nil {
		prevMountId = s.Volume.MountId
	}

	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	var newVolume *models.Volume
	if prevMountId == nil {
		vs.log.Info("mounting volume")
		mountRequest := models.VolumeMountRequest{
			BoxId: vs.opts.BoxId,
		}
		newVolume, err = c2.VolumeMount(ctx, vs.opts.VolumeId, mountRequest)
		if err != nil {
			return err
		}
	} else {
		vs.log.Debug("refreshing mount", slog.Any("prevMountId", *prevMountId))
		refreshMountRequest := models.VolumeRefreshMountRequest{
			MountId: *prevMountId,
		}
		newVolume, err = c2.VolumeRefreshMount(ctx, vs.opts.VolumeId, refreshMountRequest)
		if err != nil {
			return err
		}
	}

	vs.volume = newVolume
	s.Volume = newVolume

	err = vs.saveVolumeState(*s)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) buildVolumeBackup(ctx context.Context, s *VolumeState) (*volume_backup.VolumeBackup, error) {
	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return nil, err
	}

	mount := filepath.Join(vs.opts.Dir, "mount")
	snapshotMount := filepath.Join(vs.opts.Dir, "snapshot")

	vb := &volume_backup.VolumeBackup{
		Client:                c,
		Volume:                vs.LocalVolume,
		VolumeId:              vs.volume.ID,
		MountId:               *vs.volume.MountId,
		RusticPassword:        vs.volume.Rustic.Password,
		Mount:                 mount,
		SnapshotMount:         snapshotMount,
		WebdavProxyListenAddr: vs.opts.WebdavProxyListen,
	}
	return vb, nil
}

func (vs *VolumeServe) BackupOnce(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}

	vb, err := vs.buildVolumeBackup(ctx, s)
	if err != nil {
		return err
	}

	err = vb.Backup(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) IsRestoreDone() (bool, error) {
	s, err := vs.loadVolumeState()
	if err != nil {
		return false, err
	}
	return s.RestoreDone, nil
}

func (vs *VolumeServe) RestoreSnapshot(ctx context.Context, snapshotId string, delete bool) error {
	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}
	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	snapshot, err := c2.GetVolumeSnapshotById(ctx, vs.volume.ID, snapshotId)
	if err != nil {
		return err
	}

	vb, err := vs.buildVolumeBackup(ctx, s)
	if err != nil {
		return err
	}

	err = vb.RestoreSnapshot(ctx, snapshot.Rustic.SnapshotId, delete)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) RestoreFromLatestSnapshot(ctx context.Context) error {
	if vs.volume.LatestSnapshotId != nil {
		err := vs.RestoreSnapshot(ctx, *vs.volume.LatestSnapshotId, true)
		if err != nil {
			return err
		}
	}

	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}

	s.RestoreDone = true
	s.RestoreSnapshot = vs.volume.LatestSnapshotId

	err = vs.saveVolumeState(*s)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) periodicBackup(ctx context.Context) error {
	refreshMountInterval := time.Second * 15
	nextRefreshMountTimer := time.NewTimer(refreshMountInterval)
	nextBackupTimer := time.NewTimer(vs.opts.BackupInterval)

	doRefreshMountVolume := func() error {
		err := vs.refreshVolumeMount(ctx)
		if err != nil {
			vs.log.Error("refreshing volume mount", slog.Any("error", err))
			if baseclient.IsBadRequest(err) {
				// volume most likely got force-released
				return err
			}
		}
		nextRefreshMountTimer = time.NewTimer(refreshMountInterval)
		return nil
	}

	for {
		select {
		case <-vs.stop:
			return nil
		case <-nextBackupTimer.C:
			err := doRefreshMountVolume()
			if err != nil {
				return err
			}
			err = vs.BackupOnce(ctx)
			if err != nil {
				vs.log.Error("backup failed", slog.Any("error", err))
			}
			nextBackupTimer = time.NewTimer(vs.opts.BackupInterval)
		case <-nextRefreshMountTimer.C:
			err := doRefreshMountVolume()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func normalizePath(path string) (realPath string, err error) {
	if realPath, err = filepath.Abs(path); err != nil {
		return "", err
	}
	if realPath, err = filepath.EvalSymlinks(realPath); err != nil {
		return "", err
	}
	if _, err := os.Stat(realPath); err != nil {
		return "", err
	}
	return realPath, nil
}
