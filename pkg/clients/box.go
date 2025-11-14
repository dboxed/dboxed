package clients

import (
	"context"
	"encoding/json"
	"net/url"

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

func (c *BoxClient) GetBoxById(ctx context.Context, id string) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
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

func (c *BoxClient) GetBoxSpecById(ctx context.Context, id string) (*boxspec.BoxSpec, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id, "box-spec")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[boxspec.BoxSpec](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) DeleteBox(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) UpdateBox(ctx context.Context, id string, req models.UpdateBox) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) StartBox(ctx context.Context, id string) (*models.Box, error) {
	desiredState := "up"
	return c.UpdateBox(ctx, id, models.UpdateBox{
		DesiredState: &desiredState,
	})
}

func (c *BoxClient) StopBox(ctx context.Context, id string) (*models.Box, error) {
	desiredState := "down"
	return c.UpdateBox(ctx, id, models.UpdateBox{
		DesiredState: &desiredState,
	})
}

func (c *BoxClient) GetSandboxStatus(ctx context.Context, boxId string) (*models.BoxSandboxStatus, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "sandbox-status")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.BoxSandboxStatus](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) UpdateSandboxStatus(ctx context.Context, boxId string, req models.UpdateBoxSandboxStatus) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "sandbox-status")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "PATCH", p, req)
	return err
}

func (c *BoxClient) PostLogLines(ctx context.Context, boxId string, req models.PostLogs) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "logs")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *BoxClient) ListLogs(ctx context.Context, boxId string) ([]models.LogMetadataModel, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "logs")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.LogMetadataModel]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) ListComposeProjects(ctx context.Context, boxId string) ([]models.BoxComposeProject, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.BoxComposeProject]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) CreateComposeProject(ctx context.Context, boxId string, req models.CreateBoxComposeProject) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *BoxClient) UpdateComposeProject(ctx context.Context, boxId string, composeName string, req models.UpdateBoxComposeProject) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects", composeName)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "PATCH", p, req)
	return err
}

func (c *BoxClient) DeleteComposeProject(ctx context.Context, boxId string, composeName string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects", composeName)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) ListAttachedVolumes(ctx context.Context, boxId string) ([]models.VolumeAttachment, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.VolumeAttachment]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) AttachVolume(ctx context.Context, boxId string, req models.AttachVolumeRequest) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *BoxClient) UpdateAttachedVolume(ctx context.Context, boxId string, volumeId string, req models.UpdateVolumeAttachmentRequest) (*models.VolumeAttachment, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes", volumeId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeAttachment](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) DetachVolume(ctx context.Context, boxId string, volumeId string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes", volumeId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

// StreamLogs streams logs from the specified log ID and calls the callback for each event
func (c *BoxClient) StreamLogs(ctx context.Context, boxId string, logId string, since string, callback func(interface{}) error) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "logs", logId, "stream")
	if err != nil {
		return err
	}

	q := url.Values{}
	if since != "" {
		q.Set("since", since)
	}

	return baseclient.RequestApiSSE(ctx, c.Client, p, q, func(m baseclient.SSEMessage) error {
		switch m.Event {
		case "metadata":
			var x models.LogMetadataModel
			err := json.Unmarshal([]byte(m.Data), &x)
			if err != nil {
				return err
			}
			return callback(x)
		case "logs-batch":
			var x boxspec.LogsBatch
			err := json.Unmarshal([]byte(m.Data), &x)
			if err != nil {
				return err
			}
			return callback(x)
		default:
			return nil
		}
	})
}

func (c *BoxClient) ListPortForwards(ctx context.Context, boxId string) ([]models.BoxPortForward, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "port-forwards")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.BoxPortForward]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) CreatePortForward(ctx context.Context, boxId string, req models.CreateBoxPortForward) (*models.BoxPortForward, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "port-forwards")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.BoxPortForward](ctx, c.Client, "POST", p, req)
}

func (c *BoxClient) UpdatePortForward(ctx context.Context, boxId string, portForwardId string, req models.UpdateBoxPortForward) (*models.BoxPortForward, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "port-forwards", portForwardId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.BoxPortForward](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) DeletePortForward(ctx context.Context, boxId string, portForwardId string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "port-forwards", portForwardId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) ListLoadBalancerServices(ctx context.Context, boxId string) ([]models.LoadBalancerService, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "load-balancer-services")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.LoadBalancerService]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *BoxClient) CreateLoadBalancerService(ctx context.Context, boxId string, req models.CreateLoadBalancerService) (*models.LoadBalancerService, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "load-balancer-services")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.LoadBalancerService](ctx, c.Client, "POST", p, req)
}

func (c *BoxClient) UpdateLoadBalancerService(ctx context.Context, boxId string, lbServiceId string, req models.UpdateLoadBalancerService) (*models.LoadBalancerService, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "load-balancer-services", lbServiceId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.LoadBalancerService](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) DeleteLoadBalancerService(ctx context.Context, boxId string, lbServiceId string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "load-balancer-services", lbServiceId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
