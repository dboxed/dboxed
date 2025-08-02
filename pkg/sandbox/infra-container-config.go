package sandbox

import (
	"encoding/json"
	"fmt"
	"github.com/dboxed/dboxed/pkg/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

func (rn *Sandbox) buildInfraContainerMounts(hostNetwork bool) []specs.Mount {
	mounts := []specs.Mount{
		{
			Destination: "/proc",
			Type:        "proc",
			Source:      "proc",
			Options:     []string{"rw", "nosuid", "nodev", "noexec", "relatime"},
		},
		{
			Destination: "/sys",
			Type:        "sysfs",
			Source:      "sysfs",
			Options:     []string{"rw", "nosuid", "nodev", "noexec", "relatime"},
		},
		{
			Source:      "devtmpfs",
			Destination: "/dev",
			Type:        "devtmpfs",
			Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
		},
		{
			Destination: "/sys/fs/cgroup",
			Type:        "cgroup",
			Source:      "cgroup",
			Options:     []string{"nosuid", "noexec", "nodev", "relatime", "rw"},
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
			Destination: "/hostfs",
			Type:        "rbind",
			Source:      "/",
			Options:     []string{"rbind"},
		},
		{
			Destination: "/var/lib/docker",
			Type:        "rbind",
			Source:      filepath.Join(rn.SandboxDir, "docker"),
			Options:     []string{"rbind"},
		},
		{
			Destination: types.LogsDir,
			Type:        "bind",
			Source:      filepath.Join(rn.SandboxDir, "logs"),
			Options:     []string{"bind"},
		},
	}

	if hostNetwork {
		mounts = append(mounts, specs.Mount{
			Destination: "/run/netns",
			Type:        "rbind",
			Source:      "/run/netns",
			Options:     []string{"rbind"},
		})
	}

	return mounts
}

func (rn *Sandbox) buildInfraContainerProcessSpec(image *v1.Image, cmd []string) (*specs.Process, error) {
	caps := rn.buildContainerCaps(true)

	usr := specs.User{} // root user

	var env []string
	env = append(env, image.Config.Env...)

	args := []string{"tini"}
	args = append(args, cmd...)

	workingDir := image.Config.WorkingDir
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
		User:            usr,
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

func (rn *Sandbox) buildInfraContainerOciSpec(image *v1.Image, hostNetwork bool, name string, cmd []string) (*specs.Spec, error) {
	process, err := rn.buildInfraContainerProcessSpec(image, cmd)
	if err != nil {
		return nil, err
	}

	namespaces := []specs.LinuxNamespace{
		{Type: specs.MountNamespace},
		{Type: specs.UTSNamespace},
		{Type: specs.IPCNamespace},
		{Type: specs.PIDNamespace},
		{Type: specs.CgroupNamespace},
	}
	if !hostNetwork {
		namespaces = append(namespaces, specs.LinuxNamespace{
			Type: specs.NetworkNamespace,
			Path: filepath.Join("/run/netns", rn.network.NamesAndIps.SandboxNamespaceName),
		})
	}

	mounts := rn.buildInfraContainerMounts(hostNetwork)

	spec := &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     rn.getInfraRoot(),
			Readonly: false,
		},
		Process: process,
		//Hostname: rn.BoxSpec.Hostname,
		Mounts: mounts,
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
			CgroupsPath: fmt.Sprintf(":unboxed:%s:%s", rn.SandboxName, name),
		},
	}

	return spec, nil
}

func (rn *Sandbox) writeInfraContainerOciSpec(name string, spec *specs.Spec) error {
	pth := filepath.Join(rn.getInfraContainerDir(name), "config.json")

	err := os.MkdirAll(filepath.Dir(pth), 0700)
	if err != nil {
		return err
	}

	b, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	err = os.WriteFile(pth, b, 0600)
	if err != nil {
		return err
	}
	return nil
}
