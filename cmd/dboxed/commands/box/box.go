package box

import (
	"context"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type BoxCommands struct {
	Create CreateCmd `cmd:"" help:"Create a box"`
	Get    GetCmd    `cmd:"" help:"Get a box"`
	List   ListCmd   `cmd:"" help:"List boxes"`
	Delete DeleteCmd `cmd:"" help:"Delete a box"`

	onlyLinux
}

func GetBox(ctx context.Context, c *baseclient.Client, box string) (*models.Box, error) {
	c2 := clients.BoxClient{Client: c}
	id, err := strconv.ParseInt(box, 10, 64)
	if err == nil {
		v, err := c2.GetBoxById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetBoxByName(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
