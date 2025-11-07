package volume_serve

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"sigs.k8s.io/yaml"
)

type VolumeState struct {
	ClientAuth *baseclient.ClientAuth `json:"clientAuth,omitempty"`

	MountName string `json:"mountName"`

	Volume *models.Volume `json:"volume"`

	ServeStartTime *time.Time `json:"serveStartTime"`
	ServeStopTime  *time.Time `json:"serveStopTime"`

	LastFinishedSnapshot *models.VolumeSnapshot `json:"lastFinishedSnapshot"`
	SnapshotStartTime    *time.Time             `json:"snapshotStartTime"`
	SnapshotEndTime      *time.Time             `json:"snapshotEndTime"`
	SnapshotError        *string                `json:"snapshotError"`

	RestoreDone     bool    `json:"restoreDone"`
	RestoreSnapshot *string `json:"restoreSnapshot"`
}

func (vs *VolumeServe) loadVolumeState() (*VolumeState, error) {
	s, err := LoadVolumeState(vs.opts.Dir)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func LoadVolumeState(dir string) (*VolumeState, error) {
	s, err := util.UnmarshalYamlFile[VolumeState](filepath.Join(dir, "volume-state.yaml"))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (vs *VolumeServe) saveVolumeState(s *VolumeState) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	err = util.AtomicWriteFile(filepath.Join(vs.opts.Dir, "volume-state.yaml"), b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VolumeServe) updateVolumeState(ctx context.Context, sendStatus bool, fn func(s *VolumeState) error) error {
	if sendStatus {
		// we do it in defer so that the lock from below is actually unlocked
		defer func() {
			err := vs.refreshVolumeMount(ctx)
			if err != nil {
				vs.log.ErrorContext(ctx, "failed to refresh volume mount", "error", err)
			}
		}()
	}

	vs.volumeStateMutex.Lock()
	defer vs.volumeStateMutex.Unlock()

	s, err := vs.loadVolumeState()
	if err != nil {
		return err
	}
	err = fn(s)
	if err != nil {
		return err
	}
	err = vs.saveVolumeState(s)
	if err != nil {
		return err
	}
	return nil
}

func ListVolumeState(baseDir string) ([]*VolumeState, error) {
	des, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ret []*VolumeState
	for _, de := range des {
		if !de.IsDir() {
			continue
		}
		dir := filepath.Join(baseDir, de.Name())
		info, err := LoadVolumeState(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		ret = append(ret, info)
	}

	return ret, nil
}
