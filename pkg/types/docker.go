package types

type DockerVolume struct {
	Name       string `json:"Name"`
	Mountpoint string `json:"Mountpoint"`
}
