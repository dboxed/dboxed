package volume_backup

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/dboxed/dboxed/pkg/volume/webdavproxy"
	"github.com/pelletier/go-toml/v2"
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

	_ = util.RunCommand("sync")

	err := vb.Volume.UnmountSnapshot(snapshotName)
	if err != nil {
		return err
	}

	err = vb.Volume.CreateSnapshot(snapshotName, true)
	if err != nil {
		return err
	}
	defer func() {
		err := vb.Volume.DeleteSnapshot(snapshotName)
		if err != nil {
			slog.ErrorContext(ctx, "backup snapshot deletion failed", slog.Any("error", err))
		}
	}()

	err = vb.Volume.MountSnapshot(snapshotName, vb.SnapshotMount)
	if err != nil {
		return err
	}
	defer func() {
		err := vb.Volume.UnmountSnapshot(snapshotName)
		if err != nil {
			slog.Error("deferred unmounting failed", slog.Any("error", err))
		}
	}()

	var stdout []byte
	err = vb.runRustic(ctx, func(configDir string) error {
		rusticArgs := []string{
			"backup",
			"--init",
			"--host", rusticHost,
			"--as-path", "/",
			"--with-atime",
			"--no-scan",
			"--json",
			vb.SnapshotMount,
		}
		c := util.CommandHelper{
			Command:     "rustic",
			Args:        rusticArgs,
			Dir:         configDir,
			CatchStdout: true,
		}
		err := c.Run()
		if err != nil {
			return err
		}
		stdout = c.Stdout
		return nil
	})
	if err != nil {
		return err
	}

	var rs RusticSnapshot
	err = json.Unmarshal(stdout, &rs)
	if err != nil {
		return err
	}

	req := models.CreateVolumeSnapshot{
		LockID:   vb.LockId,
		IsLatest: true,
	}
	req.Rustic = util.Ptr(rs.ToApi())

	c2 := clients.VolumesClient{Client: vb.Client}

	slog.InfoContext(ctx, "creating snapshot api object")
	_, err = c2.VolumeCreateSnapshot(ctx, vb.VolumeId, req)
	if err != nil {
		return err
	}

	return nil
}

func (vb *VolumeBackup) RestoreSnapshot(ctx context.Context, snapshotId string) error {
	err := vb.runRustic(ctx, func(configDir string) error {
		rusticArgs := []string{"restore",
			"--numeric-id",
			snapshotId,
			vb.Mount,
		}
		c := util.CommandHelper{
			Command:     "rustic",
			Args:        rusticArgs,
			Dir:         configDir,
			CatchStdout: true,
		}
		err := c.Run()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (vb *VolumeBackup) runRustic(ctx context.Context, fn func(configDir string) error) error {
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

	configDir, err := vb.buildRusticConfigDir(wdpAddr.String())
	if err != nil {
		return err
	}
	defer os.RemoveAll(configDir)

	return fn(configDir)
}

func (vb *VolumeBackup) buildRusticConfigDir(webdavAddr string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	doRm := true
	defer func() {
		if doRm {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	config := RusticConfig{
		Repository: RusticConfigRepository{
			Repository: "opendal:webdav",
			Password:   vb.RusticPassword,
			Options: RusticConfigRepositoryOptions{
				Endpoint: fmt.Sprintf("http://%s", webdavAddr),
			},
		},
	}
	configBytes, err := toml.Marshal(config)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filepath.Join(tmpDir, "rustic.toml"), configBytes, 0600)
	if err != nil {
		return "", err
	}

	doRm = false
	return tmpDir, nil
}
