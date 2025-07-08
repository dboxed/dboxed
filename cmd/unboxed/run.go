package main

import (
	"context"
	"github.com/koobox/unboxed/pkg/run-box"
	"github.com/spf13/cobra"
	"net"
	"time"
)

var runWorkDirArg string

var runBoxUrlArg string
var runBoxFileArg string
var runBoxNameArg string
var runVethCidrArgNode string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Download, unpack and run the a box",
	Long:  ``,
	RunE:  doRun,
}

func init() {
	runCmd.Flags().StringVar(&runWorkDirArg, "work-dir", "/var/lib/unboxed", "unboxed work dir")

	runCmd.Flags().StringVar(&runBoxUrlArg, "box-url", "", "box url")
	runCmd.Flags().StringVar(&runBoxFileArg, "box-file", "", "box url")
	runCmd.Flags().StringVar(&runBoxNameArg, "box-name", "", "box name")
	_ = runCmd.MarkFlagRequired("box-name")

	runCmd.Flags().StringVar(&runVethCidrArgNode, "veth-cidr", "1.2.3.0/30", "veth cidr")

	rootCmd.AddCommand(runCmd)
}

func doRun(cmd *cobra.Command, args []string) error {
	url, err := getBoxUrl(runBoxUrlArg, runBoxFileArg)
	if err != nil {
		return err
	}

	_, vethCidr, err := net.ParseCIDR(runVethCidrArgNode)
	if err != nil {
		return err
	}

	runNode := run_box.RunBox{
		BoxUrl:          url,
		BoxName:         runBoxNameArg,
		WorkDir:         runWorkDirArg,
		VethNetworkCidr: vethCidr,
	}

	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = runNode.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
