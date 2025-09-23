package cli

import "github.com/dboxed/dboxed/cmd/dboxed/commands/runner"

type cliOnlyLinux struct {
	Runner runner.RunnerCommands `cmd:"" help:"runner commands"`
}
