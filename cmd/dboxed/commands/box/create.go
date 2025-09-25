package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the box name. Must be unique." required:"" arg:""`
}

func (cmd *CreateCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	req := models.CreateBox{
		Name: cmd.Name,
	}

	b, err := c2.CreateBox(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("box created", slog.Any("id", b.ID), slog.Any("name", b.Name))

	return nil
}
