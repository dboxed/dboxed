package docker_volume_plugin

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	plugin_config "github.com/dboxed/dboxed/pkg/docker-volume-plugin/config"
	"github.com/dboxed/dboxed/pkg/services"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
	volume_helper "github.com/docker/go-plugins-helpers/volume"
)

const pluginRoot = "/var/lib/dboxed/volumes"

type Driver struct{}

type driver_opts struct {
	volumeName     string
	volumeProvider string
	fsSize         string
	fsType         string
	backupInterval string
}

func parseVolumeOptions(options map[string]string) driver_opts {
	return driver_opts{
		volumeName:     options["volume_name"], // might not be needed, already a part of docker api request model?
		volumeProvider: options["volume_provider"],
		fsSize:         options["fs_size"],
		fsType:         options["fs_type"],
		backupInterval: options["backup_interval"],
	}
}

func buildClient(ctx context.Context, config *plugin_config.Config) (*baseclient.Client, error) {
	c, err := baseclient.New(nil, nil, false)
	return c, err
}

func (d *Driver) Create(req *volume_helper.CreateRequest) error {

	// Attempt to get + mount the volume, will create if necessary

	// backupInterval is only included on the initial Create request for plugin api, but would also be needed on mount.
	// instead, should there be a config.json for default backupInterval, and allow per-volume override elsewhere?
	// Options are environment, driver_opts, file-based overrides, or maybe something from dboxed UI?

	ctx := context.Background()
	opts := parseVolumeOptions(req.Options)
	config, err := plugin_config.LoadConfig("")

	c, err := buildClient(ctx, config)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	v, err := c2.GetVolumeByName(ctx, req.Name)
	if err != nil {
		return err
	}

	if v != nil {
		// if volume already matches requested options, return success? Or start serving immediately?
		// If volume doesn't match, throw error?
		return errors.New("volume already exists")
	}

	s := &services.VolumesService{Client: c}
	err = s.CreateVolume(&services.CreateVolumeCmdOpts{
		Name:           req.Name,
		VolumeProvider: opts.volumeProvider,
		FsType:         opts.fsType,
		FsSize:         opts.fsSize,
	})
	if err != nil {
		return err
	}

	err = s.RunServeVolumeCmd(ctx, pluginRoot, services.RunServeVolumeCmdOpts{
		Volume:  req.Name,
		Create:  false, //seems to refer to volumeState, not the volume itself
		Mount:   true,
		Serve:   false,
		Release: false,
	})
	if err != nil {
		return err
	}

	return err
}

func (d *Driver) Mount(req *volume_helper.MountRequest) (*volume_helper.MountResponse, error) {

	// serve the volume

	//TODO: How to handle if volume isn't already created/mounted?
	//TODO: do I need to do anything with req.ID?
	ctx := context.Background()
	volumeState, err := volume_serve.LoadVolumeState(filepath.Join(pluginRoot, req.Name))
	if err != nil {
		return nil, err
	}

	if volumeState != nil {
		//TODO: update volumeState with new options if necessary?
	}

	c, err := buildClient(ctx, plugin_config.GetConfig(ctx))
	if err != nil {
		return nil, err
	}

	s := &services.VolumesService{Client: c}

	// TODO: figure out how to handle backupInterval
	// TODO: need to run serve in background? how to hold onto a reference of this process for Unmount to stop later?
	err = s.RunServeVolumeCmd(ctx, pluginRoot, services.RunServeVolumeCmdOpts{
		Volume: req.Name,
		// BackupInterval:    &cmd.BackupInterval,
		// WebdavProxyListen: &cmd.WebdavProxyListen,
		Create:  true,
		Mount:   true,
		Serve:   true,
		Release: true,
	})

	return &volume_helper.MountResponse{Mountpoint: filepath.Join(pluginRoot, req.Name)}, nil
}

func (d *Driver) Unmount(req *volume_helper.UnmountRequest) error {
	//TODO: simply send stop request to volume serve?
	return nil
}

func (d *Driver) Remove(req *volume_helper.RemoveRequest) error {
	// get volumeState and remove mount, do nothing if volumeState is not found for some reason?

	volumeState, err := volume_serve.LoadVolumeState(filepath.Join(pluginRoot, req.Name))
	if err != nil {
		return err
	}

	if volumeState == nil {
		return nil
	}

	ctx := context.Background()
	c, err := buildClient(ctx, plugin_config.GetConfig(ctx))
	if err != nil {
		return nil
	}
	s := &services.VolumesService{Client: c}

	err = s.RunServeVolumeCmd(ctx, pluginRoot, services.RunServeVolumeCmdOpts{
		Volume: req.Name,
		// WebdavProxyListen: &cmd.WebdavProxyListen,
		Create:  false,
		Mount:   false,
		Serve:   false,
		Release: true,
	})
	if err != nil {
		return nil
	}

	return nil
}

func (d *Driver) Get(req *volume_helper.GetRequest) (*volume_helper.GetResponse, error) {
	//TODO: check locally stored metadata for volume info?

	// ctx := context.Background()
	mountPoint := filepath.Join(pluginRoot, req.Name)

	volumeState, err := volume_serve.LoadVolumeState(filepath.Join(pluginRoot, req.Name))
	if err != nil {
		return nil, err
	}

	var res *volume_helper.GetResponse

	var status map[string]interface{} //FIXME: What are available statuses?

	res.Volume = &volume_helper.Volume{
		Name:       req.Name,
		Mountpoint: mountPoint,
		CreatedAt:  volumeState.Volume.CreatedAt.Format(time.RFC3339), //TODO: verify format
		Status:     status,
	}
	return &volume_helper.GetResponse{}, nil
}

func (d *Driver) List() (*volume_helper.ListResponse, error) {
	volumeStates, err := volume_serve.ListVolumeState(pluginRoot)
	if err != nil {
		return nil, err
	}

	var res *volume_helper.ListResponse
	volumes := make([]*volume_helper.Volume, len(volumeStates))

	for _, vs := range volumeStates {
		var status map[string]interface{} //FIXME: What are available statuses?

		volumes = append(volumes, &volume_helper.Volume{
			Name:       vs.Volume.Name,
			Mountpoint: filepath.Join(pluginRoot, vs.Volume.Name),
			CreatedAt:  vs.Volume.CreatedAt.Format(time.RFC3339), //TODO: verify format
			Status:     status,
		})
	}

	res.Volumes = volumes
	return res, nil
}

func (d *Driver) Capabilities() *volume_helper.CapabilitiesResponse {
	return &volume_helper.CapabilitiesResponse{
		Capabilities: volume_helper.Capability{
			Scope: "local", // or "global" for multi-host plugins
		},
	}
}

func (d *Driver) Path(req *volume_helper.PathRequest) (*volume_helper.PathResponse, error) {
	//TODO: check volume state and return mount path

	// ctx := context.Background()
	_, err := volume_serve.LoadVolumeState(filepath.Join(pluginRoot, req.Name))
	if err != nil {
		return nil, err
	}

	// volumeState.Volume.Attachment?

	return &volume_helper.PathResponse{Mountpoint: filepath.Join(pluginRoot, req.Name)}, nil
}
