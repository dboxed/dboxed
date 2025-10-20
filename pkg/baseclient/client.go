package baseclient

import (
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dboxed/dboxed/pkg/util"
)

const defaultApiUrl = "https://api.dboxed.io"

type Client struct {
	clientAuthFile  *string
	clientAuth      *ClientAuth
	writeClientAuth bool

	overrideApiUrl      *string
	overrideApiToken    *string
	overrideWorkspaceId *int64

	debug bool

	m        sync.Mutex
	provider *oidc.Provider
}

func New(clientAuthFile *string, clientAuth *ClientAuth, writeClientAuth bool) (*Client, error) {
	clientAuth, err := util.CopyViaJson(clientAuth)
	if err != nil {
		return nil, err
	}

	if clientAuth.ApiUrl == "" {
		clientAuth.ApiUrl = defaultApiUrl
	}

	c := &Client{
		clientAuthFile:  clientAuthFile,
		clientAuth:      clientAuth,
		writeClientAuth: writeClientAuth,
	}

	return c, nil
}

func (c *Client) SetOverrideApiUrl(url string) {
	c.overrideApiUrl = &url
}

func (c *Client) SetOverrideApiToken(token string) {
	c.overrideApiToken = &token
}

func (c *Client) SetOverrideWorkspaceId(id int64) {
	c.overrideWorkspaceId = &id
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) getApiUrl() string {
	if c.overrideApiUrl != nil {
		return *c.overrideApiUrl
	}
	return c.clientAuth.ApiUrl
}

func (c *Client) GetApiToken() *string {
	if c.overrideApiToken != nil {
		return c.overrideApiToken
	}
	return c.clientAuth.StaticToken
}

func (c *Client) getWorkspaceId() *int64 {
	if c.overrideWorkspaceId != nil {
		return c.overrideWorkspaceId
	}
	return c.clientAuth.WorkspaceId
}
