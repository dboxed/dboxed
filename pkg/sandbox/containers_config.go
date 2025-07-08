package sandbox

import (
	"fmt"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/moby/sys/user"
	"github.com/opencontainers/cgroups"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
	"path/filepath"
	"strings"
)

func (rn *Sandbox) buildMounts(c *types.ContainerSpec) []specs.Mount {
	var mounts []specs.Mount
	devMount := specs.Mount{
		Source:      "tmpfs",
		Destination: "/dev",
		Type:        "tmpfs",
		Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
	}
	if c.UseDevTmpFs {
		devMount.Source = "devtmpfs"
		devMount.Type = "devtmpfs"
	}
	mounts = append(mounts, devMount)

	mounts = append(mounts, []specs.Mount{
		{
			Destination: "/proc",
			Type:        "proc",
			Source:      "proc",
			Options:     nil,
		},
		{
			Destination: "/dev/pts",
			Type:        "devpts",
			Source:      "devpts",
			Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
		},
		{
			Destination: "/dev/shm",
			Type:        "tmpfs",
			Source:      "shm",
			Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
		},
		{
			Destination: "/dev/mqueue",
			Type:        "mqueue",
			Source:      "mqueue",
			Options:     []string{"nosuid", "noexec", "nodev"},
		},
		{
			Destination: "/sys",
			Type:        "sysfs",
			Source:      "sysfs",
			Options:     []string{"nosuid", "noexec", "nodev", "ro"},
		},
		{
			Destination: "/sys/fs/cgroup",
			Type:        "cgroup",
			Source:      "cgroup",
			Options:     []string{"nosuid", "noexec", "nodev", "relatime", "ro"},
		},
		{
			Destination: logs.RootLogDir,
			Type:        "bind",
			Source:      logs.RootLogDir,
			Options:     []string{"bind"},
		},
	}...)

	return mounts
}

func (rn *Sandbox) buildUserSpec(c *types.ContainerSpec, image *v1.Image) (*specs.User, error) {
	passwdPath, err := user.GetPasswdPath()
	if err != nil {
		return nil, err
	}
	groupPath, err := user.GetGroupPath()
	if err != nil {
		return nil, err
	}
	passwdPath = filepath.Join(rn.getContainerBundleDir(c.Name), strings.TrimPrefix(passwdPath, "/"))
	groupPath = filepath.Join(rn.getContainerBundleDir(c.Name), strings.TrimPrefix(groupPath, "/"))

	username := c.User
	if username == "" {
		username = image.Config.User
	}
	u, err := user.GetExecUserPath(username, nil, passwdPath, groupPath)
	if err != nil {
		return nil, err
	}

	ret := &specs.User{
		UID:            uint32(u.Uid),
		GID:            uint32(u.Gid),
		AdditionalGids: []uint32{uint32(u.Gid)},
	}
	for _, g := range u.Sgids {
		ret.AdditionalGids = append(ret.AdditionalGids, uint32(g))
	}

	return ret, nil
}

func (rn *Sandbox) buildProcessSpec(c *types.ContainerSpec, image *v1.Image) (*specs.Process, error) {
	caps := rn.buildContainerCaps(c)

	usr, err := rn.buildUserSpec(c, image)
	if err != nil {
		return nil, err
	}

	var env []string
	env = append(env, image.Config.Env...)
	env = append(env, c.Env...)

	entrypoint := image.Config.Entrypoint
	if c.Entrypoint != nil {
		entrypoint = c.Entrypoint
	}
	cmd := image.Config.Cmd
	if c.Cmd != nil {
		cmd = c.Cmd
	}
	args := append([]string{}, entrypoint...)
	args = append(args, cmd...)

	workingDir := image.Config.WorkingDir
	if c.WorkingDir != "" {
		workingDir = c.WorkingDir
	}
	if workingDir == "" {
		workingDir = "/"
	}

	rlimits := []specs.POSIXRlimit{
		{
			Type: "RLIMIT_NOFILE",
			Hard: uint64(1048576),
			Soft: uint64(1048576),
		},
		{
			Type: "RLIMIT_NPROC",
			Hard: uint64(unix.RLIM_INFINITY),
			Soft: uint64(unix.RLIM_INFINITY),
		},
		{
			Type: "RLIMIT_CORE",
			Hard: uint64(unix.RLIM_INFINITY),
			Soft: uint64(unix.RLIM_INFINITY),
		},
	}

	process := &specs.Process{
		User:            *usr,
		Args:            args,
		Env:             env,
		Cwd:             workingDir,
		NoNewPrivileges: false,
		Capabilities: &specs.LinuxCapabilities{
			Bounding:  caps,
			Permitted: caps,
			Effective: caps,
		},
		Rlimits: rlimits,
	}

	return process, nil
}

func (rn *Sandbox) buildOciSpec(c *types.ContainerSpec, image *v1.Image) (*specs.Spec, error) {
	process, err := rn.buildProcessSpec(c, image)
	if err != nil {
		return nil, err
	}

	namespaces := []specs.LinuxNamespace{
		{Type: specs.MountNamespace},
		{Type: specs.UTSNamespace},
		{Type: specs.IPCNamespace},
		{Type: specs.PIDNamespace},
	}
	if !c.HostNetwork {
		namespaces = append(namespaces, specs.LinuxNamespace{
			Type: specs.NetworkNamespace,
			Path: filepath.Join("/run/netns", rn.network.NetworkNamespaceName),
		})
	}
	if cgroups.IsCgroup2UnifiedMode() {
		namespaces = append(namespaces, specs.LinuxNamespace{
			Type: specs.CgroupNamespace,
		})
	}

	mounts := rn.buildMounts(c)

	spec := &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     rn.getContainerRoot(c.Name),
			Readonly: false,
		},
		Process:  process,
		Hostname: rn.BoxSpec.Hostname,
		Mounts:   mounts,
		Linux: &specs.Linux{
			MaskedPaths: []string{
				"/proc/acpi",
				"/proc/asound",
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/sys/firmware",
				"/proc/scsi",
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{
						Allow:  false,
						Access: "rwm",
					},
				},
			},
			Namespaces:  namespaces,
			CgroupsPath: fmt.Sprintf(":unboxed:%s:%s", rn.SandboxName, c.Name),
		},
	}

	return spec, nil
}
