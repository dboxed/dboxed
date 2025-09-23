package box_specs

import (
	"fmt"
	"time"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

func AddNetbirdService(n2 dmodel.NetworkNetbird, box *dmodel.Box, composeProject *ctypes.Project) (*boxspec.BoxVolumeSpec, error) {
	if box.Netbird.SetupKey == nil {
		return nil, fmt.Errorf("box %d has no setup key", box.ID)
	}

	if composeProject.Services == nil {
		composeProject.Services = map[string]ctypes.ServiceConfig{}
	}
	if composeProject.Volumes == nil {
		composeProject.Volumes = map[string]ctypes.VolumeConfig{}
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

	// this bundle also contains the setup key, so it should be considered a secret
	scriptVolume := &boxspec.BoxVolumeSpec{
		Name:     "netbird-scripts",
		RootMode: "0700",
		FileBundle: &boxspec.FileBundle{
			Files: []boxspec.FileBundleEntry{
				{
					Path:       "setup-key",
					Mode:       "0700",
					StringData: *box.Netbird.SetupKey,
				},
				{
					Path: "run-netbird.sh",
					Mode: "0700",
					StringData: `
set -e
export NB_FOREGROUND_MODE=false
export NB_SETUP_KEY=$(cat /scripts/setup-key)
export NB_MANAGEMENT_URL=` + n2.ApiUrl.V + `
export NB_HOSTNAME=` + box.Name + `
netbird service run &
NETBIRD_PID=$!
netbird up
wait $NETBIRD_PID
`,
				},
				{
					Path: "healthcheck.sh",
					Mode: "0700",
					StringData: `
set -e
set -o pipefail
IP=$(netbird status --ipv4)
echo IP=$IP
echo $IP > /netbird-ip/ip
`,
				},
			},
		},
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
			"/scripts/run-netbird.sh",
		},
		HealthCheck: &ctypes.HealthCheckConfig{
			Test:     []string{"CMD", "sh", "/scripts/healthcheck.sh"},
			Interval: util.Ptr(ctypes.Duration(time.Second * 10)),
			Retries:  util.Ptr(uint64(1)),
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
			{
				Type:   "dboxed",
				Source: scriptVolume.Name,
				Target: "/scripts",
			},
		},
	}

	return scriptVolume, nil
}
