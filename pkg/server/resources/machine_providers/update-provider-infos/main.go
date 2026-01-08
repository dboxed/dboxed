package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func main() {
	ctx := context.Background()

	err := updateAllInfos(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed: %s", err.Error())
		os.Exit(1)
	}
}

func updateAllInfos(ctx context.Context) error {
	err := updateHetznerInfos(ctx)
	if err != nil {
		return err
	}

	return nil
}

func updateHetznerInfos(ctx context.Context) error {
	hcloudToken := os.Getenv("HCLOUD_TOKEN")
	if hcloudToken == "" {
		return fmt.Errorf("missing HCLOUD_TOKEN")
	}

	hcloudClient := hcloud.NewClient(hcloud.WithToken(hcloudToken))

	locations, _, err := hcloudClient.Location.List(ctx, hcloud.LocationListOpts{})
	if err != nil {
		return err
	}

	err = util.AtomicWriteFileYaml("./hetzner-locations.yaml", locations, 0644)
	if err != nil {
		return err
	}

	return nil
}
