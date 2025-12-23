package volume_backup

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/dboxed/dboxed/pkg/volume/mount"
	"github.com/dboxed/dboxed/pkg/volume/restic"
	"github.com/dboxed/dboxed/pkg/volume/restic_rest_server"
	"github.com/dboxed/dboxed/pkg/volume/volume"
)

type VolumeBackup struct {
	Client *baseclient.Client
	Volume *volume.Volume

	VolumeId string
	MountId  string

	ResticPassword string
	Mount          string
	SnapshotMount  string

	ResticRestServerListenAddr string
}

func (vb *VolumeBackup) Backup(ctx context.Context) (*models.VolumeSnapshot, error) {
	snapshotName := "backup-snapshot"
	resticHost := fmt.Sprintf("dboxed-volume-%s", vb.VolumeId)

	_ = command_helper.RunCommand(ctx, "sync")

	c2 := clients.VolumesClient{Client: vb.Client}
	v, err := c2.GetVolumeById(ctx, vb.VolumeId)
	if err != nil {
		return nil, err
	}

	tags := BuildBackupTags(v.VolumeProvider.ID, &vb.VolumeId, &vb.MountId)

	err = vb.Volume.UnmountSnapshot(ctx, snapshotName)
	if err != nil {
		return nil, err
	}

	err = vb.Volume.CreateSnapshot(ctx, snapshotName, true, tags)
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

	var excludes []string
	m, err := mount.GetMountByMountpoint(vb.SnapshotMount)
	if err != nil {
		return nil, err
	}
	if m.FSType == "ext2" || m.FSType == "ext3" || m.FSType == "ext4" {
		excludes = append(excludes, filepath.Join(vb.SnapshotMount, "lost+found"))
	}

	var rsSnapshot *restic.Snapshot
	err = vb.runWithResticRestServer(ctx, v, func(env []string) error {
		var err error
		rsSnapshot, err = restic.RunBackup(ctx, env, vb.SnapshotMount, restic.BackupOpts{
			Host:      &resticHost,
			WithAtime: true,
			NoScan:    true,
			Tags:      tags,
			Exclude:   excludes,
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
		Restic:  rsSnapshot.ToApi(),
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

	err = vb.runWithResticRestServer(ctx, v, func(env []string) error {
		return restic.RunRestore(ctx, env, snapshotId, vb.Mount, restic.RestoreOpts{
			Delete: delete,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (vb *VolumeBackup) runWithResticRestServer(ctx context.Context, v *models.Volume, fn func(env []string) error) error {
	if v.VolumeProvider.Type != dmodel.VolumeProviderTypeRestic {
		return fmt.Errorf("not a restic volume provider")
	}
	if v.VolumeProvider.Restic.StorageType != dmodel.VolumeProviderStorageTypeS3 {
		return fmt.Errorf("not a S3 based restic volume provider")
	}

	c3 := clients.S3BucketsClient{Client: vb.Client}
	b, err := c3.GetS3BucketById(ctx, *v.VolumeProvider.Restic.S3BucketId)
	if err != nil {
		return err
	}

	restServer, err := restic_rest_server.NewServer(vb.Client, b.ID, v.VolumeProvider.Restic.StoragePrefix, vb.ResticRestServerListenAddr)
	if err != nil {
		return err
	}
	restAddr, err := restServer.Start(ctx)
	if err != nil {
		return err
	}
	defer restServer.Stop()

	env := vb.buildResticEnv(restAddr.String())

	return fn(env)
}

func (vb *VolumeBackup) buildResticEnv(restAddr string) []string {
	env := []string{
		fmt.Sprintf("RESTIC_REPOSITORY=rest:http://%s", restAddr),
		fmt.Sprintf("RESTIC_PASSWORD=%s", vb.ResticPassword),
	}
	return env
}
