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

type LogsClient struct {
	Client *baseclient.Client
}

func (c *LogsClient) PostLogLines(ctx context.Context, req models.PostLogs) error {
	p, err := c.Client.BuildApiPath(true, "logs")
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "POST", p, req)
	return err
}

func (c *LogsClient) ListLogs(ctx context.Context, ownerType string, ownerId string) ([]models.LogMetadataModel, error) {
	p, err := c.Client.BuildApiPath(true, "logs")
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("owner_type", ownerType)
	q.Set("owner_id", ownerId)
	p += "?" + q.Encode()

	l, err := baseclient.RequestApi[huma_utils.ListBody[models.LogMetadataModel]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

// StreamLogs streams logs from the specified log ID and calls the callback for each event
func (c *LogsClient) StreamLogs(ctx context.Context, logId string, since string, callback func(interface{}) error) error {
	p, err := c.Client.BuildApiPath(true, "logs", logId, "stream")
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
