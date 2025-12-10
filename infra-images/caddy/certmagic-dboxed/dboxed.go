package certmagic_dboxed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
	"github.com/danielgtaylor/huma/v2"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

var (
	// implementing these interfaces
	_ caddy.Module          = (*CertMagicDboxed)(nil)
	_ certmagic.Storage     = (*CertMagicDboxed)(nil)
	_ certmagic.Locker      = (*CertMagicDboxed)(nil)
	_ caddy.Provisioner     = (*CertMagicDboxed)(nil)
	_ caddyfile.Unmarshaler = (*CertMagicDboxed)(nil)
)

func init() {
	caddy.RegisterModule(&CertMagicDboxed{})
}

type CertMagicDboxed struct {
	logger *zap.Logger
	client *minio.Client

	// CertMagicDboxed configuration
	ApiUrl   string `json:"apiUrl"`
	ApiToken string `json:"apiToken"`

	WorkspaceId    string `json:"workspaceId"`
	LoadBalancerId string `json:"loadBalancerId"`
}

func (c *CertMagicDboxed) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		key := d.Val()

		var value string
		if !d.Args(&value) {
			continue
		}

		switch key {
		case "apiUrl":
			c.ApiUrl = value
		case "apiToken":
			c.ApiToken = value
		case "workspaceId":
			c.WorkspaceId = value
		case "loadBalancerId":
			c.LoadBalancerId = value
		}
	}

	return nil
}

func (c *CertMagicDboxed) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()

	c.ApiUrl = repl.ReplaceKnown(c.ApiUrl, "")
	c.ApiToken = repl.ReplaceKnown(c.ApiToken, "")

	c.logger = ctx.Logger(c)

	c.logger.Info("hello world")

	return nil
}

func (*CertMagicDboxed) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.storage.dboxed",
		New: func() caddy.Module {
			return &CertMagicDboxed{}
		},
	}
}

func (c *CertMagicDboxed) CertMagicStorage() (certmagic.Storage, error) {
	return c, nil
}

var (
	LockPollInterval = 1 * time.Second
	LockTimeout      = 15 * time.Second
)

func (c *CertMagicDboxed) Lock(ctx context.Context, key string) error {
	startedAt := time.Now()
	for {
		err := c.tryLock(ctx, key)
		if err == nil {
			return nil
		}
		c.logger.Info(fmt.Sprintf("tryLock failed: %s", err.Error()))

		if startedAt.Add(LockTimeout).Before(time.Now()) {
			return errors.New("acquiring lock failed")
		}
		time.Sleep(LockPollInterval)
	}
}

func (c *CertMagicDboxed) tryLock(ctx context.Context, key string) error {
	c.logger.Info(fmt.Sprintf("tryLock: %v", key))

	resp, err := c.requestApi(ctx, "PUT", fmt.Sprintf("certmagic/locks/%s", key), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CertMagicDboxed) Unlock(ctx context.Context, key string) error {
	c.logger.Info(fmt.Sprintf("Release lock: %v", key))
	resp, err := c.requestApi(ctx, "DELETE", fmt.Sprintf("certmagic/locks/%s", key), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CertMagicDboxed) Store(ctx context.Context, key string, value []byte) error {
	length := int64(len(value))

	c.logger.Info(fmt.Sprintf("Store: %s, %d bytes", key, length))

	resp, err := c.requestApi(ctx, "PUT", fmt.Sprintf("certmagic/objects/%s", key), value)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusConflict {
			return fs.ErrExist
		}
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CertMagicDboxed) Load(ctx context.Context, key string) ([]byte, error) {
	c.logger.Info(fmt.Sprintf("Load key: %s", key))

	resp, err := c.requestApi(ctx, "GET", fmt.Sprintf("certmagic/objects/%s", key), nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *CertMagicDboxed) Delete(ctx context.Context, key string) error {
	c.logger.Info(fmt.Sprintf("Delete key: %s", key))

	resp, err := c.requestApi(ctx, "DELETE", fmt.Sprintf("certmagic/objects/%s", key), nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fs.ErrNotExist
		}
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CertMagicDboxed) Exists(ctx context.Context, key string) bool {
	_, err := c.Stat(ctx, key)
	exists := err == nil
	c.logger.Info(fmt.Sprintf("Check exists: %s, %t", key, exists))
	return exists
}

func (c *CertMagicDboxed) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
	prefix = strings.TrimSuffix(prefix, "/")
	resp, err := c.requestApi(ctx, "GET", fmt.Sprintf("certmagic/objects/%s?list=true&recursive=%v", prefix, recursive), nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keys []string
	err = json.Unmarshal(b, &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (c *CertMagicDboxed) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	resp, err := c.requestApi(ctx, "HEAD", fmt.Sprintf("certmagic/objects/%s", key), nil)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Stat key: %s, error: %v", key, err))
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return certmagic.KeyInfo{}, fs.ErrNotExist
		}
		return certmagic.KeyInfo{}, err
	}
	defer resp.Body.Close()

	sizeStr := resp.Header.Get("X-Size")
	size, err := strconv.ParseInt(sizeStr, 10, 32)
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	key = resp.Header.Get("X-Key")
	lastModifiedStr := resp.Header.Get("X-Last-Modified")
	lastModified, err := time.Parse(time.RFC3339Nano, lastModifiedStr)
	if err != nil {
		return certmagic.KeyInfo{}, err
	}

	c.logger.Info(fmt.Sprintf("Stat key: %s, size: %d bytes, lastModified: %s", key, size, lastModified))

	return certmagic.KeyInfo{
		Key:        key,
		Modified:   lastModified,
		Size:       size,
		IsTerminal: strings.HasSuffix(key, "/"),
	}, err
}

func (c *CertMagicDboxed) requestApi(ctx context.Context, method string, relPath string, body []byte) (*http.Response, error) {
	u := c.ApiUrl
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += fmt.Sprintf("v1/workspaces/%s/load-balancers/%s/%s", c.WorkspaceId, c.LoadBalancerId, relPath)
	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiToken))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	var em huma.ErrorModel
	err = json.Unmarshal(b, &em)
	if err != nil {
		return resp, fmt.Errorf("%s request returned http status %s", u, resp.Status)
	}
	return resp, &em
}

func (c *CertMagicDboxed) String() string {
	return fmt.Sprintf("CertMagicDboxed Storage ApiUrl: %s, WorkspaceId: %s, LoadBalancerId: %s", c.ApiUrl, c.WorkspaceId, c.LoadBalancerId)
}
