package volume_provider

import (
	"context"
	"fmt"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type ListCmd struct {
}

type PrintVolumeProvider struct {
	ID      int64  `col:"Id"`
	Name    string `col:"Name"`
	Type    string `col:"Type"`
	Status  string `col:"Status"`
	Storage string `col:"Storage"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}

	repos, err := c2.ListVolumeProviders(ctx)
	if err != nil {
		return err
	}

	var table []PrintVolumeProvider
	for _, r := range repos {
		storage := ""
		switch r.Type {
		case dmodel.VolumeProviderTypeRustic:
			switch r.Rustic.StorageType {
			case dmodel.VolumeProviderStorageTypeS3:
				storage = fmt.Sprintf("%s/%s/%s", r.Rustic.StorageS3.Endpoint,
					r.Rustic.StorageS3.Bucket, r.Rustic.StorageS3.Prefix)
			}
		}

		table = append(table, PrintVolumeProvider{
			ID:      r.ID,
			Name:    r.Name,
			Type:    string(r.Type),
			Status:  r.Status,
			Storage: storage,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
