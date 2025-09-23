package server_utils

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
)

func BuildBoxSpecNatsUrl(ctx context.Context, workspaceId int64, boxId int64) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s?bucket=%s&key=box-spec-%d", config.Nats.Url,
		nats_utils.BuildBoxSpecsKVStoreName(ctx, workspaceId), boxId)
}
