package volume_serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
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

	log *slog.Logger

	LocalVolume *volume.Volume

	stop chan struct{}

	volumeStateMutex sync.Mutex
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

func (vs *VolumeServe) buildClient(s *VolumeState) (*baseclient.Client, error) {
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

	err = vs.mountVolumeViaApi(ctx)
	if err != nil {
		return err
	}

	s, err = vs.loadVolumeState()
	if err != nil {
		return err
	}

	vs.log = vs.log.With(slog.Any("mountId", *s.Volume.MountId))

	if s.Volume.Rustic == nil {
		return fmt.Errorf("only rustic is supported for now")
	}

	lvmTags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-%s", s.Volume.ID),
		fmt.Sprintf("dboxed-volume-mount-%s", *s.Volume.MountId),
	}

	image := filepath.Join(vs.opts.Dir, "image")

	imageSize := s.Volume.Rustic.FsSize * 2
	vs.log.Info("creating local volume image",
		slog.Any("path", image),
		slog.Any("imageSize", humanize.Bytes(uint64(imageSize))),
		slog.Any("fsSize", humanize.Bytes(uint64(s.Volume.Rustic.FsSize))),
		slog.Any("fsType", s.Volume.Rustic.FsType),
		slog.Any("lvmTags", lvmTags),
	)
	err = volume.Create(ctx, volume.CreateOptions{
		MountId:   *s.Volume.MountId,
		ImagePath: image,
		ImageSize: imageSize,
		FsSize:    s.Volume.Rustic.FsSize,
		FsType:    s.Volume.Rustic.FsType,
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

	vs.LocalVolume, err = volume.Open(ctx, image, *s.Volume.MountId)
	if err != nil {
		return err
	}

	refMountDir := filepath.Join(vs.opts.Dir, "loop-ref")
	err = volume.WriteLoopRef(ctx, refMountDir, *s.Volume.MountId)
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

func (vs *VolumeServe) ReleaseVolumeMountViaApi(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if s == nil || s.Volume == nil || s.Volume.MountId == nil {
		return fmt.Errorf("volume is not mounted")
	}

	c, err := vs.buildClient(s)
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

	err = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
		s.Volume = newVolume
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Run(ctx context.Context) error {
	err := vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
		s.ServeStartTime = util.Ptr(time.Now())
		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		err := vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
			s.ServeStopTime = util.Ptr(time.Now())
			return nil
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to update volume state", "error", err)
		}
	}()

	return vs.periodicBackup(ctx)
}

func (vs *VolumeServe) Stop() {
	close(vs.stop)
}

func (vs *VolumeServe) mountVolumeViaApi(ctx context.Context) error {
	if vs.opts.Client == nil {
		return fmt.Errorf("missing client")
	}

	c2 := clients.VolumesClient{Client: vs.opts.Client}

	var newVolume *models.Volume
	vs.log.Info("mounting volume")
	mountRequest := models.VolumeMountRequest{
		BoxId: vs.opts.BoxId,
	}
	newVolume, err := c2.VolumeMount(ctx, vs.opts.VolumeId, mountRequest)
	if err != nil {
		return err
	}

	s := &VolumeState{
		ClientAuth: vs.opts.Client.GetClientAuth(true),
		MountName:  vs.opts.MountName,
		Volume:     newVolume,
	}
	err = vs.saveVolumeState(s)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) refreshVolumeMount(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}
	if s.Volume.MountId == nil {
		return fmt.Errorf("volume has no mount id")
	}

	c, err := vs.buildClient(s)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	vs.log.Debug("refreshing mount", slog.Any("mountId", *s.Volume.MountId))

	refreshMountRequest := models.VolumeRefreshMountRequest{
		MountId:           *s.Volume.MountId,
		SnapshotStartTime: s.SnapshotStartTime,
		SnapshotEndTime:   s.SnapshotEndTime,
	}
	if s.LastFinishedSnapshot != nil {
		refreshMountRequest.LastFinishedSnapshotId = &s.LastFinishedSnapshot.ID
	}

	st, err := mount.StatFS(vs.GetMountDir())
	if err != nil {
		vs.log.Error(err.Error())
	} else {
		refreshMountRequest.VolumeTotalSize = st.TotalSize
		refreshMountRequest.VolumeFreeSize = st.FreeSize
	}

	newVolume, err := c2.VolumeRefreshMount(ctx, vs.opts.VolumeId, refreshMountRequest)
	if err != nil {
		return err
	}

	err = vs.updateVolumeState(ctx, false, func(s *VolumeState) error {
		s.Volume = newVolume
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) buildVolumeBackup(ctx context.Context, s *VolumeState) (*volume_backup.VolumeBackup, error) {
	c, err := vs.buildClient(s)
	if err != nil {
		return nil, err
	}

	mount := filepath.Join(vs.opts.Dir, "mount")
	snapshotMount := filepath.Join(vs.opts.Dir, "snapshot")

	vb := &volume_backup.VolumeBackup{
		Client:                c,
		Volume:                vs.LocalVolume,
		VolumeId:              s.Volume.ID,
		MountId:               *s.Volume.MountId,
		RusticPassword:        s.Volume.Rustic.Password,
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

	err = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
		s.SnapshotStartTime = util.Ptr(time.Now())
		s.SnapshotEndTime = nil
		s.SnapshotError = nil
		return nil
	})
	if err != nil {
		return err
	}

	snapshot, err := vb.Backup(ctx)
	if err != nil {
		_ = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
			s.SnapshotError = util.Ptr(err.Error())
			return nil
		})
		return err
	}

	err = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
		s.LastFinishedSnapshot = snapshot
		s.SnapshotEndTime = util.Ptr(time.Now())
		return nil
	})
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
	c, err := vs.buildClient(s)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	snapshot, err := c2.GetVolumeSnapshotById(ctx, s.Volume.ID, snapshotId)
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

	err = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
		s.RestoreDone = true
		s.RestoreSnapshot = &snapshot.ID
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) RestoreFromLatestSnapshot(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}

	if s.Volume.LatestSnapshotId != nil {
		err := vs.RestoreSnapshot(ctx, *s.Volume.LatestSnapshotId, true)
		if err != nil {
			return err
		}
	} else {
		err = vs.updateVolumeState(ctx, true, func(s *VolumeState) error {
			s.RestoreDone = true
			s.RestoreSnapshot = nil
			return nil
		})
		if err != nil {
			return err
		}
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
