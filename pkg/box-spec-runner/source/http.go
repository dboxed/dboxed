package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dustin/go-humanize"
)

func NewHttpSource(ctx context.Context, url url.URL, interval time.Duration) (*BoxSpecSource, error) {
	return NewPollSource(ctx, func(ctx context.Context) ([]byte, error) {
		return downloadSpec(ctx, url)
	}, interval)
}

func downloadSpec(ctx context.Context, url url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.LimitReader(resp.Body, humanize.MByte))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download box spec: status=%d, body=%s", resp.StatusCode, string(b))
	}

	return b, nil
}
