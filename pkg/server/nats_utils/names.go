package nats_utils

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dboxed/dboxed/pkg/server/config"
)

func BuildBoxSpecsUrl(ctx context.Context, workspaceId int64, boxId int64) string {
	config := config.GetConfig(ctx)
	u, err := url.Parse(config.Nats.Url)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("bucket", BuildBoxSpecsKVStoreName(ctx, workspaceId))
	q.Set("key", fmt.Sprintf("box-spec-%d", boxId))
	u.RawQuery = q.Encode()
	return u.String()
}

func BuildBoxSpecsKVStoreName(ctx context.Context, workspaceId int64) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-ws-%d-box-specs", config.InstanceName, workspaceId)
}

func BuildMetadataKVStoreName(ctx context.Context, workspaceId int64) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-ws-%d-logs-metadata", config.InstanceName, workspaceId)
}

func BuildLogsStreamName(ctx context.Context, workspaceId int64) string {
	config := config.GetConfig(ctx)
	return fmt.Sprintf("%s-ws-%d-logs", config.InstanceName, workspaceId)
}

func BuildLogsSubjectName(ctx context.Context, workspaceId int64, boxId int64, fileNameHash string) string {
	config := config.GetConfig(ctx)
	boxOrWildcard := "*"
	if boxId != -1 {
		boxOrWildcard = fmt.Sprintf("box-%d", boxId)
	}
	return fmt.Sprintf("%s-ws-%d-logs.%s.%s", config.InstanceName, workspaceId, boxOrWildcard, fileNameHash)
}
