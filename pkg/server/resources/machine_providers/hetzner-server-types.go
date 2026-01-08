package machine_providers

import (
	_ "embed"
	"encoding/json"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

//go:embed hetzner-server-types.json
var hetznerServerTypesJson string

var hetznerServerTypes []models.HetznerServerType

func init() {
	var orig []hcloud.ServerType

	err := json.Unmarshal([]byte(hetznerServerTypesJson), &orig)
	if err != nil {
		panic(err)
	}

	for _, x := range orig {
		e := models.HetznerServerType{
			ID:           x.ID,
			Name:         x.Name,
			Description:  x.Description,
			Category:     x.Category,
			Cores:        x.Cores,
			Memory:       x.Memory,
			Disk:         x.Disk,
			StorageType:  string(x.StorageType),
			CPUType:      string(x.CPUType),
			Architecture: string(x.Architecture),
		}
		for _, p := range x.Pricings {
			e.Pricings = append(e.Pricings, models.HetznerServerTypeLocationPricing{
				Location:        p.Location.Name,
				Hourly:          convertHetznerPrice(p.Hourly),
				Monthly:         convertHetznerPrice(p.Monthly),
				IncludedTraffic: p.IncludedTraffic,
				PerTBTraffic:    convertHetznerPrice(p.PerTBTraffic),
			})
		}
		hetznerServerTypes = append(hetznerServerTypes, e)
	}
}

func convertHetznerPrice(p hcloud.Price) models.HetznerPrice {
	return models.HetznerPrice{
		Currency: p.Currency,
		VATRate:  p.VATRate,
		Net:      p.Net,
		Gross:    p.Gross,
	}
}
