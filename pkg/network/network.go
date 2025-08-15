package network

import (
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/vishvananda/netns"
)

const namePrefix = "ub"

type Network struct {
	Config types.NetworkConfig

	InfraContainerRoot string
	NamesAndIps        NamesAndIps

	HostNetworkNamespace netns.NsHandle
	NetworkNamespace     netns.NsHandle
}
