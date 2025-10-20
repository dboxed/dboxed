package flags

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/pkg/baseclient"
)

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode"`

	ClientAuthFile *string `help:"Override client auth file. Defaults to ~/.dboxed/client-auth.yaml" type:"path"`

	ApiUrl    *string `help:"Override API url"`
	ApiToken  *string `help:"Override API token"`
	Workspace *string `help:"Override workspace"`

	WorkDir string `help:"dboxed work dir" default:"/var/lib/dboxed"`
}

func (f *GlobalFlags) BuildClient(ctx context.Context) (*baseclient.Client, error) {
	clientAuth, err := baseclient.ReadClientAuth(f.ClientAuthFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		clientAuth = &baseclient.ClientAuth{}
	}

	c, err := baseclient.New(f.ClientAuthFile, clientAuth, true)
	if err != nil {
		return nil, err
	}

	c.SetDebug(f.Debug)

	if f.ApiUrl != nil {
		c.SetOverrideApiUrl(*f.ApiUrl)
	}
	if f.ApiToken != nil {
		c.SetOverrideApiToken(*f.ApiToken)

		t, err := c.CurrentToken(ctx)
		if err != nil {
			return nil, err
		}
		if f.Workspace == nil {
			c.SetOverrideWorkspaceId(t.Workspace)
		}
	}
	if f.Workspace != nil {
		w, err := commandutils.GetWorkspace(ctx, c, *f.Workspace)
		if err != nil {
			return nil, err
		}

		c.SetOverrideWorkspaceId(w.ID)
	}

	return c, nil
}
