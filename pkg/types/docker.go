package types

import "time"

type DockerVolume struct {
	Name       string `json:"Name"`
	Mountpoint string `json:"Mountpoint"`
}

type DockerContainerConfig struct {
	ID      string    `json:"ID"`
	Created time.Time `json:"Created"`
	Name    string    `json:"Name"`
}
