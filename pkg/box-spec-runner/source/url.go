package source

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func NewUrlSource(ctx context.Context, url url.URL) (*BoxSpecSource, error) {
	switch url.Scheme {
	case "file":
		path := strings.TrimPrefix(url.String(), "file://")
		return NewFileSource(ctx, path, 5*time.Second)
	case "http", "https":
		return NewHttpSource(ctx, url, 5*time.Second)
	default:
		return nil, fmt.Errorf("unsupported url scheme %s", url.Scheme)
	}
}
