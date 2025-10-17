package commandutils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

func GetSandboxInfo(baseDir string, sandboxArg *string) (*sandbox.SandboxInfo, error) {
	if sandboxArg == nil {
		return nil, fmt.Errorf("missing sandbox arg")
	}

	sandboxes, err := sandbox.ListSandboxes(baseDir)
	if err != nil {
		return nil, err
	}

	sandboxId, err := strconv.ParseInt(*sandboxArg, 10, 64)
	if err != nil {
		sandboxId = -1
	}

	for _, s := range sandboxes {
		if s.SandboxName == *sandboxArg || s.Box.Name == *sandboxArg || s.Box.ID == sandboxId || s.Box.Uuid == *sandboxArg {
			return &s, nil
		}
	}
	return nil, os.ErrNotExist
}

func GetOneOrAllSandboxInfos(baseDir string, sandboxArg *string, all bool) ([]sandbox.SandboxInfo, error) {
	var sandboxes []sandbox.SandboxInfo
	if sandboxArg != nil {
		si, err := GetSandboxInfo(baseDir, sandboxArg)
		if err != nil {
			return nil, err
		}
		sandboxes = append(sandboxes, *si)
	} else if all {
		var err error
		sandboxes, err = sandbox.ListSandboxes(baseDir)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("you must either specify a sandbox name or use --all")
	}
	return sandboxes, nil
}
