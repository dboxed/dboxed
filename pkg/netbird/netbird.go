package netbird

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/util"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync/atomic"
	"time"
)

type RunNetbird struct {
	NetbirdManagementUrl string
	NetbirdSetupKey      string
	NetbirdPeerName      string

	status atomic.Pointer[NetbirdStatus]

	logNetbird io.WriteCloser
}

func (rn *RunNetbird) Start(ctx context.Context) error {
	var err error

	rn.logNetbird, err = logs.BuildRotatingLogger("netbird")
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "starting netbird")

	go func() {
		for {
			err := rn.runNetbirdService(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "error while running netbird service", slog.Any("error", err))
			}
			if !util.SleepWithContext(ctx, 5*time.Second) {
				return
			}
		}
	}()

	err = rn.runNetbirdUp(ctx)
	if err != nil {
		return err
	}

	go rn.runNetbirdStatusLoop(ctx)

	slog.InfoContext(ctx, "waiting for initial netbird status")
	for {
		s := rn.status.Load()
		if s != nil {
			break
		}
		if !util.SleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}

	return nil
}

func (rn *RunNetbird) runNetbirdService(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "netbird", "service", "run",
		"--hostname", rn.NetbirdPeerName,
		"--log-file", "/dev/stderr",
		"--log-level", "debug",
	)

	cmd.Stdout = rn.logNetbird
	cmd.Stderr = rn.logNetbird

	return cmd.Run()
}

func (rn *RunNetbird) runNetbirdStatus(ctx context.Context) (*NetbirdStatus, error) {
	_, _ = fmt.Fprintf(rn.logNetbird, "running netbird status\n")

	if _, err := os.Stat("/etc/netbird/config.json"); err != nil {
		// running netbird status before the config is available will result in an invalid config
		return nil, fmt.Errorf("netbird config.json not present yet")
	}

	buf := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "netbird", "status", "--json")
	cmd.Stdout = buf
	cmd.Stderr = rn.logNetbird

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("netbird status failed: %w", err)
	}

	_, _ = fmt.Fprintf(rn.logNetbird, "netbird status returned: %s\n", buf.String())

	var ret NetbirdStatus
	err = json.Unmarshal(buf.Bytes(), &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse netbird status: %w", err)
	}

	return &ret, nil
}

func (rn *RunNetbird) runNetbirdStatusLoop(ctx context.Context) {
	slog.InfoContext(ctx, "waiting for netbird service to become ready")
	for {
		status, err := rn.runNetbirdStatus(ctx)
		if err == nil {
			rn.status.Store(status)
		}
		if !util.SleepWithContext(ctx, 5*time.Second) {
			return
		}
	}
}

func (rn *RunNetbird) runNetbirdUp(ctx context.Context) error {
	_, _ = fmt.Fprintf(rn.logNetbird, "running netbird up\n")
	slog.InfoContext(ctx, "running netbird up")

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(rn.NetbirdSetupKey)
	if err != nil {
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "netbird",
		"up",
		"--management-url", rn.NetbirdManagementUrl,
		"--admin-url", rn.NetbirdManagementUrl,
		"--setup-key-file", tmpFile.Name(),
		"--hostname", rn.NetbirdPeerName,
	)
	cmd.Stdout = rn.logNetbird
	cmd.Stderr = rn.logNetbird
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunNetbird) GetLastStatus() NetbirdStatus {
	s := rn.status.Load()
	return *s
}
