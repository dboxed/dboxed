package machines

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/resources/tokens"
	"github.com/dboxed/dboxed/pkg/util"
)

func ListMachineTokens(ctx context.Context, workspaceId string, machineId string) ([]dmodel.Token, error) {
	q := querier2.GetQuerier(ctx)
	prefix := BuildMachineTokenPrefix(machineId)
	return dmodel.ListTokensWithNamePrefix(q, workspaceId, prefix)
}

func GetFirstValidMachineToken(ctx context.Context, workspaceId string, machineId string) (*dmodel.Token, error) {
	tokens, err := ListMachineTokens(ctx, workspaceId, machineId)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no machine tokens for machine %s", machineId)
	}
	for _, t := range tokens {
		if t.ValidUntil == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("no valid machine tokens for machine %s", machineId)
}

func ListBoxTokens(ctx context.Context, workspaceId string, machineId string, boxId *string) ([]dmodel.Token, error) {
	q := querier2.GetQuerier(ctx)
	prefix := BuildBoxTokenNamePrefix(machineId, boxId)
	return dmodel.ListTokensWithNamePrefix(q, workspaceId, prefix)
}

func InvalidateBoxTokens(ctx context.Context, workspaceId string, machineId string, boxId *string) error {
	q := querier2.GetQuerier(ctx)
	oldTokens, err := ListBoxTokens(ctx, workspaceId, machineId, boxId)
	if err != nil {
		return err
	}
	for _, t := range oldTokens {
		if t.ValidUntil == nil {
			slog.InfoContext(ctx, "invalidating machine box token", "workspaceId", workspaceId, "machineId", machineId, "boxId", t.BoxID, "tokenId", t.ID, "tokenName", t.Name)
			err = t.UpdateValidUntil(q, util.Ptr(time.Now().Add(5*time.Minute)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func BuildMachineTokenPrefix(machineId string) string {
	prefix := tokens.InternalTokenNamePrefix + fmt.Sprintf("machine_%s_", machineId)
	return prefix
}

func BuildBoxTokenNamePrefix(machineId string, boxId *string) string {
	prefix := tokens.InternalTokenNamePrefix + fmt.Sprintf("machine_box_%s_", machineId)
	if boxId != nil {
		prefix += fmt.Sprintf("%s_", *boxId)
	}
	return prefix
}
