package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type BoxClient struct {
	Client *baseclient.Client
}

func (c *BoxClient) CreateBox(ctx context.Context, req models.CreateBox) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "POST", p, req)
}

func (c *BoxClient) ListBoxes(ctx context.Context) ([]models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Box]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) GetBoxById(ctx context.Context, id int64) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) GetBoxByUuid(ctx context.Context, uuid string) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", "by-uuid", uuid)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) GetBoxByName(ctx context.Context, name string) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", "by-name", name)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) GetBoxSpecById(ctx context.Context, id int64) (*boxspec.BoxFile, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id, "box-spec")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[boxspec.BoxFile](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) DeleteBox(ctx context.Context, id int64) error {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) PostLogLines(ctx context.Context, boxId int64, req models.PostLogs) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "logs")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}
