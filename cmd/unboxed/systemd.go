package main

import (
	"github.com/koobox/unboxed/pkg/systemd"
	"github.com/spf13/cobra"
)

var systemdInstallBoxUrlArg string
var systemdInstallBoxFileArg string
var systemdInstallBoxNameArg string

var cmdSystemd = &cobra.Command{
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
	rootCmd.AddCommand(cmdSystemd)
	cmdSystemd.AddCommand(runSystemdInstallCmd)

	runSystemdInstallCmd.Flags().StringVar(&systemdInstallBoxUrlArg, "box-url", "", "box url")
	runSystemdInstallCmd.Flags().StringVar(&systemdInstallBoxFileArg, "box-file", "", "box url")
	runSystemdInstallCmd.Flags().StringVar(&systemdInstallBoxNameArg, "box-name", "", "box name")
	_ = runSystemdInstallCmd.MarkFlagRequired("box-name")
}

func runSystemdInstallNode(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	url, err := getBoxUrl(systemdInstallBoxUrlArg, systemdInstallBoxFileArg)
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		BoxUrl:  url,
		BoxName: systemdInstallBoxNameArg,
	}

	return s.Run(ctx)
}
