package commandutils

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type ClientTool struct {
	Client *baseclient.Client

	workspaceCache map[int64]*models.Workspace
	boxCache       map[int64]*models.Box
}

func (ct *ClientTool) GetWorkspaceColumn(ctx context.Context, id int64) string {
	if ct.workspaceCache == nil {
		ct.workspaceCache = map[int64]*models.Workspace{}
	}
	w, ok := ct.workspaceCache[id]
	if !ok {
		var err error
		w, err = ct.Client.GetWorkspaceById(ctx, id)
		ct.workspaceCache[id] = w
		if err != nil {
			slog.WarnContext(ctx, "failed to retrieve workspace", slog.Any("error", err))
		}
	}

	var ret string
	if w != nil {
		ret = fmt.Sprintf("%s (id=%d)", w.Name, w.ID)
	} else {
		ret = fmt.Sprintf("<unknown> (id=%d)", id)
	}

	return ret
}

func (ct *ClientTool) GetBoxColumn(ctx context.Context, id int64) string {
	if ct.boxCache == nil {
		ct.boxCache = map[int64]*models.Box{}
	}

	c := clients.BoxClient{Client: ct.Client}

	w, ok := ct.boxCache[id]
	if !ok {
		var err error
		w, err = c.GetBoxById(ctx, id)
		ct.boxCache[id] = w
		if err != nil {
			slog.WarnContext(ctx, "failed to retrieve box", slog.Any("error", err))
		}
	}

	var ret string
	if w != nil {
		ret = fmt.Sprintf("%s (id=%d)", w.Name, w.ID)
	} else {
		ret = fmt.Sprintf("<unknown> (id=%d)", id)
	}

	return ret
}
