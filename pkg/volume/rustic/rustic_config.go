package rustic

type RusticConfig struct {
	Repository RusticConfigRepository `toml:"repository"`
}

type RusticConfigRepository struct {
	Repository string `toml:"repository"`
	Password   string `toml:"password"`

	Options RusticConfigRepositoryOptions `toml:"options"`
}

type RusticConfigRepositoryOptions struct {
	Endpoint        string  `toml:"endpoint"`
	Bucket          string  `toml:"bucket,omitempty"`
	Region          *string `toml:"region,omitempty"`
	AccessKeyId     string  `toml:"access_key_id,omitempty"`
	SecretAccessKey string  `toml:"secret_access_key,omitempty"`
	Root            string  `toml:"root,omitempty"`
}
