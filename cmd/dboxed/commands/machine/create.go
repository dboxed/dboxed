package machine

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the machine name. Must be unique." required:""`

	MachineProvider *string `help:"Machine provider ID or name"`

	HetznerServerType     *string `help:"Hetzner server type (e.g., cx11, cpx11)" group:"hetzner"`
	HetznerServerLocation *string `help:"Hetzner server location (e.g., fsn1, nbg1)" group:"hetzner"`

	AwsInstanceType   *string `help:"AWS instance type (e.g., t3.micro)" group:"aws"`
	AwsSubnetId       *string `help:"AWS subnet ID" group:"aws"`
	AwsRootVolumeSize *int64  `help:"AWS root volume size in GB" group:"aws"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	req := models.CreateMachine{
		Name: cmd.Name,
	}

	var mp *models.MachineProvider
	if cmd.MachineProvider != nil {
		mp, err = commandutils.GetMachineProvider(ctx, c, *cmd.MachineProvider)
		if err != nil {
			return err
		}
		req.MachineProvider = &mp.ID
	}

	if mp != nil && mp.Type == dmodel.MachineProviderTypeHetzner {
		req.Hetzner = &models.CreateMachineHetzner{}
		if cmd.HetznerServerType != nil {
			req.Hetzner.ServerType = *cmd.HetznerServerType
		}
		if cmd.HetznerServerLocation != nil {
			req.Hetzner.ServerLocation = *cmd.HetznerServerLocation
		}
	}

	if mp != nil && mp.Type == dmodel.MachineProviderTypeAws {
		req.Aws = &models.CreateMachineAws{}
		if cmd.AwsInstanceType != nil {
			req.Aws.InstanceType = *cmd.AwsInstanceType
		}
		if cmd.AwsSubnetId != nil {
			req.Aws.SubnetId = *cmd.AwsSubnetId
		}
		if cmd.AwsRootVolumeSize != nil {
			req.Aws.RootVolumeSize = cmd.AwsRootVolumeSize
		}
	}

	c2 := &clients.MachineClient{Client: c}

	m, err := c2.CreateMachine(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("machine created", slog.Any("id", m.ID), slog.Any("name", m.Name))

	return nil
}
