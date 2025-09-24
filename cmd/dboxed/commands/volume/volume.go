package volume

import (
	"context"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type VolumeCommands struct {
	Create CreateCmd `cmd:"" help:"Create a volume"`
	Delete DeleteCmd `cmd:"" help:"Delete a volume"`
	List   ListCmd   `cmd:"" help:"List volumes"`

	Debug DebugCmd `cmd:"" help:"Debug commands"`

	OnlyLinuxCmds
}

func getVolume(ctx context.Context, c *baseclient.Client, volume string) (*models.Volume, error) {
	c2 := clients.VolumesClient{Client: c}
	id, err := strconv.ParseInt(volume, 10, 64)
	if err == nil {
		v, err := c2.GetVolumeById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetVolumeByName(ctx, volume)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
