package types

const UnboxedInfraImage = "ghcr.io/koobox/unboxed-infra"

const UnboxedConfDir = "/etc/unboxed"
const UnboxedDataDir = "/var/lib/unboxed"

const LogsDir = UnboxedDataDir + "/logs"

const InfraConfFile = UnboxedConfDir + "/infra-conf.json"
const LogsTailDbFilename = "tail.db"
const InfraHostReadyMarkerFile = UnboxedConfDir + "/infra-host-ready"
const DnsMapFile = UnboxedConfDir + "/dns-map.json"

const VethIPStoreFile = "veth-ip"
