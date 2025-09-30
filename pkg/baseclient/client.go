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

func FromClientAuthFile(clientAuthFile *string) (*Client, error) {
	return FromClientAuthFile2(clientAuthFile, true)
}

func FromClientAuthFile2(clientAuthFile *string, writeClientAuth bool) (*Client, error) {
	clientAuth, err := ReadClientAuth(clientAuthFile)
	if err != nil {
		return nil, err
	}
	return FromClientAuth(clientAuthFile, clientAuth, writeClientAuth)
}

func FromClientAuth(clientAuthFile *string, clientAuth *ClientAuth, writeClientAuth bool) (*Client, error) {
	return &Client{
		clientAuthFile:  clientAuthFile,
		clientAuth:      clientAuth,
		writeClientAuth: writeClientAuth,
	}, nil
}
