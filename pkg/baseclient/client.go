package baseclient

import (
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dboxed/dboxed/pkg/util"
)

const defaultApiUrl = "https://api.dboxed.io"

type Client struct {
	clientAuth      *ClientAuth
	writeClientAuth bool

	m        sync.Mutex
	provider *oidc.Provider
}

func New(url *string, writeClientAuth bool) (*Client, error) {
	if url == nil {
		url = util.Ptr(defaultApiUrl)
	}

	c := &Client{
		clientAuth: &ClientAuth{
			ApiUrl: *url,
		},
		writeClientAuth: writeClientAuth,
	}

	return c, nil
}

func FromClientAuthFile() (*Client, error) {
	return FromClientAuthFile2(true)
}

func FromClientAuthFile2(writeClientAuth bool) (*Client, error) {
	clientAuth, err := ReadClientAuth()
	if err != nil {
		return nil, err
	}
	return FromClientAuth(clientAuth, writeClientAuth)
}

func FromClientAuth(clientAuth *ClientAuth, writeClientAuth bool) (*Client, error) {
	return &Client{
		clientAuth:      clientAuth,
		writeClientAuth: writeClientAuth,
	}, nil
}
