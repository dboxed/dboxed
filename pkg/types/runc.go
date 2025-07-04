package types

import "time"

type RuncState struct {
	OciVersion string    `json:"ociVersion"`
	Id         string    `json:"id"`
	Status     string    `json:"status"`
	Bundle     string    `json:"bundle"`
	Rootfs     string    `json:"rootfs"`
	Created    time.Time `json:"created"`
	Owner      string    `json:"owner"`
}
