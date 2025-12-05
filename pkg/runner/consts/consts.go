package consts

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/version"
)

const dboxedInfraImage = "ghcr.io/dboxed/dboxed-infra"

const DboxedDataDir = "/var/lib/dboxed"
const DboxedMachineDataDir = DboxedDataDir + "/machine"
const BoxClientAuthFile = DboxedDataDir + "/client-auth.yaml"

const SandboxShortPrefix = "dbx"
const SandboxEnvironmentFile = DboxedDataDir + "/sandbox.env"
const NetworkConfFile = DboxedDataDir + "/network.yaml"
const HostResolvConfFile = DboxedDataDir + "/host-resolv.conf"
const SandboxDnsProxyIp = "127.1.0.53"
const SandboxDnsStaticMapFile = DboxedDataDir + "/dns-static-map.yaml"

const NetNsInitialUnixSocket = DboxedDataDir + "/netns-initial.socket"
const NetNsHolderUnixSocket = DboxedDataDir + "/netns-holder.socket"

const NetbirdDir = DboxedDataDir + "/netbird"

const ContainersDir = DboxedDataDir + "/containers"
const LogsDir = DboxedDataDir + "/logs"
const LogsTailDbFilename = "multitail.db"

const MachineLogsDir = DboxedMachineDataDir + "/logs"
const VolumesDir = DboxedDataDir + "/volumes"

const VethIPStoreFile = "veth-ip"
const SandboxInfoFile = "sandbox-info.yaml"

const ShutdownSandboxMarkerFile = DboxedDataDir + "/" + "stop-sandbox"

func GetDefaultInfraImage() string {
	tag := "nightly"
	if !version.IsDummyVersion() {
		tag = "v" + version.Version
	}
	infraImage := fmt.Sprintf("%s:%s", dboxedInfraImage, tag)
	return infraImage
}
