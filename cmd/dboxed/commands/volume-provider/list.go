package volume_provider

import (
	"context"
	"os"
	"path"

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
	ct := commandutils.NewClientTool(c)

	repos, err := c2.ListVolumeProviders(ctx)
	if err != nil {
		return err
	}

	var table []PrintVolumeProvider
	for _, r := range repos {
		storage := ""
		switch r.Type {
		case dmodel.VolumeProviderTypeRustic:
			if r.Rustic != nil {
				switch r.Rustic.StorageType {
				case dmodel.VolumeProviderStorageTypeS3:
					if r.Rustic.S3BucketId != nil {
						s3BucketName := ct.S3Buckets.GetColumn(ctx, *r.Rustic.S3BucketId)
						storage = path.Join(s3BucketName, r.Rustic.StoragePrefix)
					}
				}
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
