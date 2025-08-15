package box_spec

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/dboxed/dboxed/pkg/types"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func NewNatsSource(ctx context.Context, url url.URL, nkeySeed []byte) (*BoxSpecSource, error) {
	bucket := url.Query().Get("bucket")
	if bucket == "" {
		return nil, fmt.Errorf("missing bucket in nats url")
	}
	key := url.Query().Get("key")
	if key == "" {
		return nil, fmt.Errorf("missing key in nats url")
	}

	kp, err := nkeys.FromSeed(nkeySeed)
	if err != nil {
		return nil, err
	}
	nkey, err := kp.PublicKey()
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "connecting to nats",
		slog.Any("url", url.String()),
		slog.Any("nkey", nkey),
	)
	nc, err := nats.Connect(url.String(), nats.Nkey(nkey, kp.Sign))
	if err != nil {
		return nil, err
	}
	doClose := true
	defer func() {
		if doClose {
			nc.Close()
		}
	}()

	js, err := nc.JetStream()
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

	doClose = false
	go func() {
		select {
		case <-s.stopChan:
			close(s.Chan)
			_ = keyWatcher.Stop()
			return
		case <-ctx.Done():
			close(s.Chan)
			_ = keyWatcher.Stop()
			return
		case kve = <-keyWatcher.Updates():
			_ = s.trySetNewSpec(ctx, kve.Value(), false, true)
			break
		}
	}()

	return s, nil
}
