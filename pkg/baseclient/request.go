package baseclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

func RequestApi[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, body RequestBody) (*ReplyBody, error) {
	return requestApi2[ReplyBody, RequestBody](ctx, c, method, p, body, true)
}

func requestApi2[ReplyBody any, RequestBody any](ctx context.Context, c *Client, method string, p string, body RequestBody, withToken bool) (*ReplyBody, error) {
	if withToken && c.clientAuth.StaticToken == nil {
		err := c.RefreshToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(c.clientAuth.ApiUrl)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, p)

	//fmt.Fprintf(os.Stderr, "request: %s\n", u.String())

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
		if c.clientAuth.StaticToken != nil {
			req.Header.Set("Authorization", "Bearer "+*c.clientAuth.StaticToken)
		} else if c.clientAuth.Oauth2Token != nil {
			req.Header.Set("Authorization", "Bearer "+c.clientAuth.Oauth2Token.AccessToken)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s request returned http status %s", p, resp.Status)
	}

	b, err = io.ReadAll(resp.Body)
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
