package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/runner/compose"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *BoxSpecRunner) reconcileNetwork(ctx context.Context) error {
	err := rn.reconcileStaticHosts(ctx)
	if err != nil {
		return err
	}

	p := &ctypes.Project{
		Name:     "dboxed-network",
		Services: map[string]ctypes.ServiceConfig{},
		Volumes:  map[string]ctypes.VolumeConfig{},
	}

	ch := &compose.ComposeHelper{
		BaseDir: rn.composeBaseDir,
		Project: p,
	}

	if rn.BoxSpec.Network.Netbird != nil {
		err := rn.buildNetbirdComposeProject(ctx, p)
		if err != nil {
			return err
		}
	}

	err = ch.RunPull(ctx)
	if err != nil {
		return err
	}
	err = ch.RunUp(ctx, true)
	if err != nil {
		return err
	}

	if rn.BoxSpec.Network.Netbird != nil {
		err = rn.runNetbirdUp(ctx, ch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) reconcileStaticHosts(ctx context.Context) error {
	staticHosts := map[string]string{}
	for _, nh := range rn.BoxSpec.Network.NetworkHosts {
		fqdn := fmt.Sprintf("%s.dboxed.", nh.Name)
		staticHosts[fqdn] = nh.IP4
	}
	slog.InfoContext(ctx, "writing new static hosts map", "map", staticHosts)
	err := util.AtomicWriteFileYaml(consts.SandboxDnsStaticMapFile, staticHosts, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) buildNetbirdComposeProject(ctx context.Context, p *ctypes.Project) error {
	err := rn.writeNetbirdConfigs()
	if err != nil {
		return err
	}

	cmd := []string{
		"netbird",
		"service",
		"run",
		"--log-file=console",
	}
	if rn.BoxSpec.Network.Netbird.ManagementUrl != "" {
		cmd = append(cmd, "--management-url", rn.BoxSpec.Network.Netbird.ManagementUrl)
	}
	if rn.BoxSpec.Network.Netbird.Hostname != "" {
		cmd = append(cmd, "--hostname", rn.BoxSpec.Network.Netbird.Hostname)
	}

	p.Volumes["run"] = ctypes.VolumeConfig{}

	image := "netbirdio/netbird:" + rn.BoxSpec.Network.Netbird.Version
	dataVolume := ctypes.ServiceVolumeConfig{
		Type:   "bind",
		Source: consts.NetbirdDir,
		Target: "/var/lib/netbird",
	}
	runVolume := ctypes.ServiceVolumeConfig{
		Type:   "volume",
		Source: "run",
		Target: "/run",
	}

	p.Services["netbird"] = ctypes.ServiceConfig{
		Image:       image,
		NetworkMode: "host",
		CapAdd: []string{
			"NET_ADMIN",
			"SYS_ADMIN",
			"SYS_RESOURCE",
		},
		Restart:    "unless-stopped",
		Entrypoint: cmd,
		EnvFiles: []ctypes.EnvFile{
			{
				Path: filepath.Join(consts.NetbirdDir, "netbird.env"),
			},
		},
		Volumes: []ctypes.ServiceVolumeConfig{dataVolume, runVolume},
	}
	p.Services["netbird-status"] = ctypes.ServiceConfig{
		Image:   image,
		Restart: "unless-stopped",
		Entrypoint: []string{
			"sh",
			"-c",
			`
trap exit TERM
while true; do
  if [ -e /var/run/netbird.sock ]; then
    if netbird status --json > /var/lib/netbird/status.json.tmp; then
      mv /var/lib/netbird/status.json.tmp /var/lib/netbird/status.json
    else
      rm -f /var/lib/netbird/status.json
    fi
  fi
  sleep 10
done
`,
		},
		Volumes: []ctypes.ServiceVolumeConfig{dataVolume, runVolume},
	}

	return nil
}

func (rn *BoxSpecRunner) writeNetbirdConfigs() error {
	err := os.WriteFile(filepath.Join(consts.NetbirdDir, "netbird.setup-key"), []byte(rn.BoxSpec.Network.Netbird.SetupKey), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) runNetbirdUp(ctx context.Context, ch *compose.ComposeHelper) error {
	args := []string{
		"netbird",
		"up",
		"--disable-dns",
		"--hostname", rn.BoxSpec.Network.Netbird.Hostname,
		"--extra-dns-labels", fmt.Sprintf("%s.%s", rn.BoxSpec.Name, *rn.BoxSpec.Network.Name),
		"--setup-key-file", "/var/lib/netbird/netbird.setup-key",
	}

	err := ch.RunExec(ctx, "netbird", true, args...)
	if err != nil {
		return err
	}
	return nil
}
