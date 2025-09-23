package workspaces

import (
	"context"
	"errors"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
	"github.com/nats-io/nats.go/jetstream"
)

func (r *reconciler) reconcileNats(ctx context.Context, w *dmodel.Workspace) error {
	config := config.GetConfig(ctx)
	ncp := global.GetNatsConnPool(ctx)

	nc, err := ncp.Connect(config.Nats.Url, w.NkeySeed)
	if err != nil {
		return err
	}
	defer ncp.Release(nc)

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	_, err = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:   nats_utils.BuildBoxSpecsKVStoreName(ctx, w.ID),
		Replicas: config.Nats.Replicas,
	})
	if err != nil {
		return err
	}

	_, err = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:   nats_utils.BuildMetadataKVStoreName(ctx, w.ID),
		Replicas: config.Nats.Replicas,
	})
	if err != nil {
		return err
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     nats_utils.BuildLogsStreamName(ctx, w.ID),
		Subjects: []string{nats_utils.BuildLogsSubjectName(ctx, w.ID, -1, "*")},
		Replicas: config.Nats.Replicas,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) reconcileDeleteNats(ctx context.Context, w *dmodel.Workspace) error {
	config := config.GetConfig(ctx)
	ncp := global.GetNatsConnPool(ctx)

	nc, err := ncp.Connect(config.Nats.Url, w.NkeySeed)
	if err != nil {
		return err
	}
	defer ncp.Release(nc)

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	err = js.DeleteStream(ctx, nats_utils.BuildLogsStreamName(ctx, w.ID))
	if err != nil {
		if !errors.Is(err, jetstream.ErrStreamNotFound) {
			return err
		}
	}
	err = js.DeleteKeyValue(ctx, nats_utils.BuildMetadataKVStoreName(ctx, w.ID))
	if err != nil {
		if !errors.Is(err, jetstream.ErrBucketNotFound) {
			return err
		}
	}
	err = js.DeleteKeyValue(ctx, nats_utils.BuildBoxSpecsKVStoreName(ctx, w.ID))
	if err != nil {
		if !errors.Is(err, jetstream.ErrBucketNotFound) {
			return err
		}
	}

	return nil
}
