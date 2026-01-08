package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type MachineProviderHetzner struct {
	HetznerNetworkName string  `json:"hetznerNetworkName"`
	HetznerNetworkID   *int64  `json:"hetznerNetworkId"`
	HetznerNetworkZone *string `json:"hetznerNetworkZone"`
	HetznerNetworkCidr *string `json:"hetznerNetworkCidr"`
	CloudSubnetCidr    *string `json:"cloudSubnetCidr"`
	RobotSubnetCidr    *string `json:"robotSubnetCidr"`
	RobotVswitchID     *int64  `json:"robotVswitchId"`
}

type UpdateMachineProviderHetzner struct {
	CloudToken    *string `json:"cloudToken,omitempty"`
	RobotUsername *string `json:"robotUsername,omitempty"`
	RobotPassword *string `json:"robotPassword,omitempty"`
}

type HetznerLocation struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Description string  `json:"description"`
	Id          int64   `json:"id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Name        string  `json:"name"`
	NetworkZone string  `json:"network_zone"`
}

type HetznerServerType struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	Cores        int     `json:"cores"`
	Memory       float32 `json:"memory"`
	Disk         int     `json:"disk"`
	StorageType  string  `json:"storageType"`
	CPUType      string  `json:"cpuType"`
	Architecture string  `json:"architecture"`

	Pricings  []HetznerServerTypeLocationPricing `json:"pricings"`
	Locations []HetznerServerTypeLocation        `json:"locations"`
}

type HetznerServerTypeLocationPricing struct {
	Location string       `json:"location"`
	Hourly   HetznerPrice `json:"hourly"`
	Monthly  HetznerPrice `json:"monthly"`

	// IncludedTraffic is the free traffic per month in bytes
	IncludedTraffic uint64       `json:"includedTraffic"`
	PerTBTraffic    HetznerPrice `json:"perTBTraffic"`
}

type HetznerServerTypeLocation struct {
	Location    string              `json:"location"`
	Deprecation *HetznerDeprecation `json:"deprecation,omitempty"`
}

type HetznerDeprecation struct {
	Announced        time.Time `json:"announced"`
	UnavailableAfter time.Time `json:"unavailableAfter"`
}

type HetznerPrice struct {
	Net   string `json:"net"`
	Gross string `json:"gross"`
}

func MachineProviderHetznerFromDB(v dmodel.MachineProviderHetzner) *MachineProviderHetzner {
	return &MachineProviderHetzner{
		HetznerNetworkName: v.HetznerNetworkName.V,
		HetznerNetworkID:   v.Status.HetznerNetworkID,
		HetznerNetworkZone: v.Status.HetznerNetworkZone,
		HetznerNetworkCidr: v.Status.HetznerNetworkCidr,
		CloudSubnetCidr:    v.Status.CloudSubnetCidr,
		RobotSubnetCidr:    v.Status.RobotSubnetCidr,
		RobotVswitchID:     v.Status.RobotVswitchID,
	}
}
