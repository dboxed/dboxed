package consts

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/version"
)

const dboxedInfraSandboxImage = "ghcr.io/dboxed/dboxed-infra-sandbox"
const dboxedInfraVolumeImage = "ghcr.io/dboxed/dboxed-infra-volume"

const DboxedDataDir = "/var/lib/dboxed"
const SandboxClientAuthFile = DboxedDataDir + "/client-auth.yaml"

const SandboxShortPrefix = "dbx"
const SandboxEnvironmentFile = DboxedDataDir + "/sandbox.env"
const NetworkConfFile = DboxedDataDir + "/network.yaml"
const HostResolvConfFile = DboxedDataDir + "/host-resolv.conf"
const SandboxDnsProxyIp = "127.1.0.53"
const SandboxDnsStaticMapFile = DboxedDataDir + "/dns-static-map.yaml"

const NetNsInitialUnixSocket = DboxedDataDir + "/netns-initial.socket"
const NetNsHolderUnixSocket = DboxedDataDir + "/netns-holder.socket"

const NetbirdDir = DboxedDataDir + "/netbird"

const LogsDir = DboxedDataDir + "/logs"
const LogsTailDbFilename = "multitail.db"

const VolumesDir = DboxedDataDir + "/volumes"

const VethIPStoreFile = "veth-ip"
const SandboxInfoFile = "sandbox-info.yaml"

const ShutdownSandboxMarkerFile = DboxedDataDir + "/" + "stop-sandbox"

func GetDefaultInfraImageTag() string {
	tag := "nightly"
	if !version.IsDummyVersion() {
		tag = "v" + version.Version
	}
	return tag
}

func GetDefaultSandboxInfraImage() string {
	return fmt.Sprintf("%s:%s", dboxedInfraSandboxImage, GetDefaultInfraImageTag())
}

func GetDefaultVolumeInfraImage() string {
	return fmt.Sprintf("%s:%s", dboxedInfraVolumeImage, GetDefaultInfraImageTag())
}
