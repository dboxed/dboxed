package start

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/utils"
	"github.com/koobox/unboxed/pkg/start-box"
	"github.com/spf13/cobra"
	"net"
)

var workDirArg string

var boxUrlArg string
var boxFileArg string
var boxNameArg string
var vethCidrArg string

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Download, unpack and start the a box",
	Long:  ``,
	RunE:  doStart,
}

func init() {
	StartCmd.Flags().StringVar(&workDirArg, "work-dir", utils.DefaultWorkdir, "unboxed work dir")

	StartCmd.Flags().StringVar(&boxUrlArg, "box-url", "", "box url")
	StartCmd.Flags().StringVar(&boxFileArg, "box-file", "", "box url")
	StartCmd.Flags().StringVar(&boxNameArg, "box-name", "", "box name")
	_ = StartCmd.MarkFlagRequired("box-name")

	StartCmd.Flags().StringVar(&vethCidrArg, "veth-cidr", "1.2.3.0/30", "veth cidr")
}

func doStart(cmd *cobra.Command, args []string) error {
	url, err := utils.GetBoxUrl(boxUrlArg, boxFileArg)
	if err != nil {
		return err
	}

	_, vethCidr, err := net.ParseCIDR(vethCidrArg)
	if err != nil {
		return err
	}

	startBox := start_box.StartBox{
		BoxUrl:          url,
		BoxName:         boxNameArg,
		WorkDir:         workDirArg,
		VethNetworkCidr: vethCidr,
	}

	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = startBox.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
