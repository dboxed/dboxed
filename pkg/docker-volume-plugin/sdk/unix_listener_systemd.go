//go:build (linux || freebsd) && !nosystemd

package sdk

import (
	"os"
	// "github.com/coreos/go-systemd/activation"
)

// isRunningSystemd checks whether the host was booted with systemd as its init
// system. This functions similarly to systemd's `sd_booted(3)`: internally, it
// checks whether /run/systemd/system/ exists and is a directory.
// http://www.freedesktop.org/software/systemd/man/sd_booted.html
//
// Copied from github.com/coreos/go-systemd/util.IsRunningSystemd
func isRunningSystemd() bool {
	fi, err := os.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}
	return fi.IsDir()
}
