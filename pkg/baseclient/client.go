package baseclient

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/vishvananda/netns"
)

const defaultApiUrl = "https://api.dboxed.io"

type Client struct {
	clientAuthFile  *string
	clientAuth      *ClientAuth
	writeClientAuth bool

	overrideApiUrl      *string
	overrideApiToken    *string
	overrideWorkspaceId *string

	resolveNs  *netns.NsHandle
	connNs     *netns.NsHandle
	httpClient *http.Client

	debug bool

	m        sync.Mutex
	provider *oidc.Provider
}

type ClientOpt func(c *Client)

func WithNetworkNamespace(resolveNs *netns.NsHandle, connNs *netns.NsHandle) ClientOpt {
	return func(c *Client) {
		c.resolveNs = resolveNs
		c.connNs = connNs
	}
}

func New(clientAuthFile *string, clientAuth *ClientAuth, writeClientAuth bool, opts ...ClientOpt) (*Client, error) {
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

	for _, o := range opts {
		o(c)
	}

	err = c.setupHttpClient()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) SetOverrideApiUrl(url string) {
	c.overrideApiUrl = &url
}

func (c *Client) SetOverrideApiToken(token string) {
	c.overrideApiToken = &token
}

func (c *Client) SetOverrideWorkspaceId(id string) {
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

func (c *Client) GetWorkspaceId() *string {
	if c.overrideWorkspaceId != nil {
		return c.overrideWorkspaceId
	}
	return c.clientAuth.WorkspaceId
}

func (c *Client) setupHttpClient() error {
	resolveHost := func(addr string) (string, error) {
		// when using host namespace, we must still resolve the hostname inside the sandbox namespace as otherwise
		// we'll end up connecting to the dns proxy from the host namespace, which would fail
		ret, err := net.ResolveTCPAddr("", addr)
		if err != nil {
			return "", err
		}
		return ret.String(), nil
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var resolvedAddr string
		err := util.RunInNetNsOptional(c.resolveNs, func() error {
			var err error
			resolvedAddr, err = resolveHost(addr)
			return err
		})
		if err != nil {
			return nil, err
		}

		var conn net.Conn
		err = util.RunInNetNsOptional(c.connNs, func() error {
			var err error
			conn, err = net.Dial(network, resolvedAddr)
			return err
		})
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	c.httpClient = &http.Client{
		Transport: transport,
	}
	return nil
}
