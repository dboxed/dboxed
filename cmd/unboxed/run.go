package main

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/run-box"
	"github.com/spf13/cobra"
	"net"
	"time"
)

var workDirArg string

var boxUrlArg string
var boxFileArg string
var boxNameArg string
var vethCidrArgNode string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Download, unpack and run the a box",
	Long:  ``,
	RunE:  doRun,
}

func init() {
	runCmd.Flags().StringVar(&workDirArg, "work-dir", "/var/lib/unboxed", "unboxed work dir")

	runCmd.Flags().StringVar(&boxUrlArg, "box-url", "", "box url")
	runCmd.Flags().StringVar(&boxFileArg, "box-file", "", "box url")
	runCmd.Flags().StringVar(&boxNameArg, "box-name", "", "box name")
	_ = runCmd.MarkFlagRequired("box-name")

	runCmd.Flags().StringVar(&vethCidrArgNode, "veth-cidr", "1.2.3.0/30", "veth cidr")

	rootCmd.AddCommand(runCmd)
}

func doRun(cmd *cobra.Command, args []string) error {
	if boxUrlArg == "" && boxFileArg == "" {
		return fmt.Errorf("either --box-url or --box-file must be set")
	} else if boxUrlArg != "" && boxFileArg != "" {
		return fmt.Errorf("only one of --box-url or --box-file must be set")
	}

	url := boxUrlArg
	if url == "" {
		url = "file://" + boxFileArg
	}

	_, vethCidr, err := net.ParseCIDR(vethCidrArgNode)
	if err != nil {
		return err
	}

	runNode := run_box.RunBox{
		BoxUrl:          url,
		BoxName:         boxNameArg,
		WorkDir:         workDirArg,
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
