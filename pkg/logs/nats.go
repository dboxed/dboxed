package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/logs/jsonlog"
	"github.com/koobox/unboxed/pkg/logs/multitail"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
	"log/slog"
	"sync"
	"time"
)

type TailToNats struct {
	ctx context.Context

	natsConn      *nats.Conn
	jetStream     jetstream.JetStream
	metadataMapKV jetstream.KeyValue
	logStreamName string
	logStream     jetstream.Stream
	logId         string

	MultiTail *multitail.MultiTail

	localMetadataMap map[string]string
	m                sync.Mutex
}

func NewTailToNats(ctx context.Context, natsUrl string, nkeySeed string, tailDbFile string, metadataKVName string, logStreamName string, logId string) (*TailToNats, error) {
	var natsOpts []nats.Option

	if nkeySeed != "" {
		nkPair, err := nkeys.FromSeed([]byte(nkeySeed))
		if err != nil {
			return nil, err
		}
		nkPub, err := nkPair.PublicKey()
		if err != nil {
			return nil, err
		}

		slog.InfoContext(ctx, "loaded nats nkey", slog.Any("nkey", nkPub))
		natsOpts = append(natsOpts, nats.Nkey(nkPub, nkPair.Sign))
	}

	nc, err := nats.Connect(natsUrl, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}
	doCloseNc := true
	defer func() {
		if doCloseNc {
			nc.Close()
		}
	}()
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	kv, err := js.KeyValue(ctx, metadataKVName)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata key-value store: %w", err)
	}

	logStream, err := js.Stream(ctx, logStreamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stream: %w", err)
	}

	ttn := &TailToNats{
		ctx:              ctx,
		natsConn:         nc,
		jetStream:        js,
		metadataMapKV:    kv,
		logStreamName:    logStreamName,
		logStream:        logStream,
		logId:            logId,
		localMetadataMap: map[string]string{},
	}

	ttn.MultiTail, err = multitail.NewMultiTail(ctx, tailDbFile, multitail.MultiTailOptions{
		LineBatchSize:    10,
		LineBatchLinger:  time.Millisecond * 100,
		LineBatchHandler: ttn.handleLineBatch,
	})
	if err != nil {
		return nil, err
	}

	doCloseNc = false

	return ttn, nil
}

func (ttn *TailToNats) putLogMetadata(metadata multitail.LogMetadata) (string, error) {
	ttn.m.Lock()
	defer ttn.m.Unlock()

	hash, ok := ttn.localMetadataMap[metadata.FileName]

	if !ok {
		hash = util.Sha256Sum([]byte(metadata.FileName))
		key := fmt.Sprintf("%s.%s", ttn.logId, hash)

		b, err := json.Marshal(metadata)
		if err != nil {
			return "", err
		}
		_, err = ttn.metadataMapKV.PutString(ttn.ctx, key, string(b))
		if err != nil {
			return "", err
		}

		ttn.localMetadataMap[metadata.FileName] = hash
	}
	return hash, nil
}

func (ttn *TailToNats) handleLineBatch(metadata multitail.LogMetadata, lines []*multitail.Line) error {
	hash, err := ttn.putLogMetadata(metadata)
	if err != nil {
		return err
	}
	sub := fmt.Sprintf("%s.%s.%s", ttn.logStreamName, ttn.logId, hash)

	buf := bytes.NewBuffer(nil)

	var nl jsonlog.JSONLogs

	for i, l := range lines {
		if l.Err != nil {
			continue
		}
		if i != 0 {
			buf.Write([]byte("\n"))
		}
		nl.Log = []byte(l.Line)
		nl.Created = l.Time
		err := nl.MarshalJSONBuf(buf)
		if err != nil {
			return err
		}
	}

	_, err = ttn.jetStream.Publish(ttn.ctx, sub, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
