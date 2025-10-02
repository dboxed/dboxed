package flags

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/pkg/baseclient"
)

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode"`

	ClientAuthFile *string `help:"Override client auth file. Defaults to ~/.dboxed/client-auto.yaml" type:"path"`

	ApiUrl    *string `help:"Specify the API url"`
	ApiToken  *string `help:"Specify a static API token"`
	Workspace *string `help:"Specify workspace"`

	WorkDir string `help:"dboxed work dir" default:"/var/lib/dboxed"`
}

func (f *GlobalFlags) BuildClient(ctx context.Context) (*baseclient.Client, error) {
	clientAuth, err := baseclient.ReadClientAuth(f.ClientAuthFile)
	if err != nil {
		return nil, err
	}

	c, err := baseclient.New(f.ClientAuthFile, clientAuth, true)
	if err != nil {
		return nil, err
	}

	if f.ApiUrl != nil {
		c.SetOverrideApiUrl(*f.ApiUrl)
	}
	if f.ApiToken != nil {
		c.SetOverrideApiToken(*f.ApiToken)
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
