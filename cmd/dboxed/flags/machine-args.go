package flags

type MachineRunArgs struct {
	Machine string `help:"Specify machine name or id" required:"" arg:""`

	InfraImage string `help:"Specify the infra/sandbox image to use" default:"${default_infra_image}"`
	VethCidr   string `help:"CIDR to use for veth pairs. dboxed will dynamically allocate 2 IPs from this CIDR per box" default:"1.2.3.0/24"`
}
