package s3_bucket

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
}

type PrintS3Bucket struct {
	ID            string `col:"Id"`
	Endpoint      string `col:"Endpoint"`
	Bucket        string `col:"Bucket"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := clients.S3BucketsClient{Client: c}

	s3Buckets, err := c2.ListS3Buckets(ctx)
	if err != nil {
		return err
	}

	var table []PrintS3Bucket
	for _, s := range s3Buckets {
		p := PrintS3Bucket{
			ID:            s.ID,
			Endpoint:      s.Endpoint,
			Bucket:        s.Bucket,
			Status:        s.Status,
			StatusDetails: s.StatusDetails,
		}

		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
