package docker_volume_plugin

import (
	// "fmt"
	"context"
	"log"
	"errors"
	// "os"
	// "path/filepath"

	// "github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	volume_mount "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-mount"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	volume_helper "github.com/docker/go-plugins-helpers/volume"
	// volume_plugin "github.com/dboxed/dboxed/pkg/docker-volume-plugin"
	// plugin_config "github.com/dboxed/dboxed/pkg/docker-volume-plugin/config"
)

const pluginRoot = "/var/lib/dboxed/volumes"

type Driver struct{}

type driver_opts struct {
}

func GetGlobals() *flags.GlobalFlags {
	// TODO load from env or config file?
	return &flags.GlobalFlags{}
}

func parseVolumeOptions(options map[string]string) driver_opts {
	parsedOptions := driver_opts{}
	return parsedOptions
}

func (d *Driver) Create(req *volume_helper.CreateRequest) error {
	// req.Name
	// req.Options

	gf := GetGlobals()

	ctx := context.Background()

	c, err := gf.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}
	v, err := c2.GetVolumeByName(ctx, req.Name)

	if v != nil {
		//TODO: check if volume is already mounted elsewhere... if mounted by self, return and give no error?
		return errors.New("volume already exists")
	}

	_, err = c2.CreateVolume(ctx, models.CreateVolume{
		Name: req.Name,
		// TODO other params?
	})

	return err

	// v.Attachment
	// v.MountStatus
	// v.VolumeProvider.Name

	// ct := commandutils.NewClientTool(c)

	// TODO check if dboxed volume already exists before creating and validate
	// cmd := &volume.CreateCmd{
	// 	Name:           req.Name,
	// 	VolumeProvider: "rustic-test",
	// 	FsType:         "btrfs",
	// 	// FsSize:         "",
	// }

	//TODO: write metadata for later use?

	// return cmd.Run(gf)

	// volPath := filepath.Join(pluginRoot, req.Name)
	// log.Printf("Create volume: %s", volPath)
	// return os.MkdirAll(volPath, 0755)
}

func (d *Driver) Mount(req *volume_helper.MountRequest) (*volume_helper.MountResponse, error) {

	// TODO: run serve in background?
	// args := &flags.VolumeServeArgs{
	// 	Volume:         "test",
	// 	BackupInterval: "30m",
	// }

	// cmd := &volume_mount.ServeCmd{
	// 	args,
	// }

	// TODO check if dboxed volume-mount already exists before creating?
	cmd := &volume_mount.CreateCmd{
		Volume: req.Name,
	}

	gf := GetGlobals()

	err := cmd.Run(gf)
	if err != nil {
		return nil, err
	}

	var res *volume_helper.MountResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_helper.MountResponse{}, err
	// }
	// pth, err := v.Mount(req.ID)
	// if err != nil {
	// 	return &volume_helper.MountResponse{}, err
	// }
	// res.Mountpoint = pth
	return res, nil

	// volPath := filepath.Join(pluginRoot, req.Name)

	// // Write a hello.txt file
	// helloFile := filepath.Join(volPath, "hello.txt")
	// err := os.WriteFile(helloFile, []byte("Hello, world!\n"), 0644)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to write hello.txt: %w", err)
	// }

	// log.Printf("Mount volume: %s -> %s", req.Name, volPath)
	// return &volume_helper.MountResponse{Mountpoint: volPath}, nil
}

func (d *Driver) Unmount(req *volume_helper.UnmountRequest) error {
	// TODO check if dboxed volume already exists before creating
	cmd := &volume_mount.ReleaseCmd{
		Volume: req.Name,
	}

	gf := GetGlobals()

	err := cmd.Run(gf)
	if err != nil {
		return err
	}

	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return err
	// }
	// if err := v.Unmount(req.ID); err != nil {
	// 	return err
	// }
	return nil

	// volPath := filepath.Join(pluginRoot, req.Name)
	// helloFile := filepath.Join(volPath, "hello.txt")

	// // Simulate cleanup
	// if err := os.Remove(helloFile); err != nil && !os.IsNotExist(err) {
	// 	return fmt.Errorf("unmount cleanup error: %w", err)
	// }

	// log.Printf("Unmount volume: %s (removed hello.txt)", req.Name)
	// return nil
}

func (d *Driver) Remove(req *volume_helper.RemoveRequest) error {
	// TODO is anything needed here? Might need to unmount?

	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return err
	// }
	// if err := d.d.Remove(v); err != nil {
	// 	return err
	// }

	// volPath := filepath.Join(pluginRoot, req.Name)

	// log.Printf("Remove volume: %s", volPath)
	// err := os.RemoveAll(volPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to remove volume: %w", err)
	// }

	return nil
}

func (d *Driver) Get(req *volume_helper.GetRequest) (*volume_helper.GetResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_helper.GetResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_helper.GetResponse{}, err
	// }

	var mountPoint string
	var status map[string]interface{}

	res.Volume = &volume_helper.Volume{
		Name:       req.Name,
		Mountpoint: mountPoint,
		CreatedAt:  "", // TODO fill created at
		Status:     status,
	}
	return &volume_helper.GetResponse{}, nil

	// volPath := filepath.Join(pluginRoot, req.Name)

	// Confirm it exists
	// info, err := os.Stat(volPath)
	// if os.IsNotExist(err) {
	// 	return nil, fmt.Errorf("volume %s not found", req.Name)
	// } else if err != nil || !info.IsDir() {
	// 	return nil, fmt.Errorf("invalid volume path: %w", err)
	// }

	// return &volume_helper.GetResponse{
	// 	Volume: &volume_helper.Volume{
	// 		Name:       req.Name,
	// 		Mountpoint: volPath,
	// 	},
	// }, nil
}

func (d *Driver) List() (*volume_helper.ListResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_helper.ListResponse

	ls := []volume_helper.Volume{} // TODO get actual list

	vols := make([]*volume_helper.Volume, len(ls))

	res.Volumes = vols

	return res, nil

	// entries, err := os.ReadDir(pluginRoot)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to list volumes: %w", err)
	// }

	// var volumes []*volume_helper.Volume
	// for _, entry := range entries {
	// 	if entry.IsDir() {
	// 		volPath := filepath.Join(pluginRoot, entry.Name())
	// 		volumes = append(volumes, &volume_helper.Volume{
	// 			Name:       entry.Name(),
	// 			Mountpoint: volPath,
	// 		})
	// 	}
	// }

	// return &volume_helper.ListResponse{Volumes: volumes}, nil
}

func (d *Driver) Capabilities() *volume_helper.CapabilitiesResponse {
	return &volume_helper.CapabilitiesResponse{
		Capabilities: volume_helper.Capability{
			Scope: "local", // or "global" for multi-host plugins
		},
	}
}

func (d *Driver) Path(req *volume_helper.PathRequest) (*volume_helper.PathResponse, error) {
	//TODO: check locally stored metadata for volume info?

	var res *volume_helper.PathResponse
	// v, err := d.d.Get(req.Name)
	// if err != nil {
	// 	return &volume_helper.PathResponse{}, err
	// }
	// res.Mountpoint = v.Path()
	return res, nil

	// volPath := filepath.Join(pluginRoot, req.Name)
	// return &volume_helper.PathResponse{Mountpoint: volPath}, nil
}

func main() {
	// if needed, do migrations if an older metadata structure is found?

	Driver := &Driver{}
	h := volume_helper.NewHandler(Driver)
	log.Print("Starting plugin ...")

	//TODO customize GID?
	if err := h.ServeUnix("dboxed", 0); err != nil {
		log.Fatalf("plugin serve error: %v", err)
	}
}
