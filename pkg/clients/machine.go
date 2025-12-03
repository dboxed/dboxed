package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type MachineClient struct {
	Client *baseclient.Client
}

func (c *MachineClient) CreateMachine(ctx context.Context, req models.CreateMachine) (*models.Machine, error) {
	p, err := c.Client.BuildApiPath(true, "machines")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Machine](ctx, c.Client, "POST", p, req)
}

func (c *MachineClient) ListMachines(ctx context.Context) ([]models.Machine, error) {
	p, err := c.Client.BuildApiPath(true, "machines")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Machine]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *MachineClient) GetMachineById(ctx context.Context, id string) (*models.Machine, error) {
	p, err := c.Client.BuildApiPath(true, "machines", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Machine](ctx, c.Client, "GET", p, struct{}{})
}

func (c *MachineClient) UpdateMachine(ctx context.Context, id string, req models.UpdateMachine) (*models.Machine, error) {
	p, err := c.Client.BuildApiPath(true, "machines", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Machine](ctx, c.Client, "PATCH", p, req)
}

func (c *MachineClient) DeleteMachine(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "machines", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *MachineClient) ListBoxes(ctx context.Context, machineId string) ([]models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "machines", machineId, "boxes")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Box]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *MachineClient) AddBox(ctx context.Context, machineId string, req models.AddBoxToMachineRequest) error {
	p, err := c.Client.BuildApiPath(true, "machines", machineId, "boxes")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *MachineClient) RemoveBox(ctx context.Context, machineId string, boxId string) error {
	p, err := c.Client.BuildApiPath(true, "machines", machineId, "boxes", boxId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
