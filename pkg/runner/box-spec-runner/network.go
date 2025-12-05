package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *BoxSpecRunner) reconcileNetwork(ctx context.Context) error {
	s6 := &service.S6Helper{
		Log: slog.Default(),
	}

	if rn.BoxSpec.Network == nil {
		err := s6.S6SvcDown(ctx, "netbird")
		if err != nil {
			return err
		}
		err = s6.SetDownMarker("netbird", true)
		if err != nil {
			return err
		}
		return nil
	}

	if rn.BoxSpec.Network.Netbird != nil {
		err := rn.writeNetbirdConfigs()
		if err != nil {
			return err
		}
		err = s6.SetDownMarker("netbird", false)
		if err != nil {
			return err
		}
		err = s6.S6SvcUp(ctx, "netbird")
		if err != nil {
			return err
		}
		err = rn.runNetbirdUp(ctx)
		if err != nil {
			return err
		}
	}

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

func (rn *BoxSpecRunner) writeNetbirdConfigs() error {
	envFile := ""
	if rn.BoxSpec.Network.Netbird.ManagementUrl != "" {
		envFile += fmt.Sprintf("export NETBIRD_MANAGEMENT_URL=%s\n", rn.BoxSpec.Network.Netbird.ManagementUrl)
	}

	err := os.WriteFile(filepath.Join(consts.NetbirdDir, "netbird.env"), []byte(envFile), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(consts.NetbirdDir, "netbird.setup-key"), []byte(rn.BoxSpec.Network.Netbird.SetupKey), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) runNetbirdUp(ctx context.Context) error {
	cmd := util.CommandHelper{
		Command: "netbird",
		Args: []string{
			"up",
			"--disable-dns",
			"--hostname", rn.BoxSpec.Network.Netbird.Hostname,
			"--extra-dns-labels", fmt.Sprintf("%s.%s", rn.BoxSpec.Name, *rn.BoxSpec.Network.Name),
			"--setup-key-file", filepath.Join(consts.NetbirdDir, "netbird.setup-key"),
		},
		Logger: slog.Default(),
		LogCmd: true,
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}
