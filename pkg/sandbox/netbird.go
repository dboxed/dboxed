package sandbox

import (
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
)

func (rn *Sandbox) buildNetbirdContainerSpec() (types.ContainerSpec, error) {
	if rn.BoxSpec.Netbird.Image == "" && rn.BoxSpec.Netbird.Version == "" {
		return types.ContainerSpec{}, fmt.Errorf("one of image or version must be specified for netbird")
	}

	image := rn.BoxSpec.Netbird.Image
	if image == "" {
		image = fmt.Sprintf("netbirdio/netbird:%s", rn.BoxSpec.Netbird.Version)
	}

	var entrypoint []string
	entrypoint = append(entrypoint, "netbird", "service", "run")
	entrypoint = append(entrypoint, "--log-file", "/dev/stdout")

	var env []string
	env = append(env, "NB_FOREGROUND_MODE=false")

	s := types.ContainerSpec{
		Name:       "_netbird",
		Image:      image,
		Entrypoint: entrypoint,
		Env:        env,
		BindMounts: []types.BindMount{
			{HostPath: rn.getSharedDirOnHost(), ContainerPath: types.SharedDir},
		},
		Privileged: true,
	}
	return s, nil
}
