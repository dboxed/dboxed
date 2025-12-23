package consts

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
const SandboxStatusFile = DboxedDataDir + "/sandbox-status.yaml"

const VolumesDir = DboxedDataDir + "/volumes"

const VethIPStoreFile = "veth-ip"
const SandboxInfoFile = "sandbox-info.yaml"

const ShutdownSandboxMarkerFile = DboxedDataDir + "/" + "stop-sandbox"
