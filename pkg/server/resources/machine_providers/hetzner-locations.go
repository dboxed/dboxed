package machine_providers

import (
	"encoding/json"

	"github.com/dboxed/dboxed/pkg/server/models"
)

var hetznerNetworkZones = []string{
	"eu-central",
	"us-east",
	"us-west",
	"ap-southeast",
}

var hetznerLocations []models.HetznerLocation

func init() {
	err := json.Unmarshal([]byte(_hetznerLocations), &hetznerLocations)
	if err != nil {
		panic(err)
	}
}

const _hetznerLocations = `
[
	{
		"city": "Falkenstein",
		"country": "DE",
		"description": "Falkenstein DC Park 1",
		"id": 1,
		"latitude": 50.47612,
		"longitude": 12.370071,
		"name": "fsn1",
		"network_zone": "eu-central"
	},
	{
		"city": "Nuremberg",
		"country": "DE",
		"description": "Nuremberg DC Park 1",
		"id": 2,
		"latitude": 49.452102,
		"longitude": 11.076665,
		"name": "nbg1",
		"network_zone": "eu-central"
	},
	{
		"city": "Helsinki",
		"country": "FI",
		"description": "Helsinki DC Park 1",
		"id": 3,
		"latitude": 60.169855,
		"longitude": 24.938379,
		"name": "hel1",
		"network_zone": "eu-central"
	},
	{
		"city": "Ashburn, VA",
		"country": "US",
		"description": "Ashburn, VA",
		"id": 4,
		"latitude": 39.045821,
		"longitude": -77.487073,
		"name": "ash",
		"network_zone": "us-east"
	},
	{
		"city": "Hillsboro, OR",
		"country": "US",
		"description": "Hillsboro, OR",
		"id": 5,
		"latitude": 45.54222,
		"longitude": -122.951924,
		"name": "hil",
		"network_zone": "us-west"
	},
	{
		"city": "Singapore",
		"country": "SG",
		"description": "Singapore",
		"id": 6,
		"latitude": 1.283333,
		"longitude": 103.833333,
		"name": "sin",
		"network_zone": "ap-southeast"
	}
]
`
