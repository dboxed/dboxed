package baseclient

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/oauth2"
)

func (c *Client) GetClientAuth(withOverrides bool) *ClientAuth {
	ret := util.MustCopyViaJson(c.clientAuth)
	if withOverrides {
		ret.ApiUrl = c.getApiUrl()
		ret.StaticToken = c.GetApiToken()
		ret.WorkspaceId = c.getWorkspaceId()
	}
	return ret
}

func (c *Client) WriteClientAuth() error {
	return WriteClientAuth(c.clientAuthFile, c.clientAuth)
}

func (c *Client) LoginOAuth2(ctx context.Context) error {
	authInfo, err := requestApi2[models.AuthInfo](ctx, c, "GET", "v1/auth/info", struct{}{}, false)
	if err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.clientAuth.AuthInfo = authInfo
	c.clientAuth.Oauth2Token = nil
	c.clientAuth.StaticToken = nil

	ocfg, err := c.buildOAuth2Config(ctx)
	if err != nil {
		return err
	}

	deviceAuth, err := ocfg.DeviceAuth(ctx, oauth2.AccessTypeOffline)
	if err != nil {
		return err
	}

	fmt.Printf("Please visit %s to login\nYour user code is %s\n", deviceAuth.VerificationURIComplete, deviceAuth.UserCode)

	token, err := ocfg.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		return err
	}

	c.clientAuth.AuthInfo = authInfo
	c.clientAuth.Oauth2Token = token

	if c.writeClientAuth {
		err = c.WriteClientAuth()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) LoginStaticToken(ctx context.Context, staticToken string) error {
	authInfo, err := requestApi2[models.AuthInfo](ctx, c, "GET", "v1/auth/info", struct{}{}, false)
	if err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.clientAuth.AuthInfo = authInfo
	c.clientAuth.Oauth2Token = nil
	c.clientAuth.StaticToken = &staticToken

	if c.writeClientAuth {
		err = c.WriteClientAuth()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) buildOAuth2Config(ctx context.Context) (*oauth2.Config, error) {
	provider := c.provider
	if provider == nil {
		var err error
		provider, err = oidc.NewProvider(ctx, c.clientAuth.AuthInfo.OidcIssuerUrl)
		if err != nil {
			return nil, err
		}
		c.provider = provider
	}

	cfg := &oauth2.Config{
		ClientID: c.clientAuth.AuthInfo.OidcClientId,
		Endpoint: provider.Endpoint(),
	}

	return cfg, nil
}

func (c *Client) RefreshToken(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.clientAuth == nil || c.clientAuth.Oauth2Token == nil {
		return fmt.Errorf("client has no token, please login first")
	}

	if c.clientAuth.Oauth2Token.Valid() {
		return nil
	}

	slog.InfoContext(ctx, "refreshing token")

	ocfg, err := c.buildOAuth2Config(ctx)
	if err != nil {
		return err
	}
	ts := ocfg.TokenSource(ctx, c.clientAuth.Oauth2Token)
	newToken, err := ts.Token()
	if err != nil {
		return err
	}

	c.clientAuth.Oauth2Token = newToken
	if c.writeClientAuth {
		err = c.WriteClientAuth()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CheckAuth(ctx context.Context) error {
	if c.GetApiToken() == nil {
		_, err := c.CurrentUser(ctx)
		if err != nil {
			return err
		}
	} else {
		t, err := c.CurrentToken(ctx)
		if err != nil {
			return err
		}
		if t.ForWorkspace {
			c.clientAuth.WorkspaceId = &t.Workspace
		}
	}
	workspaceId := c.getWorkspaceId()
	if workspaceId != nil {
		_, err := c.GetWorkspaceById(ctx, *workspaceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) CurrentUser(ctx context.Context) (*models.User, error) {
	return RequestApi[models.User](ctx, c, "GET", "v1/auth/current-user", struct{}{})
}

func (c *Client) CurrentToken(ctx context.Context) (*models.Token, error) {
	return RequestApi[models.Token](ctx, c, "GET", "v1/auth/current-token", struct{}{})
}

func (c *Client) GetWorkspaceById(ctx context.Context, workspaceId int64) (*models.Workspace, error) {
	return RequestApi[models.Workspace](ctx, c, "GET", fmt.Sprintf("v1/workspaces/%d", workspaceId), struct{}{})
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]models.Workspace, error) {
	l, err := RequestApi[[]models.Workspace](ctx, c, "GET", "v1/workspaces", struct{}{})
	if err != nil {
		return nil, err
	}
	return *l, nil
}

func (c *Client) SwitchWorkspaceById(ctx context.Context, workspaceId int64) (*models.Workspace, error) {
	w, err := c.GetWorkspaceById(ctx, workspaceId)
	if err != nil {
		return nil, err
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.clientAuth.WorkspaceId = &workspaceId
	if c.writeClientAuth {
		err = c.WriteClientAuth()
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (c *Client) WorkspaceApiPath() (string, error) {
	workspaceId := c.getWorkspaceId()
	if workspaceId == nil {
		return "", fmt.Errorf("no workspace selected")
	}
	return fmt.Sprintf("v1/workspaces/%d", *workspaceId), nil
}

func (c *Client) BuildApiPath(workspace bool, pathElems ...any) (string, error) {
	p := "v1"
	if workspace {
		var err error
		p, err = c.WorkspaceApiPath()
		if err != nil {
			return "", err
		}
	}
	pathElems2 := make([]string, 0, len(pathElems)+1)
	pathElems2 = append(pathElems2, p)
	for _, e := range pathElems {
		pathElems2 = append(pathElems2, fmt.Sprint(e))
	}
	return path.Join(pathElems2...), nil
}
