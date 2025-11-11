//go:build linux

package sandbox

import (
	"fmt"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/opencontainers/cgroups"
	"github.com/opencontainers/cgroups/devices/config"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

func (rn *Sandbox) buildSandboxContainerMounts() []*configs.Mount {
	mounts := []*configs.Mount{
		{
			Destination: "/proc",
			Device:      "proc",
			Source:      "proc",
			Flags:       unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_RELATIME,
		},
		{
			Destination: "/sys",
			Device:      "sysfs",
			Source:      "sysfs",
			Flags:       unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_RELATIME,
		},
		{
			Source:      "devtmpfs",
			Destination: "/dev",
			Device:      "devtmpfs",
			Flags:       unix.MS_NOSUID | unix.MS_STRICTATIME | unix.MS_RELATIME,
			Data:        "size=65536k",
		},
		{
			Destination: "/sys/fs/cgroup",
			Device:      "cgroup",
			Source:      "cgroup",
			Flags:       unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_RELATIME,
		},
		{
			Destination: "/dev/pts",
			Device:      "devpts",
			Source:      "devpts",
			Flags:       unix.MS_NOSUID | unix.MS_NOEXEC,
			Data:        "newinstance,ptmxmode=0666,mode=0620,gid=5",
		},
		{
			Destination: "/dev/shm",
			Device:      "tmpfs",
			Source:      "shm",
			Flags:       unix.MS_NOSUID | unix.MS_NOEXEC | unix.MS_NODEV,
			Data:        "mode=1777,size=65536k",
		},
		{
			Destination: "/dev/mqueue",
			Device:      "mqueue",
			Source:      "mqueue",
			Flags:       unix.MS_NOSUID | unix.MS_NOEXEC | unix.MS_NODEV,
		},
		{
			Destination: consts.ContainersDir,
			Device:      "bind",
			Source:      filepath.Join(rn.SandboxDir, "containers"),
			Flags:       unix.MS_BIND,
		},
		{
			Destination: consts.LogsDir,
			Device:      "bind",
			Source:      filepath.Join(rn.SandboxDir, "logs"),
			Flags:       unix.MS_BIND,
		},
		{
			Destination: consts.NetbirdDir,
			Device:      "bind",
			Source:      filepath.Join(rn.SandboxDir, "netbird"),
			Flags:       unix.MS_BIND,
		},
		{
			Destination: consts.VolumesDir,
			Device:      "rbind",
			Source:      filepath.Join(rn.SandboxDir, "volumes"),
			Flags:       unix.MS_BIND | unix.MS_REC | unix.MS_SHARED,
		},
	}

	return mounts
}

func (rn *Sandbox) buildSandboxContainerProcessSpec(image *v1.Image) (*libcontainer.Process, error) {
	var env []string
	env = append(env, image.Config.Env...)

	var args []string
	args = append(args, image.Config.Entrypoint...)
	args = append(args, image.Config.Cmd...)

	workingDir := image.Config.WorkingDir
	if workingDir == "" {
		workingDir = "/"
	}

	process := &libcontainer.Process{
		Args: args,
		Env:  env,
		Cwd:  workingDir,
		Init: true,
	}

	return process, nil
}

func (rn *Sandbox) buildSandboxContainerConfig(image *v1.Image) (*configs.Config, error) {
	namespaces := []configs.Namespace{
		{Type: configs.NEWNS},
		{Type: configs.NEWUTS},
		{Type: configs.NEWIPC},
		{Type: configs.NEWPID},
		{Type: configs.NEWCGROUP},
		{Type: configs.NEWNET, Path: filepath.Join("/run/netns", rn.network.NamesAndIps.SandboxNamespaceName)},
	}

	mounts := rn.buildSandboxContainerMounts()

	cg := &cgroups.Cgroup{
		Path: fmt.Sprintf(":dboxed:%s", rn.SandboxName),
		Resources: &cgroups.Resources{
			Devices: []*config.Rule{
				{
					Type:        config.CharDevice,
					Major:       config.Wildcard,
					Minor:       config.Wildcard,
					Allow:       true,
					Permissions: "rwm",
				},
				{
					Type:        config.BlockDevice,
					Major:       config.Wildcard,
					Minor:       config.Wildcard,
					Allow:       true,
					Permissions: "rwm",
				},
			},
		},
	}

	rlimits := []configs.Rlimit{
		{
			Type: unix.RLIMIT_NOFILE,
			Hard: uint64(1048576),
			Soft: uint64(1048576),
		},
		{
			Type: unix.RLIMIT_NPROC,
			Hard: uint64(unix.RLIM_INFINITY),
			Soft: uint64(unix.RLIM_INFINITY),
		},
		{
			Type: unix.RLIMIT_CORE,
			Hard: uint64(unix.RLIM_INFINITY),
			Soft: uint64(unix.RLIM_INFINITY),
		},
	}

	caps := rn.buildContainerCaps(true)

	config := &configs.Config{
		Version: specs.Version,
		Rootfs:  rn.GetSandboxRoot(),
		//Hostname: rn.BoxSpec.Hostname,
		Mounts:     mounts,
		Cgroups:    cg,
		Namespaces: namespaces,
		Rlimits:    rlimits,
		Capabilities: &configs.Capabilities{
			Bounding:  caps,
			Permitted: caps,
			Effective: caps,
		},
	}

	return config, nil
}
