//go:build linux

package sandbox

import (
	"context"

	"github.com/dboxed/dboxed/pkg/runner/run-in-sandbox"
)

type RunInSandbox struct {
}

func (cmd *RunInSandbox) Run() error {
	ctx := context.Background()

	runBox := run_in_sandbox.RunInSandbox{}

	err := runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
