package box_specs

import (
	"fmt"
	"time"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

func AddNetbirdService(n2 dmodel.NetworkNetbird, box *dmodel.Box, composeProject *ctypes.Project) error {
	if box.Netbird.SetupKey == nil {
		return fmt.Errorf("box %d has no setup key", box.ID)
	}

	if composeProject.Services == nil {
		composeProject.Services = map[string]ctypes.ServiceConfig{}
	}
	if composeProject.Volumes == nil {
		composeProject.Volumes = map[string]ctypes.VolumeConfig{}
	}
	if composeProject.Configs == nil {
		composeProject.Configs = map[string]ctypes.ConfigObjConfig{}
	}
	if composeProject.Secrets == nil {
		composeProject.Secrets = map[string]ctypes.SecretConfig{}
	}

	composeProject.Volumes["netbird-client-socket"] = ctypes.VolumeConfig{
		Name: "netbird-client-socket",
	}
	composeProject.Volumes["netbird-client-config"] = ctypes.VolumeConfig{
		Name: "netbird-client-config",
	}
	composeProject.Volumes["netbird-client-ip"] = ctypes.VolumeConfig{
		Name: "netbird-client-ip",
	}

	composeProject.Configs["run-netbird-script"] = ctypes.ConfigObjConfig{
		Name: "run-netbird-script",
		Content: `
set -e
export NB_FOREGROUND_MODE=false
export NB_SETUP_KEY=$$(cat /setup-key)
export NB_MANAGEMENT_URL=` + n2.ApiUrl.V + `
export NB_HOSTNAME=` + box.Name + `
netbird service run &
NETBIRD_PID=$$!
netbird up
wait $$NETBIRD_PID
`,
	}
	composeProject.Configs["healthcheck-script"] = ctypes.ConfigObjConfig{
		Name: "healthcheck-script",
		Content: `
set -e
set -o pipefail
IP=$$(netbird status --ipv4)
echo IP=$$IP
echo $$IP > /netbird-ip/ip
`,
	}
	composeProject.Secrets["setup-key"] = ctypes.SecretConfig{
		Name:    "setup-key",
		Content: *box.Netbird.SetupKey,
	}

	composeProject.Services["netbird"] = ctypes.ServiceConfig{
		Name:    "netbird",
		Image:   fmt.Sprintf("netbirdio/netbird:%s", n2.NetbirdVersion.V),
		Restart: "unless-stopped",
		CapAdd: []string{
			"NET_ADMIN",
			"SYS_ADMIN",
			"SYS_RESOURCE",
		},
		NetworkMode: "host",
		Entrypoint: []string{
			"sh",
			"/run-netbird.sh",
		},
		HealthCheck: &ctypes.HealthCheckConfig{
			Test:     []string{"CMD", "sh", "/healthcheck.sh"},
			Interval: util.Ptr(ctypes.Duration(time.Second * 10)),
			Retries:  util.Ptr(uint64(1)),
		},
		Configs: []ctypes.ServiceConfigObjConfig{
			{
				Source: "run-netbird-script",
				Target: "/run-netbird.sh",
				Mode:   util.Ptr(ctypes.FileMode(0755)),
			},
			{
				Source: "healthcheck-script",
				Target: "/healthcheck.sh",
				Mode:   util.Ptr(ctypes.FileMode(0755)),
			},
		},
		Secrets: []ctypes.ServiceSecretConfig{
			{
				Source: "setup-key",
				Target: "/setup-key",
				Mode:   util.Ptr(ctypes.FileMode(0600)),
			},
		},
		Volumes: []ctypes.ServiceVolumeConfig{
			{
				Type:   "volume",
				Source: "netbird-client-config",
				Target: "/etc/netbird",
			},
			{
				Type:   "volume",
				Source: "netbird-client-ip",
				Target: "/netbird-ip",
			},
		},
	}

	return nil
}
