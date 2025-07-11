package sandbox

import (
	"encoding/json"
	"fmt"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

func (rn *Sandbox) buildInfraContainerMounts() []specs.Mount {
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
			Destination: "/var/lib/containerd",
			Type:        "rbind",
			Source:      filepath.Join(rn.SandboxDir, "containerd"),
			Options:     []string{"rbind"},
		},
		{
			Destination: "/run/netns",
			Type:        "rbind",
			Source:      "/run/netns",
			Options:     []string{"rbind"},
		},
		{
			Destination: "/var/log/unboxed",
			Type:        "bind",
			Source:      filepath.Join(rn.SandboxDir, "logs"),
			Options:     []string{"bind"},
		},
	}

	return mounts
}

func (rn *Sandbox) buildInfraContainerProcessSpec(image *v1.Image) (*specs.Process, error) {
	caps := rn.buildContainerCaps(true)

	usr := specs.User{} // root user

	var env []string
	env = append(env, image.Config.Env...)

	args := []string{"tini", "unboxed", "run-infra-sandbox"}

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

func (rn *Sandbox) buildInfraContainerOciSpec(image *v1.Image) (*specs.Spec, error) {
	process, err := rn.buildInfraContainerProcessSpec(image)
	if err != nil {
		return nil, err
	}

	// excluding network namespace
	namespaces := []specs.LinuxNamespace{
		{Type: specs.MountNamespace},
		{Type: specs.UTSNamespace},
		{Type: specs.IPCNamespace},
		{Type: specs.PIDNamespace},
		{Type: specs.CgroupNamespace},
	}

	mounts := rn.buildInfraContainerMounts()

	spec := &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     rn.getInfraRoot(),
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
			CgroupsPath: fmt.Sprintf(":unboxed:%s", rn.SandboxName),
		},
	}

	return spec, nil
}

func (rn *Sandbox) writeInfraContainerOciSpec(spec *specs.Spec) error {
	pth := filepath.Join(rn.getInfraContainerDir(), "config.json")

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
