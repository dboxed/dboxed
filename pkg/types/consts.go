package types

const DboxedInfraImage = "ghcr.io/dboxed/dboxed-infra"

const DboxedConfDir = "/etc/dboxed"
const DboxedDataDir = "/var/lib/dboxed"

const LogsDir = DboxedDataDir + "/logs"

const InfraConfFile = DboxedConfDir + "/infra-conf.json"
const LogsTailDbFilename = "multitail.db"
const InfraHostReadyMarkerFile = DboxedConfDir + "/infra-host-ready"
const DnsMapFile = DboxedConfDir + "/dns-map.json"

const VethIPStoreFile = "veth-ip"
