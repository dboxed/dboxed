package config

import "context"

func GetConfig(ctx context.Context) *Config {
	i := ctx.Value("config")
	if i == nil {
		panic("no config in context")
	}
	cfg, ok := i.(*Config)
	if !ok {
		panic("config in context has invalid type")
	}
	return cfg
}
