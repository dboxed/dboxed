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
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dboxed/dboxed/pkg/volume/volume_backup"
	"github.com/dustin/go-humanize"
	"github.com/moby/sys/mountinfo"
)

type VolumeServeOpts struct {
	Client *baseclient.Client

	VolumeId int64

	Dir            string
	BackupInterval time.Duration

	WebdavProxyListen string
}

type VolumeServe struct {
	opts VolumeServeOpts

	volumeProvider *models.VolumeProvider
	volume         *models.Volume
	log            *slog.Logger

	LocalVolume *volume.Volume

	stop chan struct{}
	wg   sync.WaitGroup
	m    sync.Mutex
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

	return vs, nil
}

func (vs *VolumeServe) buildClient(ctx context.Context, s *VolumeState) (*baseclient.Client, error) {
	if vs.opts.Client != nil {
		return vs.opts.Client, nil
	}

	c, err := baseclient.FromClientAuth(nil, s.ClientAuth, false)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (vs *VolumeServe) getVolume(ctx context.Context, s *VolumeState) error {
	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return err
	}

	c2 := &clients.VolumesClient{Client: c}
	c3 := &clients.VolumeProvidersClient{Client: c}

	vs.volume, err = c2.GetVolumeById(ctx, vs.opts.VolumeId)
	if err != nil {
		return err
	}

	vs.volumeProvider, err = c3.GetVolumeProviderById(ctx, vs.volume.VolumeProvider)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Create(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if s != nil {
		return fmt.Errorf("%s already contains a potentially locked volume", vs.opts.Dir)
	}

	err = vs.getVolume(ctx, s)
	if err != nil {
		return err
	}

	if vs.volume.Rustic == nil {
		return fmt.Errorf("only rustic is supported for now")
	}

	err = os.MkdirAll(vs.GetMountDir(), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(vs.opts.Dir, "snapshot"), 0700)
	if err != nil {
		return err
	}

	_, err = vs.lockVolume(ctx)
	if err != nil {
		return err
	}
	vs.log = vs.log.With(slog.Any("lockId", *vs.volume.LockId))

	lvmTags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-%s", vs.volume.Uuid),
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
	err = volume.Create(volume.CreateOptions{
		LockId:    *vs.volume.LockId,
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
	if s == nil || s.LockId == nil {
		return fmt.Errorf("%s does not contain a locked volume", vs.opts.Dir)
	}

	err = vs.getVolume(ctx, s)
	if err != nil {
		return err
	}

	image := filepath.Join(vs.opts.Dir, "image")
	vs.log.Info("opening local volume image",
		slog.Any("path", image),
	)

	vs.LocalVolume, err = volume.Open(image, *vs.volume.LockId)
	if err != nil {
		return err
	}

	refMountDir := filepath.Join(vs.opts.Dir, "loop-ref")
	err = volume.WriteLoopRef(refMountDir, *vs.volume.LockId)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Deactivate() error {
	err := vs.LocalVolume.Deactivate()
	if err != nil {
		return err
	}

	refMountDir := filepath.Join(vs.opts.Dir, "loop-ref")
	err = volume.UnmountLoopRefs(refMountDir)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VolumeServe) GetMountDir() string {
	return filepath.Join(vs.opts.Dir, "mount")
}

func (vs *VolumeServe) Mount(ctx context.Context, readOnly bool) error {
	mount := vs.GetMountDir()

	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		return err
	}
	var mountInfo *mountinfo.Info
	for _, m := range mounts {
		if m.Mountpoint == mount {
			mountInfo = m
			break
		}
	}

	devPath, err := vs.LocalVolume.DevPath(true)
	if err != nil {
		return err
	}

	if mountInfo != nil {
		source, err := filepath.EvalSymlinks(mountInfo.Source)
		if err != nil {
			return err
		}
		if source != devPath {
			return fmt.Errorf("mount point %s is already mounted from source %s and type %s", mount, mountInfo.Source, mountInfo.FSType)
		}
		opts := strings.Split(mountInfo.Options, ",")
		if slices.Contains(opts, "ro") {
			return fmt.Errorf("mount point %s is already mounted but it is read-only", mount)
		}
	}

	vs.log.Info("mounting volume", slog.Any("mountPath", mount))
	err = vs.LocalVolume.Mount(mount, readOnly)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Lock(ctx context.Context) error {
	_, err := vs.lockVolume(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Unlock(ctx context.Context) error {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if s == nil || s.LockId == nil {
		return fmt.Errorf("volume is not locked")
	}

	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return err
	}

	c2 := &clients.VolumesClient{Client: c}

	req := models.VolumeReleaseRequest{
		LockId: *s.LockId,
	}

	newVolume, err := c2.VolumeRelease(ctx, vs.volume.ID, req)
	if err != nil {
		return err
	}

	vs.m.Lock()
	defer vs.m.Unlock()
	vs.volume = newVolume

	s.LockId = nil
	err = vs.saveVolumeState(*s)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) Start(ctx context.Context) {
	vs.wg.Add(1)
	go vs.periodicRefreshLock(ctx)

	vs.wg.Add(1)
	go vs.periodicBackup(ctx)
}

func (vs *VolumeServe) Stop() {
	close(vs.stop)
	vs.wg.Wait()
}

func (vs *VolumeServe) lockVolume(ctx context.Context) (bool, error) {
	s, err := vs.loadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if s == nil {
		if vs.opts.Client == nil {
			return false, fmt.Errorf("no volume state loaded and missing client")
		}

		s = &VolumeState{
			ClientAuth: vs.opts.Client.GetClientAuth(),
			VolumeId:   vs.volume.ID,
			VolumeUuid: vs.volume.Uuid,
		}
	}

	prevLockId := s.LockId

	if prevLockId == nil {
		vs.log.Info("locking volume")
	} else {
		vs.log.Info("refreshing lock", slog.Any("prevLockId", *prevLockId))
	}

	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return false, err
	}

	c2 := clients.VolumesClient{Client: c}

	lockRequest := models.VolumeLockRequest{
		PrevLockId: prevLockId,
	}
	newVolume, err := c2.VolumeLock(ctx, vs.volume.ID, lockRequest)
	if err != nil {
		return false, err
	}

	vs.m.Lock()
	defer vs.m.Unlock()
	vs.volume = newVolume

	newLock := true
	if prevLockId == nil || *prevLockId != *vs.volume.LockId {
		newLock = false
		s.VolumeUuid = vs.volume.Uuid
		s.LockId = vs.volume.LockId
		err = vs.saveVolumeState(*s)
		if err != nil {
			return false, err
		}
	}
	vs.log.Info("volume locked", slog.Any("lockId", *vs.volume.LockId))

	return newLock, nil
}

func (vs *VolumeServe) buildVolumeBackup(ctx context.Context, s *VolumeState) (*volume_backup.VolumeBackup, error) {
	c, err := vs.buildClient(ctx, s)
	if err != nil {
		return nil, err
	}

	mount := filepath.Join(vs.opts.Dir, "mount")
	snapshotMount := filepath.Join(vs.opts.Dir, "snapshot")

	vs.m.Lock()
	defer vs.m.Unlock()
	vb := &volume_backup.VolumeBackup{
		Client:                c,
		Volume:                vs.LocalVolume,
		VolumeProviderId:      vs.volumeProvider.ID,
		VolumeId:              vs.volume.ID,
		VolumeUuid:            vs.volume.Uuid,
		LockId:                *vs.volume.LockId,
		RusticPassword:        vs.volumeProvider.Rustic.Password,
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

func (vs *VolumeServe) RestoreSnapshot(ctx context.Context, snapshotId int64) error {
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

	err = vb.RestoreSnapshot(ctx, snapshot.Rustic.SnapshotId)
	if err != nil {
		return err
	}

	return nil
}

func (vs *VolumeServe) RestoreFromLatestSnapshot(ctx context.Context) error {
	if vs.volume.LatestSnapshotId == nil {
		// nothing to do
		return nil
	}

	return vs.RestoreSnapshot(ctx, *vs.volume.LatestSnapshotId)
}

func (vs *VolumeServe) periodicBackup(ctx context.Context) {
	defer vs.wg.Done()

	for {
		select {
		case <-vs.stop:
			return
		case <-time.After(vs.opts.BackupInterval):
			err := vs.BackupOnce(ctx)
			if err != nil {
				vs.log.Error("backup failed", slog.Any("error", err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (vs *VolumeServe) periodicRefreshLock(ctx context.Context) {
	defer vs.wg.Done()
	for {
		_, err := vs.lockVolume(ctx)
		if err != nil {
			vs.log.Error("error in VolumeLock", slog.Any("error", err))
		}

		select {
		case <-vs.stop:
			return
		case <-time.After(15 * time.Second):
		case <-ctx.Done():
			return
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
