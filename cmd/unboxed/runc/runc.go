package runc

import (
	"github.com/koobox/unboxed/cmd/unboxed/utils"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var RuncCmd = &cobra.Command{
	Use:                "runc",
	Short:              "run runc for a box",
	Long:               ``,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runRunc,
}

func init() {
}

func runRunc(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	boxName := args[0]
	args = args[1:]

	sandboxDir := filepath.Join(utils.DefaultWorkdir, "boxes", boxName)

	c, err := sandbox.BuildRuncCmd(ctx, sandboxDir, args...)
	if err != nil {
		return err
	}
	c.Stdin = os.Stdin
	
	return c.Run()
}
