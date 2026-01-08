package machine_providers

import (
	_ "embed"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"sigs.k8s.io/yaml"
)

//go:embed hetzner-locations.yaml
var hetznerLocationsYaml []byte

var hetznerLocations []models.HetznerLocation

func init() {
	var orig []hcloud.Location

	err := yaml.Unmarshal(hetznerLocationsYaml, &orig)
	if err != nil {
		panic(err)
	}

	for _, x := range orig {
		hetznerLocations = append(hetznerLocations, models.HetznerLocation{
			City:        x.City,
			Country:     x.Country,
			Description: x.Description,
			Id:          x.ID,
			Latitude:    x.Latitude,
			Longitude:   x.Longitude,
			Name:        x.Name,
			NetworkZone: string(x.NetworkZone),
		})
	}
}
