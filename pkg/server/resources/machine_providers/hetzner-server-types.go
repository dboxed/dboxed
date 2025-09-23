package machine_providers

import (
	_ "embed"
	"encoding/json"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

//go:embed hetzner-server-types.json
var hetznerServerTypesJson string

var hetznerServerTypes []hcloud.ServerType

func init() {
	err := json.Unmarshal([]byte(hetznerServerTypesJson), &hetznerServerTypes)
	if err != nil {
		panic(err)
	}
}
