package box_spec_runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *BoxSpecRunner) reconcileNetwork(ctx context.Context) error {
	s6 := &service.S6Helper{
		Log: rn.Log,
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
			"--setup-key-file", filepath.Join(consts.NetbirdDir, "netbird.setup-key"),
		},
		Logger: rn.Log,
		LogCmd: true,
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}
