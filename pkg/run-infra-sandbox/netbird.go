package run_infra_sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"log/slog"
	"os"
	"path/filepath"
)

func (rn *RunInfraSandbox) startNetbirdServiceContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "starting netbird")

	image := rn.conf.BoxSpec.Netbird.Image
	if image == "" {
		image = fmt.Sprintf("netbirdio/netbird:%s", rn.conf.BoxSpec.Netbird.Version)
	}
	args := []string{
		"run",
		"-d",
		"--name", "unboxed-netbird",
		"--restart", "on-failure",
		"--net", fmt.Sprintf("ns:/run/netns/%s", rn.network.NamesAndIps.SandboxNamespaceName),
		"--dns", rn.conf.NetworkConfig.DnsProxyIP,
		"--privileged",
		"--entrypoint", "netbird",
		"-v/etc/unboxed:/etc/unboxed",
		"-e", "NB_FOREGROUND_MODE=false",
		image,
		"service", "run",
		"--log-file", "/dev/stdout",
	}

	_, err := rn.runNerdctl(ctx, false, args)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunInfraSandbox) runNetbirdUp(ctx context.Context) error {
	slog.InfoContext(ctx, "running netbird up")

	setupKeyFile := filepath.Join(types.UnboxedConfDir, "netbird-setup-key")

	err := os.WriteFile(setupKeyFile, []byte(rn.conf.BoxSpec.Netbird.SetupKey), 0600)
	if err != nil {
		return fmt.Errorf("failed to write netbird setup key: %w", err)
	}

	args := []string{
		"up",
		"--management-url", rn.conf.BoxSpec.Netbird.ManagementUrl,
		"--setup-key-file", setupKeyFile,
	}
	_, err = rn.runNetbirdCli(ctx, false, args...)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunInfraSandbox) runNetbirdCli(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	var args2 []string
	args2 = append(args2, "exec", "-i", "unboxed-netbird", "netbird")
	args2 = append(args2, args...)
	stdout, err := rn.runNerdctl(ctx, captureStdout, args2)
	if err != nil {
		return stdout, fmt.Errorf("netbird cli failed: %w", err)
	}
	return stdout, nil
}

func (rn *RunInfraSandbox) runNetbirdStatus(ctx context.Context) (*types.NetbirdStatus, error) {
	s, err := rn.runNetbirdCli(ctx, true, "status", "--json")
	if err != nil {
		return nil, err
	}

	var ret types.NetbirdStatus
	err = json.Unmarshal([]byte(s), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
