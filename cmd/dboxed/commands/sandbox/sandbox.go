//go:build linux

package sandbox

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type SandboxCommands struct {
	Run    RunCmd    `cmd:"" help:"Run a box inside a sandbox"`
	List   ListCmd   `cmd:"" help:"List sandboxes"`
	Kill   KillCmd   `cmd:"" help:"Kill a sandbox"`
	Remove RemoveCmd `cmd:"" help:"Remove a sandbox" aliases:"delete,rm"`

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
