package volume_backup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/rustic"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dboxed/dboxed/pkg/volume/webdavproxy"
)

type VolumeBackup struct {
	Client *baseclient.Client
	Volume *volume.Volume

	VolumeProviderId int64
	VolumeId         int64
	VolumeUuid       string
	LockId           string

	RusticPassword        string
	Mount                 string
	SnapshotMount         string
	WebdavProxyListenAddr string
}

func (vb *VolumeBackup) Backup(ctx context.Context) error {
	snapshotName := "backup-snapshot"
	rusticHost := fmt.Sprintf("dboxed-volume-%s", vb.VolumeUuid)

	_ = util.RunCommand(ctx, "sync")

	err := vb.Volume.UnmountSnapshot(ctx, snapshotName)
	if err != nil {
		return err
	}

	err = vb.Volume.CreateSnapshot(ctx, snapshotName, true)
	if err != nil {
		return err
	}
	defer func() {
		err := vb.Volume.DeleteSnapshot(ctx, snapshotName)
		if err != nil {
			slog.ErrorContext(ctx, "backup snapshot deletion failed", slog.Any("error", err))
		}
	}()

	err = vb.Volume.MountSnapshot(ctx, snapshotName, vb.SnapshotMount)
	if err != nil {
		return err
	}
	defer func() {
		err := vb.Volume.UnmountSnapshot(ctx, snapshotName)
		if err != nil {
			slog.Error("deferred unmounting failed", slog.Any("error", err))
		}
	}()

	tags := BuildBackupTags(vb.VolumeProviderId, &vb.VolumeId, &vb.VolumeUuid, &vb.LockId)

	err = vb.runWithWebdavProxy(ctx, func(config rustic.RusticConfig) error {
		var err error
		_, err = rustic.RunBackup(ctx, config, vb.SnapshotMount, rustic.BackupOpts{
			Init:      true,
			Host:      &rusticHost,
			AsPath:    util.Ptr("/"),
			WithAtime: true,
			NoScan:    true,
			Tags:      tags,
		})
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

func (vb *VolumeBackup) RestoreSnapshot(ctx context.Context, snapshotId string) error {
	err := vb.runWithWebdavProxy(ctx, func(config rustic.RusticConfig) error {
		return rustic.RunRestore(ctx, config, snapshotId, vb.Mount, rustic.RestoreOpts{
			NumericIds: true,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (vb *VolumeBackup) runWithWebdavProxy(ctx context.Context, fn func(config rustic.RusticConfig) error) error {
	fs := webdavproxy.NewFileSystem(ctx, vb.Client, vb.VolumeProviderId)

	webdavProxy, err := webdavproxy.NewProxy(fs, vb.WebdavProxyListenAddr)
	if err != nil {
		return err
	}
	wdpAddr, err := webdavProxy.Start(ctx)
	if err != nil {
		return err
	}
	defer webdavProxy.Stop()

	config := vb.buildRusticConfig(wdpAddr.String())

	return fn(config)
}

func (vb *VolumeBackup) buildRusticConfig(webdavAddr string) rustic.RusticConfig {
	config := rustic.RusticConfig{
		Repository: rustic.RusticConfigRepository{
			Repository: "opendal:webdav",
			Password:   vb.RusticPassword,
			Options: rustic.RusticConfigRepositoryOptions{
				Endpoint: fmt.Sprintf("http://%s", webdavAddr),
			},
		},
	}
	return config
}
