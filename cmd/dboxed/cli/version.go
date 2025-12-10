package cli

import (
	"fmt"
	"os"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/version"
)

type VersionCmd struct {
}

func (cmd *VersionCmd) Run() error {
	_, _ = fmt.Fprintf(os.Stdout, "dboxed version %s\ninfra-image %s\n", version.Version, consts.GetDefaultSandboxInfraImage())
	return nil
}
