package runc_exec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// getSubCgroupPaths parses --cgroup arguments, which can either be
//   - a single "path" argument (for cgroup v2);
//   - one or more controller[,controller[,...]]:path arguments (for cgroup v1).
//
// Returns a controller to path map. For cgroup v2, it's a single entity map
// with empty controller value.
func getSubCgroupPaths(args []string) (map[string]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	paths := make(map[string]string, len(args))
	for _, c := range args {
		// Split into controller:path.
		if ctr, path, ok := strings.Cut(c, ":"); ok {
			// There may be a few comma-separated controllers.
			for _, ctrl := range strings.Split(ctr, ",") {
				if ctrl == "" {
					return nil, fmt.Errorf("invalid --cgroup argument: %s (empty <controller> prefix)", c)
				}
				if _, ok := paths[ctrl]; ok {
					return nil, fmt.Errorf("invalid --cgroup argument(s): controller %s specified multiple times", ctrl)
				}
				paths[ctrl] = path
			}
		} else {
			// No "controller:" prefix (cgroup v2, a single path).
			if len(args) != 1 {
				return nil, fmt.Errorf("invalid --cgroup argument: %s (missing <controller>: prefix)", c)
			}
			paths[""] = c
		}
	}
	return paths, nil
}

type ExecOpts struct {
	Container     *libcontainer.Container
	IgnorePaused  bool
	Cgroup        []string
	ConsoleSocket string
	PidfdSocket   string
	Detach        bool
	PidFile       string
	PreserveFds   int

	Args           []string
	Cwd            string
	Apparmor       string
	ProcessLabel   string
	Caps           []string
	Env            []string
	Tty            bool
	NoNewPrivs     bool
	User           string
	AdditionalGids []int64
}

func ExecProcess(opts ExecOpts) (int, error) {
	status, err := opts.Container.Status()
	if err != nil {
		return -1, err
	}
	if status == libcontainer.Stopped {
		return -1, errors.New("cannot exec in a stopped container")
	}
	if status == libcontainer.Paused && !opts.IgnorePaused {
		return -1, errors.New("cannot exec in a paused container (use --ignore-paused to override)")
	}
	p, err := getProcess(opts)
	if err != nil {
		return -1, err
	}

	cgPaths, err := getSubCgroupPaths(opts.Cgroup)
	if err != nil {
		return -1, err
	}

	r := &runner{
		enableSubreaper: false,
		shouldDestroy:   false,
		container:       opts.Container,
		consoleSocket:   opts.ConsoleSocket,
		pidfdSocket:     opts.PidfdSocket,
		detach:          opts.Detach,
		pidFile:         opts.PidFile,
		action:          CT_ACT_RUN,
		init:            false,
		preserveFDs:     opts.PreserveFds,
		subCgroupPaths:  cgPaths,
	}
	return r.run(p)
}

func getProcess(opts ExecOpts) (*specs.Process, error) {
	p := &specs.Process{}
	args := opts.Args
	if len(args) < 2 {
		return nil, errors.New("exec args cannot be empty")
	}
	p.Args = args[1:]
	// Override the cwd, if passed.
	if cwd := opts.Cwd; cwd != "" {
		p.Cwd = cwd
	}
	if ap := opts.Apparmor; ap != "" {
		p.ApparmorProfile = ap
	}
	if l := opts.ProcessLabel; l != "" {
		p.SelinuxLabel = l
	}
	for _, c := range opts.Caps {
		p.Capabilities.Bounding = append(p.Capabilities.Bounding, c)
		p.Capabilities.Effective = append(p.Capabilities.Effective, c)
		p.Capabilities.Permitted = append(p.Capabilities.Permitted, c)
		// Since ambient capabilities can't be set without inherritable,
		// and runc exec --cap don't set inheritable, let's only set
		// ambient if we already have some inheritable bits set from spec.
		if p.Capabilities.Inheritable != nil {
			p.Capabilities.Ambient = append(p.Capabilities.Ambient, c)
		}
	}
	// append the passed env variables
	p.Env = opts.Env

	// Always set tty to false, unless explicitly enabled from CLI.
	p.Terminal = opts.Tty
	p.NoNewPrivileges = opts.NoNewPrivs
	// Override the user, if passed.
	if user := opts.User; user != "" {
		uids, gids, ok := strings.Cut(user, ":")
		if ok {
			gid, err := strconv.Atoi(gids)
			if err != nil {
				return nil, fmt.Errorf("bad gid: %w", err)
			}
			p.User.GID = uint32(gid)
		}
		uid, err := strconv.Atoi(uids)
		if err != nil {
			return nil, fmt.Errorf("bad uid: %w", err)
		}
		p.User.UID = uint32(uid)
	}
	for _, gid := range opts.AdditionalGids {
		if gid < 0 {
			return nil, fmt.Errorf("additional-gids must be a positive number %d", gid)
		}
		p.User.AdditionalGids = append(p.User.AdditionalGids, uint32(gid))
	}
	return p, validateProcessSpec(p)
}
