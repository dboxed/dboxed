package box_spec

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func NewUrlSource(ctx context.Context, url url.URL, natsNkeySeed []byte) (*BoxSpecSource, error) {
	switch url.Scheme {
	case "file":
		path := strings.TrimPrefix(url.String(), "file://")
		return NewFileSource(ctx, path, 5*time.Second)
	case "http", "https":
		return NewHttpSource(ctx, url, 5*time.Second)
	case "nats":
		return NewNatsSource(ctx, url, natsNkeySeed)
	default:
		return nil, fmt.Errorf("unsupported url scheme %s", url.Scheme)
	}
}
