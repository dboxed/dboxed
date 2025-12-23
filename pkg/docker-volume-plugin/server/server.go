package server

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	// "github.com/dboxed/dboxed/pkg/server/auth_middleware"
	// "github.com/dboxed/dboxed/commands/volume"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/volume"
	volume_mount "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-mount"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	volume_plugin "github.com/dboxed/dboxed/pkg/docker-volume-plugin"

	// "github.com/dboxed/dboxed/pkg/server/config"
	plugin_config "github.com/dboxed/dboxed/pkg/docker-volume-plugin/config"
	// "github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/healthz"
	// "github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/server/resources/logs"
	"github.com/gin-gonic/gin"
)

type PluginServer struct {
	config plugin_config.Config

	ginEngine  *gin.Engine
	api        huma.API
	humaConfig huma.Config

	// authMiddleware *auth_middleware.AuthMiddleware

	healthz *healthz.HealthzServer
	logs    *logs.LogsServer
}

func NewPluginServer(ctx context.Context, config plugin_config.Config) (*PluginServer, error) {
	s := &PluginServer{
		config: config,
	}

	// authInfo, oidcProvider, err := auth_middleware.BuildAuthProvider(ctx, plugin_config.Auth)
	// if err != nil {
	// 	return nil, err
	// }
	// s.authInfo = authInfo
	// s.oidcProvider = oidcProvider

	// s.authMiddleware = auth_middleware.NewAuthMiddleware(plugin_config.Auth, *authInfo, oidcProvider, true)

	// s.healthz = healthz.New()
	// s.logs = logs.New()

	return s, nil
}

func (s *PluginServer) InitApi(ctx context.Context) error {
	// 	var err error

	// 	s.api.UseMiddleware(s.authMiddleware.AuthMiddleware(s.api))

	// 	err = s.healthz.Init(s.api)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	workspacesGroup := huma.NewGroup(s.api, "/v1/workspaces/{workspaceId}")
	// 	workspacesGroup.UseMiddleware(s.workspaces.Middleware.WorkspaceMiddleware(s.api))
	// 	workspacesGroup.UseSimpleModifier(huma_utils.MetadataModifier(huma_metadata.AllowWorkspaceToken, true))
	// 	workspacesGroup.UseModifier(func(o *huma.Operation, next func(*huma.Operation)) {
	// 		o.Parameters = append(o.Parameters, &huma.Param{
	// 			Name:        "workspaceId",
	// 			In:          "path",
	// 			Description: "The workspace id",
	// 			Required:    true,
	// 			Schema:      huma.SchemaFromType(s.humaConfig.Components.Schemas, reflect.TypeOf("")),
	// 		})
	// 		next(o)
	// 	})

	// 	err = s.logs.Init(s.api, workspacesGroup)
	// 	if err != nil {
	// 		return err
	// 	}

	return nil
}

func Create(req *volume_plugin.CreateRequest) error {


	// req.Name
	// req.Options
	
	// TODO check if dboxed volume already exists before creating and validate
	cmd := &volume.CreateCmd{
		Name:           "default-volume",
		VolumeProvider: "rustic-test",
		FsType:         "btrfs",
		FsSize:         "",
	}

	gf := &flags.GlobalFlags{
		// TODO load from volume mount? or env?
	}

	//TODO: write metadata for later use?

	err := cmd.Run(gf)
	if err != nil {
		return err
	}

	// _, err = d.d.Create(req.Name, req.Options)
	return err
}

func List() (*volume_plugin.ListResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_plugin.ListResponse
	// ls, err := d.d.List()
	// if err != nil {
	// 	return &volume_plugin.ListResponse{}, err
	// }

	ls := []volume_plugin.Volume{} // TODO get actual list

	vols := make([]*volume_plugin.Volume, len(ls))

	// for i, v := range ls {
	// 	vol := &volume_plugin.Volume{
	// 		Name:       "",
	// 		Mountpoint: "",
	// 	}
	// 	vols[i] = vol
	// }
	res.Volumes = vols
	return res, nil
}

func Get(req *volume_plugin.GetRequest) (*volume_plugin.GetResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_plugin.GetResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_plugin.GetResponse{}, err
	// }

	var mountPoint string
	var status map[string]interface{}

	res.Volume = &volume_plugin.Volume{
		Name:       req.Name,
		Mountpoint: mountPoint,
		CreatedAt:  "", // TODO fill created at
		Status:     status,
	}
	return &volume_plugin.GetResponse{}, nil
}

func Remove(req *volume_plugin.RemoveRequest) error {
	// TODO is anything needed here? Might need to unmount?

	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return err
	// }
	// if err := d.d.Remove(v); err != nil {
	// 	return err
	// }
	return nil
}

func Path(req *volume_plugin.PathRequest) (*volume_plugin.PathResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_plugin.PathResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_plugin.PathResponse{}, err
	// }
	// res.Mountpoint = v.Path()
	return res, nil
}

func Mount(req *volume_plugin.MountRequest) (*volume_plugin.MountResponse, error) {

	// TODO: run serve

	args := &flags.VolumeServeArgs{
		Volume:         "test",
		BackupInterval: "30m",
	}

	cmd := &volume_mount.ServeCmd{
		args,
	}

	// TODO check if dboxed volume-mount already exists before creating?
	cmd := &volume_mount.CreateCmd{
		Volume: req.Name,
	}

	gf := &flags.GlobalFlags{
		// TODO load from volume mount? or env?
	}

	err := cmd.Run(gf)
	if err != nil {
		return nil, err
	}

	var res *volume_plugin.MountResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_plugin.MountResponse{}, err
	// }
	pth, err := v.Mount(req.ID)
	if err != nil {
		return &volume_plugin.MountResponse{}, err
	}
	res.Mountpoint = pth
	return res, nil
}

func Unmount(req *volume_plugin.UnmountRequest) error {
	// TODO check if dboxed volume already exists before creating
	cmd := &volume_mount.ReleaseCmd{
		Volume: req.Name,
	}

	gf := &flags.GlobalFlags{
		// TODO load from volume mount? or env?
	}

	err := cmd.Run(gf)
	if err != nil {
		return err
	}

	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return err
	// }
	if err := v.Unmount(req.ID); err != nil {
		return err
	}
	return nil
}

func Capabilities() *volume_plugin.CapabilitiesResponse {
	var res *volume_plugin.CapabilitiesResponse
	// res.Capabilities = volume_plugin.Capability{Scope: d.d.Scope()}
	return res
}
