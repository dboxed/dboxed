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
	return requestApiJson[ReplyBody, RequestBody](ctx, c, method, p, nil, body, true, nil)
}

func RequestApiQ[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, q url.Values, body RequestBody) (*ReplyBody, error) {
	return requestApiJson[ReplyBody, RequestBody](ctx, c, method, p, q, body, true, nil)
}

func RequestApiResponse[RequestBody any](ctx context.Context, c *Client, method string, p string, q url.Values, body RequestBody, header *http.Header) (*http.Response, error) {
	return requestApiResponse[RequestBody](ctx, c, method, p, q, body, true, header)
}

func requestApiJson[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, q url.Values, body RequestBody, withToken bool, header *http.Header) (*ReplyBody, error) {
	resp, err := requestApiResponse(ctx, c, method, p, q, body, withToken, header)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reply ReplyBody
	err = json.Unmarshal(b, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

func requestApiResponse[RequestBody any](ctx context.Context, c *Client, method string, p string, q url.Values, body RequestBody, withToken bool, header *http.Header) (*http.Response, error) {
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
	u.RawQuery = q.Encode()

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

	if header != nil {
		for k, vv := range *header {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	if withToken {
		if apiToken != nil {
			req.Header.Set("Authorization", "Bearer "+*apiToken)
		} else if c.clientAuth.Oauth2Token != nil {
			req.Header.Set("Authorization", "Bearer "+c.clientAuth.Oauth2Token.AccessToken)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.debug {
		slog.Debug("API response", slog.String("method", method), slog.String("url", u.String()), slog.Int("status", resp.StatusCode))
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var em huma.ErrorModel
	err = json.Unmarshal(b, &em)
	if err != nil {
		return nil, fmt.Errorf("%s request returned http status %s", p, resp.Status)
	}
	return nil, &em
}
