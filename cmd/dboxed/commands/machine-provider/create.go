package machine_provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the machine provider name. Must be unique." required:""`
	Type string `help:"Specify the provider type." required:"" enum:"aws,hetzner"`

	SshKeyPublic *string `help:"SSH public key for machines created by this provider"`

	// AWS-specific flags
	AwsRegion          *string `help:"AWS region (e.g., us-east-1)" group:"aws"`
	AwsVpcId           *string `help:"AWS VPC ID" group:"aws"`
	AwsAccessKeyId     *string `help:"AWS access key ID" group:"aws"`
	AwsSecretAccessKey *string `help:"AWS secret access key" group:"aws"`

	// Hetzner-specific flags
	HetznerCloudToken    *string `help:"Hetzner Cloud API token" group:"hetzner"`
	HetznerNetworkName   *string `help:"Hetzner network name" group:"hetzner"`
	HetznerRobotUsername *string `help:"Hetzner Robot username (optional, for dedicated servers)" group:"hetzner"`
	HetznerRobotPassword *string `help:"Hetzner Robot password (optional, for dedicated servers)" group:"hetzner"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.MachineProviderClient{Client: c}

	req := models.CreateMachineProvider{
		Type:         dmodel.MachineProviderType(cmd.Type),
		Name:         cmd.Name,
		SshKeyPublic: cmd.SshKeyPublic,
	}

	switch dmodel.MachineProviderType(cmd.Type) {
	case dmodel.MachineProviderTypeAws:
		if cmd.AwsRegion == nil {
			return fmt.Errorf("--aws-region is required for AWS provider")
		}
		if cmd.AwsVpcId == nil {
			return fmt.Errorf("--aws-vpc-id is required for AWS provider")
		}
		if cmd.AwsAccessKeyId == nil {
			return fmt.Errorf("--aws-access-key-id is required for AWS provider")
		}
		if cmd.AwsSecretAccessKey == nil {
			return fmt.Errorf("--aws-secret-access-key is required for AWS provider")
		}
		req.Aws = &models.CreateMachineProviderAws{
			Region:             *cmd.AwsRegion,
			VpcId:              *cmd.AwsVpcId,
			AwsAccessKeyId:     *cmd.AwsAccessKeyId,
			AwsSecretAccessKey: *cmd.AwsSecretAccessKey,
		}
	case dmodel.MachineProviderTypeHetzner:
		if cmd.HetznerCloudToken == nil {
			return fmt.Errorf("--hetzner-cloud-token is required for Hetzner provider")
		}
		if cmd.HetznerNetworkName == nil {
			return fmt.Errorf("--hetzner-network-name is required for Hetzner provider")
		}
		req.Hetzner = &models.CreateMachineProviderHetzner{
			CloudToken:         *cmd.HetznerCloudToken,
			HetznerNetworkName: *cmd.HetznerNetworkName,
			RobotUsername:      cmd.HetznerRobotUsername,
			RobotPassword:      cmd.HetznerRobotPassword,
		}
	default:
		return fmt.Errorf("unsupported provider type: %s", cmd.Type)
	}

	provider, err := c2.CreateMachineProvider(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("machine provider created", slog.Any("id", provider.ID), slog.Any("name", provider.Name))

	return nil
}
