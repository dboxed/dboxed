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

func (c *BoxClient) GetBoxSpecById(ctx context.Context, id int64) (*boxspec.BoxSpec, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id, "box-spec")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[boxspec.BoxSpec](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) DeleteBox(ctx context.Context, id int64) error {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) UpdateBox(ctx context.Context, id int64, req models.UpdateBox) (*models.Box, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Box](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) StartBox(ctx context.Context, id int64) (*models.Box, error) {
	desiredState := "up"
	return c.UpdateBox(ctx, id, models.UpdateBox{
		DesiredState: &desiredState,
	})
}

func (c *BoxClient) StopBox(ctx context.Context, id int64) (*models.Box, error) {
	desiredState := "down"
	return c.UpdateBox(ctx, id, models.UpdateBox{
		DesiredState: &desiredState,
	})
}

func (c *BoxClient) GetBoxRunStatus(ctx context.Context, boxId int64) (*models.BoxRunStatus, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "run-status")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.BoxRunStatus](ctx, c.Client, "GET", p, struct{}{})
}

func (c *BoxClient) UpdateBoxRunStatus(ctx context.Context, boxId int64, req models.UpdateBoxRunStatus) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "run-status")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "PATCH", p, req)
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

func (c *BoxClient) ListLogs(ctx context.Context, boxId int64) ([]models.LogMetadataModel, error) {
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

func (c *BoxClient) ListComposeProjects(ctx context.Context, boxId int64) ([]models.BoxComposeProject, error) {
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

func (c *BoxClient) CreateComposeProject(ctx context.Context, boxId int64, req models.CreateBoxComposeProject) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *BoxClient) UpdateComposeProject(ctx context.Context, boxId int64, composeName string, req models.UpdateBoxComposeProject) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects", composeName)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "PATCH", p, req)
	return err
}

func (c *BoxClient) DeleteComposeProject(ctx context.Context, boxId int64, composeName string) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "compose-projects", composeName)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *BoxClient) ListAttachedVolumes(ctx context.Context, boxId int64) ([]models.VolumeAttachment, error) {
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

func (c *BoxClient) AttachVolume(ctx context.Context, boxId int64, req models.AttachVolumeRequest) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *BoxClient) UpdateAttachedVolume(ctx context.Context, boxId int64, volumeId int64, req models.UpdateVolumeAttachmentRequest) (*models.VolumeAttachment, error) {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes", volumeId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeAttachment](ctx, c.Client, "PATCH", p, req)
}

func (c *BoxClient) DetachVolume(ctx context.Context, boxId int64, volumeId int64) error {
	p, err := c.Client.BuildApiPath(true, "boxes", boxId, "volumes", volumeId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

// StreamLogs streams logs from the specified log ID and calls the callback for each event
func (c *BoxClient) StreamLogs(ctx context.Context, boxId int64, logId int64, since string, callback func(interface{}) error) error {
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
