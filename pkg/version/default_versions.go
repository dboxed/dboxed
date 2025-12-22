package version

import "fmt"

const dboxedInfraSandboxImage = "ghcr.io/dboxed/dboxed-infra-sandbox"
const dboxedInfraVolumeImage = "ghcr.io/dboxed/dboxed-infra-volume"

func GetDefaultInfraImageTag() string {
	tag := "nightly"
	if !IsDummyVersion() {
		tag = "v" + Version
	}
	return tag
}

func GetDefaultSandboxInfraImage() string {
	return fmt.Sprintf("%s:%s", dboxedInfraSandboxImage, GetDefaultInfraImageTag())
}

func GetDefaultVolumeInfraImage() string {
	return fmt.Sprintf("%s:%s", dboxedInfraVolumeImage, GetDefaultInfraImageTag())
}

func GetDefaultMachineDboxedVersion() string {
	if IsDummyVersion() {
		return "nightly"
	}
	return Version
}
