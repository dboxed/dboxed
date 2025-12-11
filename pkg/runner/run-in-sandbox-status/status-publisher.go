package run_in_sandbox_status

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/klauspost/compress/gzip"
)

type StatusPublisher struct {
	Client *baseclient.Client
	BoxId  string

	stopCh         chan struct{}
	sendStatusDone sync.WaitGroup

	sandboxStatusSent *models.UpdateBoxSandboxStatus2
	sandboxStatusTime time.Time
	dockerPSSent      []byte
}

func (rn *StatusPublisher) Start(ctx context.Context) {
	rn.sendSandboxStatusDockerPs(ctx)
	rn.sendSandboxStatus(ctx)

	rn.stopCh = make(chan struct{})
	rn.sendStatusDone.Add(1)
	go func() {
		defer rn.sendStatusDone.Done()
		for {
			stop := false
			select {
			case <-time.After(time.Second * 5):
			case <-rn.stopCh:
				stop = true
			case <-ctx.Done():
				return
			}

			rn.sendSandboxStatusDockerPs(ctx)
			rn.sendSandboxStatus(ctx)
			if stop {
				return
			}
		}
	}()
}

func (rn *StatusPublisher) Stop() {
	if rn.stopCh != nil {
		close(rn.stopCh)
		rn.sendStatusDone.Wait()
	}
}

func (rn *StatusPublisher) sendSandboxStatus(ctx context.Context) {
	s, err := util.UnmarshalYamlFile[models.UpdateBoxSandboxStatus2](consts.SandboxStatusFile)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.ErrorContext(ctx, "error while reading sandbox status", "error", err)
		}
		return
	}

	networkIp, err := rn.readNetworkIp()
	if err != nil && !os.IsNotExist(err) {
		slog.ErrorContext(ctx, "error while reading network ip", "error", err)
		return
	}
	s.NetworkIp4 = networkIp

	if !time.Now().After(rn.sandboxStatusTime.Add(time.Second*30)) && util.EqualsViaJson(s, rn.sandboxStatusSent) {
		return
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err = boxesClient.UpdateSandboxStatus(ctx, rn.BoxId, models.UpdateBoxSandboxStatus{
		SandboxStatus: s,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to publish sandbox run status", "error", err)
	} else {
		rn.sandboxStatusSent = s
		rn.sandboxStatusTime = time.Now()
	}
}

func (rn *StatusPublisher) sendSandboxStatusDockerPs(ctx context.Context) {
	b, err := rn.runDockerPS(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run docker ps", "error", err)
		return
	}

	if bytes.Equal(b, rn.dockerPSSent) {
		return
	}

	err = rn.doSendSandboxStatusDockerPs(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "failed to report docker ps result", "error", err)
	} else {
		rn.dockerPSSent = b
	}
	return
}

func (rn *StatusPublisher) doSendSandboxStatusDockerPs(ctx context.Context, b []byte) error {
	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	w := gzip.NewWriter(buf)
	_, err := w.Write(b)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	boxesClient := clients.BoxClient{Client: rn.Client}
	err = boxesClient.UpdateSandboxStatus(ctx, rn.BoxId, models.UpdateBoxSandboxStatus{
		DockerPs: buf.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (rn *StatusPublisher) runDockerPS(ctx context.Context) ([]byte, error) {
	c := command_helper.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "-a", "--format", "json"},
	}
	b, err := c.RunStdout(ctx)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (rn *StatusPublisher) readNetworkIp() (*string, error) {
	jsonPath := filepath.Join(consts.NetbirdDir, "status.json")
	status, err := util.UnmarshalYamlFile[models.NetbirdPeerStatus](jsonPath)
	if err != nil {
		return nil, err
	}

	ip, _, err := net.ParseCIDR(status.NetbirdIp)
	if err != nil {
		return nil, err
	}

	return util.Ptr(ip.String()), nil
}
