package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/logs/multitail"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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

func NewTailToNats(ctx context.Context, natsConn *nats.Conn, tailDbFile string, metadataKVName string, logStreamName string, logId string) (*TailToNats, error) {
	js, err := jetstream.New(natsConn)
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
		natsConn:         natsConn,
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

	var batch types.LogsBatch
	batch.Lines = make([]types.LogsLine, 0, len(lines))

	for _, l := range lines {
		if l.Err != nil {
			continue
		}
		batch.Lines = append(batch.Lines, types.LogsLine{
			Line: l.Line,
			Time: l.Time,
		})
	}

	b, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	_, err = ttn.jetStream.Publish(ttn.ctx, sub, b)
	if err != nil {
		return err
	}
	return nil
}
