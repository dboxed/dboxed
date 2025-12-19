package dmodel

type VolumeProviderType string
type VolumeProviderStorageType string

const (
	VolumeProviderTypeRestic VolumeProviderType = "restic"
)

const (
	VolumeProviderStorageTypeS3 VolumeProviderStorageType = "s3"
)
