package config

import (
	"fmt"
	"os"

	"github.com/dboxed/dboxed/pkg/util"
	"sigs.k8s.io/yaml"
)

type Config struct {
	InstanceName string `json:"instanceName"`

	Auth   AuthConfig   `json:"auth"`
	DB     DbConfig     `json:"db"`
	Server ServerConfig `json:"server"`

	DefaultWorkspaceQuotas DefaultWorkspaceQuotas `json:"defaultWorkspaceQuotas"`
}

type AuthConfig struct {
	Oidc *AuthOidcConfig `json:"oidc"`

	AdminUsers []string `json:"adminUsers"`
}

type AuthOidcConfig struct {
	IssuerUrl string `json:"issuerUrl"`
	ClientId  string `json:"clientId"`
}

type DbConfig struct {
	Url              string         `json:"url"`
	Migrate          bool           `json:"migrate"`
	SlowLogThreshold *util.Duration `json:"slowLogThreshold,omitempty"`
}

type ServerConfig struct {
	ListenAddress string `json:"listenAddress"`
	BaseUrl       string `json:"baseUrl"`
}

type DefaultWorkspaceQuotas struct {
	MaxLogBytes util.HumanBytes `json:"maxLogBytes"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("missing config path")
	}

	f, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		return nil, err
	}

	if config.InstanceName == "" {
		config.InstanceName = "dboxed"
	}

	return &config, nil
}
