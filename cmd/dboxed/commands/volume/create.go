package volume

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/services"
)

type CreateCmd struct {
	Name string `help:"Specify the volume name. Must be unique in the repository." required:"true"`

	VolumeProvider string `help:"Specify the volume provider" required:""`

	FsType string `help:"Specify the filesystem type" default:"ext4"`
	FsSize string `help:"Specify the maximum filesystem size." required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	s := &services.VolumesService{Client: c}
	return s.CreateVolume(&services.CreateVolumeCmdOpts{
		Name:           cmd.Name,
		VolumeProvider: cmd.VolumeProvider,
		FsType:         cmd.FsType,
		FsSize:         cmd.FsSize,
	})
}
