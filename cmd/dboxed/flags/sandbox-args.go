package flags

import (
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type SandboxRunArgs struct {
	Box         string  `help:"Specify box name or id" required:"" arg:""`
	SandboxName *string `help:"Override local sandbox name. Defaults to the box ID"`

	InfraImage string `help:"Specify the infra/sandbox image to use" default:"${default_infra_image}"`
	VethCidr   string `help:"CIDR to use for veth pairs. dboxed will dynamically allocate 2 IPs from this CIDR per box" default:"1.2.3.0/24"`
}

type SandboxArgsOptional struct {
	Sandbox *string `help:"Specify the local sandbox name, box name, box id, or box id" optional:"" arg:""`
}
type SandboxArgsRequired struct {
	Sandbox string `help:"Specify the local sandbox name, box name, box id, or box id" required:"" arg:""`
}

func (a SandboxRunArgs) GetSandboxName(box *models.Box) (string, error) {
	var ret string
	if a.SandboxName != nil {
		err := util.CheckName(*a.SandboxName)
		if err != nil {
			return "", err
		}
		ret = *a.SandboxName
	} else {
		ret = box.ID
	}
	return ret, nil
}
