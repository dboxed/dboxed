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

	overrideApiUrl      string
	overrideApiToken    string
	overrideWorkspaceId int64

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
	c.overrideApiUrl = url
}

func (c *Client) SetOverrideApiToken(token string) {
	c.overrideApiToken = token
}

func (c *Client) SetOverrideWorkspaceId(id int64) {
	c.overrideWorkspaceId = id
}
