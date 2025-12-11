package volume_backup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/dboxed/dboxed/pkg/volume/rustic"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dboxed/dboxed/pkg/volume/webdavproxy"
)

type VolumeBackup struct {
	Client *baseclient.Client
	Volume *volume.Volume

	VolumeId string
	MountId  string

	RusticPassword        string
	Mount                 string
	SnapshotMount         string
	WebdavProxyListenAddr string
}

func (vb *VolumeBackup) Backup(ctx context.Context) (*models.VolumeSnapshot, error) {
	snapshotName := "backup-snapshot"
	rusticHost := fmt.Sprintf("dboxed-volume-%s", vb.VolumeId)

	_ = command_helper.RunCommand(ctx, "sync")

	c2 := clients.VolumesClient{Client: vb.Client}
	v, err := c2.GetVolumeById(ctx, vb.VolumeId)
	if err != nil {
		return nil, err
	}

	err = vb.Volume.UnmountSnapshot(ctx, snapshotName)
	if err != nil {
		return nil, err
	}

	err = vb.Volume.CreateSnapshot(ctx, snapshotName, true)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := vb.Volume.DeleteSnapshot(ctx, snapshotName)
		if err != nil {
			slog.ErrorContext(ctx, "backup snapshot deletion failed", slog.Any("error", err))
		}
	}()

	err = vb.Volume.MountSnapshot(ctx, snapshotName, vb.SnapshotMount)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := vb.Volume.UnmountSnapshot(ctx, snapshotName)
		if err != nil {
			slog.Error("deferred unmounting failed", slog.Any("error", err))
		}
	}()
	tags := BuildBackupTags(v.VolumeProvider.ID, &vb.VolumeId, &vb.MountId)

	var rsSnapshot *rustic.Snapshot
	err = vb.runWithWebdavProxy(ctx, v, func(config rustic.RusticConfig) error {
		var err error
		rsSnapshot, err = rustic.RunBackup(ctx, config, vb.SnapshotMount, rustic.BackupOpts{
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
	if err != nil {
		return nil, err
	}

	snapshot, err := c2.CreateSnapshot(ctx, vb.VolumeId, models.CreateVolumeSnapshot{
		MountId: vb.MountId,
		Rustic:  rsSnapshot.ToApi(),
	})
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (vb *VolumeBackup) RestoreSnapshot(ctx context.Context, snapshotId string, delete bool) error {
	c2 := clients.VolumesClient{Client: vb.Client}
	v, err := c2.GetVolumeById(ctx, vb.VolumeId)
	if err != nil {
		return err
	}

	err = vb.runWithWebdavProxy(ctx, v, func(config rustic.RusticConfig) error {
		return rustic.RunRestore(ctx, config, snapshotId, vb.Mount, rustic.RestoreOpts{
			NumericId: true,
			Delete:    delete,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (vb *VolumeBackup) runWithWebdavProxy(ctx context.Context, v *models.Volume, fn func(config rustic.RusticConfig) error) error {
	if v.VolumeProvider.Type != dmodel.VolumeProviderTypeRustic {
		return fmt.Errorf("not a rustic volume provider")
	}
	if v.VolumeProvider.Rustic.StorageType != dmodel.VolumeProviderStorageTypeS3 {
		return fmt.Errorf("not a S3 based rustic volume provider")
	}

	c3 := clients.S3BucketsClient{Client: vb.Client}
	b, err := c3.GetS3BucketById(ctx, *v.VolumeProvider.Rustic.S3BucketId)
	if err != nil {
		return err
	}

	fs := webdavproxy.NewFileSystem(ctx, vb.Client, b.ID, v.VolumeProvider.Rustic.StoragePrefix)

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
