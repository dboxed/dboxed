package dboxed_specs

type DboxedSpecs struct {
	Volumes map[string]Volume `json:"volumes"`
	Boxes   map[string]Box    `json:"boxes"`
}

type Volume struct {
	Recreate string        `json:"recreate,omitempty"`
	Provider string        `json:"provider"`
	Rustic   *VolumeRustic `json:"rustic"`
}

type VolumeRustic struct {
	FsSize string `json:"fsSize"`
	FsType string `json:"fsType"`
}

type Box struct {
	Recreate string  `json:"recreate,omitempty"`
	Network  *string `json:"network,omitempty"`

	VolumeAttachments    []VolumeAttachment        `json:"volumeAttachments,omitempty"`
	ComposeProjects      map[string]ComposeProject `json:"composeProjects,omitempty"`
	LoadBalancerServices []LoadBalancerService     `json:"loadBalancerServices,omitempty"`
}

type VolumeAttachment struct {
	Volume string `json:"volume"`

	RootUid  *int64  `json:"rootUid,omitempty"`
	RootGid  *int64  `json:"rootGid,omitempty"`
	RootMode *string `json:"rootMode,omitempty"`
}

type ComposeProject struct {
	File string `json:"file"`
}

type LoadBalancerService struct {
	LoadBalancer string  `json:"loadBalancer"`
	Host         string  `json:"host"`
	PathPrefix   string  `json:"pathPrefix"`
	Port         int     `json:"port"`
	Description  *string `json:"description,omitempty"`
}

type OnlyRecreate struct {
	Recreate string `json:"recreate,omitempty"`
}
