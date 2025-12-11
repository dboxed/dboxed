package flags

import (
	"context"
	"fmt"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
)

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode" group:"Global:"`

	ClientAuthFile *string `help:"Override client auth file. Defaults to ~/.dboxed/client-auth.yaml" type:"path" group:"Global:"`

	ApiUrl    *string `help:"Override API url" group:"Global:"`
	ApiToken  *string `help:"Override API token" group:"Global:"`
	Workspace *string `help:"Override workspace" group:"Global:"`

	WorkDir string `help:"dboxed work dir" default:"/var/lib/dboxed" group:"Global:"`
}

func (f *GlobalFlags) BuildClient(ctx context.Context, opts ...baseclient.ClientOpt) (*baseclient.Client, error) {
	clientAuthFile := f.ClientAuthFile
	isSandboxClientAuth := false
	if clientAuthFile == nil {
		if _, err := os.Stat(consts.SandboxClientAuthFile); err == nil {
			clientAuthFile = util.Ptr(consts.SandboxClientAuthFile)
			isSandboxClientAuth = true
		}
	}

	clientAuth, err := baseclient.ReadClientAuth(clientAuthFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		clientAuth = &baseclient.ClientAuth{}
	}

	c, err := baseclient.New(clientAuthFile, clientAuth, !isSandboxClientAuth, opts...)
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

	if c.GetApiToken() == nil && clientAuth.Oauth2Token != nil {
		err = c.RefreshToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("oauth2 token is invalid, you might need to re-login: %w", err)
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
