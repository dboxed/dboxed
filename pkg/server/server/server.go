package server

import (
	"context"
	"reflect"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes"
	"github.com/dboxed/dboxed/pkg/server/resources/healthz"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/server/resources/machine_providers"
	"github.com/dboxed/dboxed/pkg/server/resources/machines"
	"github.com/dboxed/dboxed/pkg/server/resources/networks"
	"github.com/dboxed/dboxed/pkg/server/resources/s3proxy"
	"github.com/dboxed/dboxed/pkg/server/resources/tokens"
	"github.com/dboxed/dboxed/pkg/server/resources/users"
	"github.com/dboxed/dboxed/pkg/server/resources/volume_providers"
	"github.com/dboxed/dboxed/pkg/server/resources/volumes"
	"github.com/dboxed/dboxed/pkg/server/resources/workspaces"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type DboxedServer struct {
	config config.Config

	oidcProvider *oidc.Provider

	ginEngine  *gin.Engine
	api        huma.API
	humaConfig huma.Config

	healthz          *healthz.HealthzServer
	auth             *auth.AuthHandler
	users            *users.Users
	tokens           *tokens.TokenServer
	workspaces       *workspaces.Workspaces
	machineProviders *machine_providers.MachineProviderServer
	volumeProviders  *volume_providers.VolumeProviderServer
	s3Proxy          *s3proxy.S3ProxyServer
	networks         *networks.NetworksServer
	volumes          *volumes.VolumeServer
	boxes            *boxes.BoxesServer
	machines         *machines.MachinesServer
}

func NewDboxedServer(ctx context.Context, config config.Config) (*DboxedServer, error) {
	s := &DboxedServer{
		config: config,
	}

	var err error
	if config.Auth.OidcIssuerUrl != "" {
		s.oidcProvider, err = oidc.NewProvider(ctx, config.Auth.OidcIssuerUrl)
		if err != nil {
			return nil, err
		}
	}

	s.healthz = healthz.New(config)
	s.auth = auth.NewAuthHandler(config)
	s.users = users.New()
	s.tokens = tokens.New()
	s.workspaces = workspaces.New()
	s.machineProviders = machine_providers.New()
	s.volumeProviders = volume_providers.New()
	s.s3Proxy = s3proxy.New()
	s.networks = networks.New()
	s.volumes = volumes.New(config)
	s.boxes = boxes.New(config)
	s.machines = machines.New(config)

	return s, nil
}

func (s *DboxedServer) InitApi(ctx context.Context) error {
	var err error

	s.api.UseMiddleware(s.auth.AuthMiddleware)

	err = s.healthz.Init(s.api)
	if err != nil {
		return err
	}

	err = s.auth.Init(ctx, s.api)
	if err != nil {
		return err
	}

	err = s.users.Init(s.api)
	if err != nil {
		return err
	}

	err = s.workspaces.Init(s.api)
	if err != nil {
		return err
	}

	workspacesGroup := huma.NewGroup(s.api, "/v1/workspaces/{workspaceId}")
	workspacesGroup.UseMiddleware(s.workspaces.WorkspaceMiddleware)
	workspacesGroup.UseSimpleModifier(huma_utils.MetadataModifier(huma_metadata.AllowWorkspaceToken, true))
	workspacesGroup.UseModifier(func(o *huma.Operation, next func(*huma.Operation)) {
		o.Parameters = append(o.Parameters, &huma.Param{
			Name:        "workspaceId",
			In:          "path",
			Description: "The workspace id",
			Required:    true,
			Schema:      huma.SchemaFromType(s.humaConfig.Components.Schemas, reflect.TypeOf(0)),
		})
		next(o)
	})

	err = s.tokens.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}

	err = s.machineProviders.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.volumeProviders.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.s3Proxy.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.networks.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.volumes.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.boxes.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.machines.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}

	return nil
}
