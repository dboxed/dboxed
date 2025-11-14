package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type LoadBalancerClient struct {
	Client *baseclient.Client
}

func (c *LoadBalancerClient) CreateLoadBalancer(ctx context.Context, req models.CreateLoadBalancer) (*models.LoadBalancer, error) {
	p, err := c.Client.BuildApiPath(true, "load-balancers")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.LoadBalancer](ctx, c.Client, "POST", p, req)
}

func (c *LoadBalancerClient) ListLoadBalancers(ctx context.Context) ([]models.LoadBalancer, error) {
	p, err := c.Client.BuildApiPath(true, "load-balancers")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.LoadBalancer]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *LoadBalancerClient) GetLoadBalancerById(ctx context.Context, id string) (*models.LoadBalancer, error) {
	p, err := c.Client.BuildApiPath(true, "load-balancers", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.LoadBalancer](ctx, c.Client, "GET", p, struct{}{})
}

func (c *LoadBalancerClient) UpdateLoadBalancer(ctx context.Context, id string, req models.UpdateLoadBalancer) (*models.LoadBalancer, error) {
	p, err := c.Client.BuildApiPath(true, "load-balancers", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.LoadBalancer](ctx, c.Client, "PATCH", p, req)
}

func (c *LoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "load-balancers", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
