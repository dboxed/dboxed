package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

func (cmd *DeleteCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	b, err := GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	err = c2.DeleteBox(ctx, b.ID)
	if err != nil {
		return err
	}

	slog.Info("box deleted", slog.Any("id", b.ID))

	return nil
}
