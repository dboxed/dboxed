package boxes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/box_spec_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func (r *reconciler) getNats(ctx context.Context, box *dmodel.Box) (*nats.Conn, jetstream.JetStream, error) {
	q := querier.GetQuerier(ctx)
	config := config.GetConfig(ctx)
	ncp := global.GetNatsConnPool(ctx)

	w, err := dmodel.GetWorkspaceById(q, box.WorkspaceID, false)
	if err != nil {
		return nil, nil, err
	}

	nc, err := ncp.Connect(config.Nats.Url, w.NkeySeed)
	if err != nil {
		return nil, nil, err
	}
	doRelease := true
	defer func() {
		if doRelease {
			ncp.Release(nc)
		}
	}()

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, nil, err
	}

	doRelease = false
	return nc, js, nil
}

func (r *reconciler) reconcileNatsBoxSpec(ctx context.Context, box *dmodel.Box, log *slog.Logger) error {
	q := querier.GetQuerier(ctx)
	ncp := global.GetNatsConnPool(ctx)
	nc, js, err := r.getNats(ctx, box)
	if err != nil {
		return err
	}
	defer ncp.Release(nc)

	kv, err := js.KeyValue(ctx, nats_utils.BuildBoxSpecsKVStoreName(ctx, box.WorkspaceID))
	if err != nil {
		return err
	}

	var network *dmodel.Network
	if box.NetworkID != nil {
		network, err = dmodel.GetNetworkById(q, nil, *box.NetworkID, true)
		if err != nil {
			return err
		}
	}

	file, err := box_spec_utils.BuildBoxSpec(ctx, box, network)
	if err != nil {
		return err
	}
	b, err := json.Marshal(file)
	if err != nil {
		return err
	}

	_, err = kv.Put(ctx, fmt.Sprintf("box-spec-%d", box.ID), b)
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) reconcileDeleteNats(ctx context.Context, box *dmodel.Box, log *slog.Logger) error {
	ncp := global.GetNatsConnPool(ctx)
	nc, js, err := r.getNats(ctx, box)
	if err != nil {
		return err
	}
	defer ncp.Release(nc)

	err = r.reconcileDeleteNatsLogs(ctx, js, box, log)
	if err != nil {
		return err
	}
	err = r.reconcileDeleteNatsBoxSpec(ctx, js, box)
	if err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileDeleteNatsLogs(ctx context.Context, js jetstream.JetStream, box *dmodel.Box, log *slog.Logger) error {
	logsStream, err := js.Stream(ctx, nats_utils.BuildLogsStreamName(ctx, box.WorkspaceID))
	if err != nil {
		return err
	}

	logsSubject := nats_utils.BuildLogsSubjectName(ctx, box.WorkspaceID, box.ID, "*")
	log.InfoContext(ctx, "purging box logs", slog.Any("subject", logsSubject))
	err = logsStream.Purge(ctx, jetstream.WithPurgeSubject(logsSubject))
	if err != nil {
		return err
	}

	kv, err := js.KeyValue(ctx, nats_utils.BuildMetadataKVStoreName(ctx, box.WorkspaceID))
	if err != nil {
		return err
	}

	keysFilter := fmt.Sprintf("box-%d.*", box.ID)
	log.InfoContext(ctx, "purging logs metadata", slog.Any("keysFilter", keysFilter))
	kl, err := kv.ListKeysFiltered(ctx, keysFilter)
	if err != nil {
		return err
	}
	defer kl.Stop()
	for k := range kl.Keys() {
		slog.InfoContext(ctx, "deleting key", slog.Any("key", k))
		err = kv.Delete(ctx, k)
		if err != nil {
			return err
		}
	}
	err = kv.PurgeDeletes(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileDeleteNatsBoxSpec(ctx context.Context, js jetstream.JetStream, box *dmodel.Box) error {
	kv, err := js.KeyValue(ctx, nats_utils.BuildBoxSpecsKVStoreName(ctx, box.WorkspaceID))
	if err != nil {
		return err
	}

	err = kv.Delete(ctx, fmt.Sprintf("box-spec-%d", box.ID))
	if err != nil {
		return err
	}

	return nil
}
