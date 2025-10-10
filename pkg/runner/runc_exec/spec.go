package runc_exec

import (
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func createLibContainerRlimit(rlimit specs.POSIXRlimit) (configs.Rlimit, error) {
	rl, err := strToRlimit(rlimit.Type)
	if err != nil {
		return configs.Rlimit{}, err
	}
	return configs.Rlimit{
		Type: rl,
		Hard: rlimit.Hard,
		Soft: rlimit.Soft,
	}, nil
}
