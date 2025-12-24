package config

import (
	"fmt"
	"os"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dustin/go-humanize"
	"sigs.k8s.io/yaml"
)

//TODO add the following?
//   --api-url=API-URL                      Override API url ($DBOXED_API_URL)
//   --api-token=API-TOKEN                  Override API token ($DBOXED_API_TOKEN)
//   --workspace=WORKSPACE                  Override workspace ($DBOXED_WORKSPACE)
//   --work-dir="/var/lib/dboxed"           dboxed work dir ($DBOXED_WORK_DIR)

type Config struct {
	InstanceName           string                 `json:"instanceName"`
	Server                 ServerConfig           `json:"server"`
	DefaultWorkspaceQuotas DefaultWorkspaceQuotas `json:"defaultWorkspaceQuotas"`
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
		//TODO: Attempt to retrieve from environment / context?
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
	if config.DefaultWorkspaceQuotas.MaxLogBytes.Bytes == 0 {
		config.DefaultWorkspaceQuotas.MaxLogBytes.Bytes = humanize.MiByte * 100
	}

	return &config, nil
}
