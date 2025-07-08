package run

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/utils"
	"github.com/koobox/unboxed/pkg/run-box"
	"github.com/spf13/cobra"
	"net"
	"time"
)

var workDirArg string

var boxUrlArg string
var boxFileArg string
var boxNameArg string
var vethCidrArg string

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Download, unpack and run the a box",
	Long:  ``,
	RunE:  doRun,
}

func init() {
	RunCmd.Flags().StringVar(&workDirArg, "work-dir", utils.DefaultWorkdir, "unboxed work dir")

	RunCmd.Flags().StringVar(&boxUrlArg, "box-url", "", "box url")
	RunCmd.Flags().StringVar(&boxFileArg, "box-file", "", "box url")
	RunCmd.Flags().StringVar(&boxNameArg, "box-name", "", "box name")
	_ = RunCmd.MarkFlagRequired("box-name")

	RunCmd.Flags().StringVar(&vethCidrArg, "veth-cidr", "1.2.3.0/30", "veth cidr")
}

func doRun(cmd *cobra.Command, args []string) error {
	url, err := utils.GetBoxUrl(boxUrlArg, boxFileArg)
	if err != nil {
		return err
	}

	_, vethCidr, err := net.ParseCIDR(vethCidrArg)
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
