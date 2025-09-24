//go:build linux

package volume

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/losetup"
	"github.com/dboxed/dboxed/pkg/volume/lvm"
	"github.com/dboxed/dboxed/pkg/volume/volume"
	"github.com/moby/sys/mountinfo"
)

type CleanupLoopDevs struct {
}

var onlyDigits = regexp.MustCompile("^[0-9]*$")

func (cmd *CleanupLoopDevs) Run(g *flags.GlobalFlags) error {
	usedRefs, err := cmd.findUsedLoopRefs()
	if err != nil {
		return err
	}

	loInfos, err := volume.ListLoopDevs()
	if err != nil {
		return err
	}

	pvs, err := lvm.ListPVs()
	if err != nil {
		return err
	}

	for _, li := range loInfos {
		loDev := losetup.New(uint64(li.Number), 0)
		fname := string(li.FileName[:])
		fname = strings.TrimRight(fname, "\000")
		if !strings.HasPrefix(fname, volume.RefPrefix+"-") {
			continue
		}

		log := slog.With(slog.Any("loopDev", loDev.Path()), slog.Any("ref", fname))

		if nsId, ok := usedRefs[fname]; ok {
			log.Info("loop dev is in use", slog.Any("mountNamespace", nsId))
			continue
		}

		log.Info("loop dev is orphaned, trying to find physical volumes and volume groups to deactivate")
		for _, pv := range pvs {
			if pv.PvName == loDev.Path() {
				err = volume.DeactivateVolume(pv.VgName)
				if err != nil {
					log.Warn("error in VGDeactivate", slog.Any("loDev", loDev.Path()), slog.Any("vgName", pv.VgName))
				}
			}
		}
	}

	return nil
}

func (cmd *CleanupLoopDevs) findUsedLoopRefs() (map[string]int64, error) {
	des, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	handledNamespaces := map[int64]struct{}{}
	foundRefs := map[string]int64{}
	for _, de := range des {
		if !de.IsDir() {
			continue
		}

		if !onlyDigits.MatchString(de.Name()) {
			continue
		}

		procDir := filepath.Join("/proc", de.Name())
		mntNsFile := filepath.Join(procDir, "ns/mnt")
		mntLnk, err := os.Readlink(mntNsFile)
		if err != nil {
			continue
		}

		var nsId int64
		_, err = fmt.Sscanf(mntLnk, "mnt:[%d]", &nsId)
		if err != nil {
			continue
		}

		if _, ok := handledNamespaces[nsId]; ok {
			continue
		}
		handledNamespaces[nsId] = struct{}{}

		mountInfoBytes, err := os.ReadFile(filepath.Join(procDir, "mountinfo"))
		if err != nil {
			continue
		}
		var mounts []*mountinfo.Info
		mounts, err = mountinfo.GetMountsFromReader(bytes.NewReader(mountInfoBytes), nil)
		if err != nil {
			continue
		}

		for _, m := range mounts {
			if m.FSType != "tmpfs" {
				continue
			}
			ref, err := volume.ReadLoopRef(filepath.Join(procDir, "root", m.Mountpoint))
			if err != nil {
				continue
			}
			foundRefs[ref] = nsId
		}
	}
	return foundRefs, nil
}
