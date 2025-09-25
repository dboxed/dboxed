package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct{}

func (cmd *ListCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	boxes, err := c2.ListBoxes(ctx)
	if err != nil {
		return err
	}

	for _, b := range boxes {
		slog.Info("box", slog.Any("id", b.ID), slog.Any("name", b.Name), slog.Any("created_at", b.CreatedAt))
	}

	return nil
}
