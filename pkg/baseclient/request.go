package baseclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/danielgtaylor/huma/v2"
)

func RequestApi[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, body RequestBody) (*ReplyBody, error) {
	return requestApi2[ReplyBody, RequestBody](ctx, c, method, p, body, true)
}

func requestApi2[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, body RequestBody, withToken bool) (*ReplyBody, error) {
	apiToken := c.GetApiToken()
	if withToken && apiToken == nil {
		err := c.RefreshToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(c.getApiUrl())
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, p)

	if c.debug {
		slog.Debug("API request", slog.String("method", method), slog.String("url", u.String()))
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))

	if withToken {
		if apiToken != nil {
			req.Header.Set("Authorization", "Bearer "+*apiToken)
		} else if c.clientAuth.Oauth2Token != nil {
			req.Header.Set("Authorization", "Bearer "+c.clientAuth.Oauth2Token.AccessToken)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if c.debug {
		slog.Debug("API response", slog.String("method", method), slog.String("url", u.String()), slog.Int("status", resp.StatusCode))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var em huma.ErrorModel
		err = json.Unmarshal(b, &em)
		if err != nil {
			return nil, fmt.Errorf("%s request returned http status %s", p, resp.Status)
		}
		return nil, &em
	}

	var reply ReplyBody
	err = json.Unmarshal(b, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}
