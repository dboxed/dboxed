package volume_provider

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintVolumeProvider struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	Type          string `col:"Type"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
	Storage       string `col:"Storage"`
	StoragePrefix string `col:"Storage Prefix"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}
	ct := commandutils.NewClientTool(c)

	vps, err := c2.ListVolumeProviders(ctx)
	if err != nil {
		return err
	}

	var table []PrintVolumeProvider
	for _, r := range vps {
		storage := ""
		storagePrefix := ""
		switch r.Type {
		case dmodel.VolumeProviderTypeRustic:
			if r.Rustic != nil {
				storagePrefix = r.Rustic.StoragePrefix
				switch r.Rustic.StorageType {
				case dmodel.VolumeProviderStorageTypeS3:
					if r.Rustic.S3BucketId != nil {
						storage = ct.S3Buckets.GetColumn(ctx, *r.Rustic.S3BucketId, false)
					}
				}
			}
		}

		table = append(table, PrintVolumeProvider{
			ID:            r.ID,
			Name:          r.Name,
			Type:          string(r.Type),
			Status:        r.Status,
			StatusDetails: r.StatusDetails,
			Storage:       storage,
			StoragePrefix: storagePrefix,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
