package systemd

import (
	"github.com/koobox/unboxed/cmd/unboxed/utils"
	"github.com/koobox/unboxed/pkg/systemd"
	"github.com/spf13/cobra"
)

var boxUrlArg string
var boxFileArg string
var boxNameArg string

var SystemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "sub-commands to control unboxed systemd integration",
}

var runSystemdInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd unit",
	Long:  ``,
	RunE:  runSystemdInstallNode,
}

func init() {
	SystemdCmd.AddCommand(runSystemdInstallCmd)

	runSystemdInstallCmd.Flags().StringVar(&boxUrlArg, "box-url", "", "box url")
	runSystemdInstallCmd.Flags().StringVar(&boxFileArg, "box-file", "", "box url")
	runSystemdInstallCmd.Flags().StringVar(&boxNameArg, "box-name", "", "box name")
	_ = runSystemdInstallCmd.MarkFlagRequired("box-name")
}

func runSystemdInstallNode(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	url, err := utils.GetBoxUrl(boxUrlArg, boxFileArg)
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		BoxUrl:  url,
		BoxName: boxNameArg,
	}

	return s.Run(ctx)
}
