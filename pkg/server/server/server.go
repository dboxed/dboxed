package server

import (
	"context"
	"reflect"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes"
	"github.com/dboxed/dboxed/pkg/server/resources/dboxed_specs"
	"github.com/dboxed/dboxed/pkg/server/resources/git_credentials"
	"github.com/dboxed/dboxed/pkg/server/resources/healthz"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/server/resources/load_balancers"
	"github.com/dboxed/dboxed/pkg/server/resources/logs"
	"github.com/dboxed/dboxed/pkg/server/resources/machine_providers"
	"github.com/dboxed/dboxed/pkg/server/resources/machines"
	"github.com/dboxed/dboxed/pkg/server/resources/networks"
	"github.com/dboxed/dboxed/pkg/server/resources/s3buckets"
	"github.com/dboxed/dboxed/pkg/server/resources/s3proxy"
	"github.com/dboxed/dboxed/pkg/server/resources/tokens"
	"github.com/dboxed/dboxed/pkg/server/resources/users"
	"github.com/dboxed/dboxed/pkg/server/resources/volume_providers"
	"github.com/dboxed/dboxed/pkg/server/resources/volumes"
	"github.com/dboxed/dboxed/pkg/server/resources/workspaces"
	"github.com/gin-gonic/gin"
)

type DboxedServer struct {
	config config.Config

	authInfo     *models.AuthInfo
	oidcProvider *oidc.Provider

	ginEngine  *gin.Engine
	api        huma.API
	humaConfig huma.Config

	authMiddleware *auth_middleware.AuthMiddleware

	healthz          *healthz.HealthzServer
	auth             *auth.AuthHandler
	users            *users.Users
	tokens           *tokens.TokenServer
	workspaces       *workspaces.WorkspacesServer
	logs             *logs.LogsServer
	machineProviders *machine_providers.MachineProviderServer
	s3BucketsServer  *s3buckets.S3BucketsServer
	volumeProviders  *volume_providers.VolumeProviderServer
	s3Proxy          *s3proxy.S3ProxyServer
	networks         *networks.NetworksServer
	volumes          *volumes.VolumeServer
	boxes            *boxes.BoxesServer
	machines         *machines.MachinesServer
	loadBalancers    *load_balancers.LoadBalancerServer
	gitCredentials   *git_credentials.GitCredentialsServer
	dboxedSpecs      *dboxed_specs.DboedSpecsServer
}

func NewDboxedServer(ctx context.Context, config config.Config) (*DboxedServer, error) {
	s := &DboxedServer{
		config: config,
	}

	authInfo, oidcProvider, err := auth_middleware.BuildAuthProvider(ctx, config.Auth)
	if err != nil {
		return nil, err
	}
	s.authInfo = authInfo
	s.oidcProvider = oidcProvider

	s.authMiddleware = auth_middleware.NewAuthMiddleware(config.Auth, *authInfo, oidcProvider, true)

	s.healthz = healthz.New()
	s.auth = auth.NewAuthHandler(*authInfo, oidcProvider)
	s.users = users.New()
	s.tokens = tokens.New()
	s.workspaces = workspaces.New()
	s.logs = logs.New()
	s.machineProviders = machine_providers.New()
	s.s3BucketsServer = s3buckets.New()
	s.volumeProviders = volume_providers.New()
	s.s3Proxy = s3proxy.New()
	s.networks = networks.New()
	s.volumes = volumes.New(config)
	s.boxes = boxes.New(config)
	s.machines = machines.New(config)
	s.loadBalancers = load_balancers.New(config)
	s.gitCredentials = git_credentials.New()
	s.dboxedSpecs = dboxed_specs.New()

	return s, nil
}

func (s *DboxedServer) InitApi(ctx context.Context) error {
	var err error

	s.api.UseMiddleware(s.authMiddleware.AuthMiddleware(s.api))

	err = s.healthz.Init(s.api)
	if err != nil {
		return err
	}

	err = s.auth.Init(s.api)
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
	workspacesGroup.UseMiddleware(s.workspaces.Middleware.WorkspaceMiddleware(s.api))
	workspacesGroup.UseSimpleModifier(huma_utils.MetadataModifier(huma_metadata.AllowWorkspaceToken, true))
	workspacesGroup.UseModifier(func(o *huma.Operation, next func(*huma.Operation)) {
		o.Parameters = append(o.Parameters, &huma.Param{
			Name:        "workspaceId",
			In:          "path",
			Description: "The workspace id",
			Required:    true,
			Schema:      huma.SchemaFromType(s.humaConfig.Components.Schemas, reflect.TypeOf("")),
		})
		next(o)
	})

	err = s.logs.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}

	err = s.tokens.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}

	err = s.machineProviders.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.s3BucketsServer.Init(s.api, workspacesGroup)
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
	err = s.loadBalancers.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.gitCredentials.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}
	err = s.dboxedSpecs.Init(s.api, workspacesGroup)
	if err != nil {
		return err
	}

	return nil
}
