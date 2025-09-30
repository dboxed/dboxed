package consts

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/version"
)

const dboxedInfraImage = "ghcr.io/dboxed/dboxed-infra"

const DboxedDataDir = "/var/lib/dboxed"

const BoxSpecFile = DboxedDataDir + "/box-spec.yaml"
const NetworkConfFile = DboxedDataDir + "/network.yaml"
const BoxClientAuthFile = DboxedDataDir + "/client-auth.yaml"

const ContainersDir = DboxedDataDir + "/containers"
const LogsDir = DboxedDataDir + "/logs"
const LogsTailDbFilename = "multitail.db"

const VolumesDir = DboxedDataDir + "/volumes"

const VethIPStoreFile = "veth-ip"
const BoxSpecUuidFile = "box-spec-uuid"

func GetDefaultInfraImage() string {
	tag := "nightly"
	if !version.IsDummyVersion() {
		tag = version.Version
	}
	infraImage := fmt.Sprintf("%s:%s", dboxedInfraImage, tag)
	return infraImage
}
