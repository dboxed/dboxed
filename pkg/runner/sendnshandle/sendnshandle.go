package sendnshandle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/vishvananda/netns"
)

func ListenSCMSocket(unixPath string) (*net.UnixListener, error) {
	err := syscall.Unlink(unixPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	addr, err := net.ResolveUnixAddr("unix", unixPath)
	if err != nil {
		return nil, err
	}

	ul, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}
	ul.SetUnlinkOnClose(true)

	return ul, nil
}

func readFD(uc *net.UnixConn) (int, error) {
	msg, oob := make([]byte, 2), make([]byte, 128)

	_, oobn, _, _, err := uc.ReadMsgUnix(msg, oob)
	if err != nil {
		return 0, err
	}

	cmsgs, err := syscall.ParseSocketControlMessage(oob[0:oobn])
	if err != nil {
		return 0, err
	} else if len(cmsgs) != 1 {
		return 0, errors.New("invalid number of cmsgs received")
	}

	fds, err := syscall.ParseUnixRights(&cmsgs[0])
	if err != nil {
		return 0, err
	} else if len(fds) != 1 {
		return 0, errors.New("invalid number of fds received")
	}
	return fds[0], nil
}

func ReadNetNsFD(ctx context.Context, unixPath string) (netns.NsHandle, error) {
	slog.InfoContext(ctx, "waiting for "+unixPath)
	for {
		_, err := os.Stat(unixPath)
		if err == nil {
			break
		}

		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	slog.InfoContext(ctx, "reading host netns handle")

	conn, err := net.Dial("unix", unixPath)
	if err != nil {
		return 0, fmt.Errorf("dial unix socket failed: %w", err)
	}
	defer conn.Close()

	uc, ok := conn.(*net.UnixConn)
	if !ok {
		return 0, fmt.Errorf("not a unix socket")
	}

	fd, err := readFD(uc)
	if err != nil {
		return 0, err
	}

	err = uc.Close()
	if err != nil {
		return 0, err
	}

	return netns.NsHandle(fd), nil
}

func SendNetNsFD(uc *net.UnixConn, fd netns.NsHandle) error {
	rights := syscall.UnixRights(int(fd))
	dummyByte := []byte{0}
	n, oobn, err := uc.WriteMsgUnix(dummyByte, rights, nil)
	if err != nil {
		return fmt.Errorf("WriteMsgUnix failed: %w", err)
	}
	if n != 1 || oobn != len(rights) {
		return fmt.Errorf("unexpected result from WriteMsgUnix: n=%d, oobn=%d", n, oobn)
	}
	return nil
}
