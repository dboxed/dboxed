package global

type MachineProviderType string

const (
	MachineProviderHetzner MachineProviderType = "hetzner"
	MachineProviderAws     MachineProviderType = "aws"
)
