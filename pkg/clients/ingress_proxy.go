package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type IngressProxyClient struct {
	Client *baseclient.Client
}

func (c *IngressProxyClient) CreateIngressProxy(ctx context.Context, req models.CreateIngressProxy) (*models.IngressProxy, error) {
	p, err := c.Client.BuildApiPath(true, "ingress-proxies")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.IngressProxy](ctx, c.Client, "POST", p, req)
}

func (c *IngressProxyClient) ListIngressProxies(ctx context.Context) ([]models.IngressProxy, error) {
	p, err := c.Client.BuildApiPath(true, "ingress-proxies")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.IngressProxy]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *IngressProxyClient) GetIngressProxyById(ctx context.Context, id string) (*models.IngressProxy, error) {
	p, err := c.Client.BuildApiPath(true, "ingress-proxies", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.IngressProxy](ctx, c.Client, "GET", p, struct{}{})
}

func (c *IngressProxyClient) UpdateIngressProxy(ctx context.Context, id string, req models.UpdateIngressProxy) (*models.IngressProxy, error) {
	p, err := c.Client.BuildApiPath(true, "ingress-proxies", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.IngressProxy](ctx, c.Client, "PATCH", p, req)
}

func (c *IngressProxyClient) DeleteIngressProxy(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "ingress-proxies", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
