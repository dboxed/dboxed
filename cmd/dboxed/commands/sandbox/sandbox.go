//go:build linux

package sandbox

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type SandboxCommands struct {
	Run  RunCmd  `cmd:"" help:"Run a box"`
	Kill KillCmd `cmd:"" help:"Kill a box"`

	Runc RuncCmd `cmd:"" help:"Run runc for a box"`

	RunInSandbox RunInSandbox `cmd:"" hidden:""`
}

func GetSandboxName(box *models.Box, sandboxName *string) (string, error) {
	var ret string
	if sandboxName != nil {
		err := util.CheckName(*sandboxName)
		if err != nil {
			return "", err
		}
		ret = *sandboxName
	} else {
		ret = fmt.Sprintf("%s-%s", box.Name, box.Uuid)
	}
	return ret, nil
}
