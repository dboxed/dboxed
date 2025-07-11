package network

import (
	"github.com/koobox/unboxed/pkg/types"
	"github.com/vishvananda/netns"
)

const namePrefix = "ub"

type Network struct {
	Config types.NetworkConfig

	NamesAndIps NamesAndIps

	HostNetworkNamespace netns.NsHandle
	NetworkNamespace     netns.NsHandle
}
