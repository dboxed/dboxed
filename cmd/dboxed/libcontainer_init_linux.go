//go:build linux

package main

import (
	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

func initLibcontainer() {
	// This is the golang entry point for runc init, executed
	// before main() but after libcontainer/nsenter's nsexec().
	libcontainer.Init()
}
