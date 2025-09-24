package volume_provider

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"sigs.k8s.io/yaml"
)

type ListCmd struct {
}

func (cmd *ListCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}

	repos, err := c2.ListVolumeProviders(ctx)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(repos)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(b)
	if err != nil {
		return err
	}
	return nil
}
