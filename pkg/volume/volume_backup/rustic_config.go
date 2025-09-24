package volume_backup

type RusticConfig struct {
	Repository RusticConfigRepository `toml:"repository"`
}

type RusticConfigRepository struct {
	Repository string `toml:"repository"`
	Password   string `toml:"password"`

	Options RusticConfigRepositoryOptions `toml:"options"`
}

type RusticConfigRepositoryOptions struct {
	Endpoint string `toml:"endpoint"`
}
