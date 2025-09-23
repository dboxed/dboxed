//go:build linux

package network

import (
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/vishvananda/netns"
)

const namePrefix = "ub"

type Network struct {
	Config *boxspec.NetworkConfig

	InfraContainerRoot string
	NamesAndIps        NamesAndIps

	HostNetworkNamespace netns.NsHandle
	NetworkNamespace     netns.NsHandle
}
