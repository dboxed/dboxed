package dockercli

import "time"

type DockerVolume struct {
	Name       string `json:"Name"`
	Mountpoint string `json:"Mountpoint"`
}

type DockerPS struct {
	Command      string `json:"Command"`
	CreatedAt    string `json:"CreatedAt"`
	ID           string `json:"ID"`
	Image        string `json:"Image"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Names        string `json:"Names"`
	Networks     string `json:"Networks"`
	Ports        string `json:"Ports"`
	RunningFor   string `json:"RunningFor"`
	Size         string `json:"Size"`
	State        string `json:"State"`
	Status       string `json:"Status"`
}

type DockerContainerConfig struct {
	ID      string    `json:"ID"`
	Created time.Time `json:"Created"`
	Name    string    `json:"Name"`
}

type DockerComposeListEntry struct {
	Name   string `json:"Name"`
	Status string `json:"Status"`
}
