package volume

import (
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dboxed/dboxed/pkg/volume/losetup"
)

var loRegex = regexp.MustCompile("^loop([0-9]*)$")

func ListLoopDevs() ([]losetup.Info, error) {
	des, err := os.ReadDir("/dev")
	if err != nil {
		return nil, err
	}

	var ret []losetup.Info
	for _, de := range des {
		m := loRegex.FindStringSubmatch(de.Name())
		if m == nil {
			continue
		}
		idxStr := m[1]
		idx, _ := strconv.ParseInt(idxStr, 10, 32)
		loDev := losetup.New(uint64(idx), 0)
		li, err := loDev.GetInfo()
		if err != nil {
			continue
		}
		ret = append(ret, li)
	}
	return ret, nil
}

// AttachLoopDev will attach the loop device, open a handle to it and then immediately detach the loop device
// The loop device will be kept alive as long as the file handle is open. The handle is returned by AttachLoopDev
// and meant to be closed by the caller.
func AttachLoopDev(image string, lockId string) (*losetup.Device, io.Closer, error) {
	ref := BuildRef(lockId)
	ref += strings.Repeat("\000", 64-len(ref))

	loDev, err := losetup.Attach(image, 0, false)
	if err != nil {
		return nil, nil, err
	}
	defer loDev.Detach()
	f, err := os.Open(loDev.Path())
	if err != nil {
		return nil, nil, err
	}
	doClose := true
	defer func() {
		if doClose {
			_ = f.Close()
		}
	}()

	loDevInfo, err := loDev.GetInfo()
	if err != nil {
		return nil, nil, err
	}
	copy(loDevInfo.FileName[:], []byte(ref))
	err = loDev.SetInfo(loDevInfo)
	if err != nil {
		return nil, nil, err
	}
	doClose = false
	return &loDev, f, nil
}

func GetLoopDev(image string, lockId string) (*losetup.Device, io.Closer, error) {
	ref := BuildRef(lockId)

	loInfos, err := ListLoopDevs()
	if err != nil {
		return nil, nil, err
	}
	for _, li := range loInfos {
		fname := string(li.FileName[:])
		fname = strings.TrimRight(fname, "\000")
		if fname == ref {
			d := losetup.New(uint64(li.Number), 0)
			handle, err := os.Open(d.Path())
			if err != nil {
				return nil, nil, err
			}
			return &d, handle, nil
		}
	}

	return nil, nil, os.ErrNotExist
}

func GetOrAttachLoopDev(image string, lockId string) (*losetup.Device, io.Closer, error) {
	d, handle, err := GetLoopDev(image, lockId)
	if err != nil {
		if os.IsNotExist(err) {
			return AttachLoopDev(image, lockId)
		}
		return nil, nil, err
	}
	return d, handle, nil
}
