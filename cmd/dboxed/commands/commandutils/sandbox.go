package commandutils

import (
	"fmt"
	"os"

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

	for _, s := range sandboxes {
		if s.SandboxName == *sandboxArg || s.Box.Name == *sandboxArg || s.Box.ID == *sandboxArg {
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
