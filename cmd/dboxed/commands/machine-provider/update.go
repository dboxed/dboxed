package machine_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	MachineProvider string `help:"Specify the machine provider ID or name" required:"" arg:""`

	SshKeyPublic *string `help:"SSH public key for machines created by this provider"`

	// AWS-specific flags
	AwsAccessKeyId     *string `help:"AWS access key ID" group:"aws"`
	AwsSecretAccessKey *string `help:"AWS secret access key" group:"aws"`

	// Hetzner-specific flags
	HetznerCloudToken    *string `help:"Hetzner Cloud API token" group:"hetzner"`
	HetznerRobotUsername *string `help:"Hetzner Robot username (for dedicated servers)" group:"hetzner"`
	HetznerRobotPassword *string `help:"Hetzner Robot password (for dedicated servers)" group:"hetzner"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	mp, err := commandutils.GetMachineProvider(ctx, c, cmd.MachineProvider)
	if err != nil {
		return err
	}

	c2 := &clients.MachineProviderClient{Client: c}

	req := models.UpdateMachineProvider{
		SshKeyPublic: cmd.SshKeyPublic,
	}

	// Set AWS-specific updates if any AWS flags are provided
	if mp.Type == dmodel.MachineProviderTypeAws {
		req.Aws = &models.UpdateMachineProviderAws{
			AwsAccessKeyId:     cmd.AwsAccessKeyId,
			AwsSecretAccessKey: cmd.AwsSecretAccessKey,
		}
	}

	// Set Hetzner-specific updates if any Hetzner flags are provided
	if mp.Type == dmodel.MachineProviderTypeHetzner {
		req.Hetzner = &models.UpdateMachineProviderHetzner{
			CloudToken:    cmd.HetznerCloudToken,
			RobotUsername: cmd.HetznerRobotUsername,
			RobotPassword: cmd.HetznerRobotPassword,
		}
	}

	updated, err := c2.UpdateMachineProvider(ctx, mp.ID, req)
	if err != nil {
		return err
	}

	slog.Info("machine provider updated", slog.Any("id", updated.ID), slog.Any("name", updated.Name))

	return nil
}
