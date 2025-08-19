package source

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/types"
	"github.com/nats-io/nats.go"
)

func NewNatsSource(ctx context.Context, natsConn *nats.Conn, bucket string, key string) (*BoxSpecSource, error) {
	js, err := natsConn.JetStream()
	if err != nil {
		return nil, err
	}

	kv, err := js.KeyValue(bucket)
	if err != nil {
		return nil, err
	}

	keyWatcher, err := kv.Watch(key)
	if err != nil {
		return nil, err
	}

	var kve nats.KeyValueEntry
	for {
		nkve := <-keyWatcher.Updates()
		if nkve == nil {
			break
		}
		kve = nkve
	}
	if kve == nil {
		return nil, fmt.Errorf("nats key value store %s has no entry with key %s", bucket, key)
	}

	s := &BoxSpecSource{
		Chan:     make(chan *types.BoxFile),
		stopChan: make(chan struct{}),
	}
	err = s.trySetNewSpec(ctx, kve.Value(), true, false)
	if err != nil {
		return nil, err
	}

	go func() {
		defer func() {
			close(s.Chan)
			_ = keyWatcher.Stop()
		}()
		for {
			select {
			case <-s.stopChan:
				return
			case <-ctx.Done():
				return
			case kve = <-keyWatcher.Updates():
				slog.InfoContext(ctx, "received box spec update", slog.Any("key", kve.Key()), slog.Any("op", kve.Operation().String()))
				if kve.Operation() == nats.KeyValueDelete || kve.Operation() == nats.KeyValuePurge {
					return
				}
				_ = s.trySetNewSpec(ctx, kve.Value(), false, true)
				break
			}
		}
	}()

	return s, nil
}
