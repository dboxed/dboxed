package boxes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/logs/multitail"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/nats-io/nats.go/jetstream"
)

func (s *BoxesServer) restListLogs(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[multitail.LogMetadata], error) {
	w := global.GetWorkspace(c)
	ncp := global.GetNatsConnPool(c)

	nc, err := ncp.Connect(s.config.Nats.Url, w.NkeySeed)
	if err != nil {
		return nil, err
	}
	defer ncp.Release(nc)

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	kv, err := js.KeyValue(c, nats_utils.BuildMetadataKVStoreName(c, w.ID))
	if err != nil {
		return nil, err
	}

	kl, err := kv.WatchFiltered(c, []string{fmt.Sprintf("box-%d.*", i.Id)}, jetstream.IgnoreDeletes())
	if err != nil {
		return nil, err
	}
	defer kl.Stop()

	var ret []multitail.LogMetadata
	for e := range kl.Updates() {
		if e == nil {
			break
		}
		b := e.Value()
		var m multitail.LogMetadata
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, m)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type sseLogsStreamInput struct {
	huma_utils.IdByPath

	File string `query:"file"`
	Seq  int64  `query:"seq"`
}

func (s *BoxesServer) sseLogsStream(c context.Context, i *sseLogsStreamInput, send sse.Sender) {
	err := s.sseLogsStreamErr(c, i, send)
	if err != nil {
		slog.ErrorContext(c, "error in sseLogsStreamErr", slog.Any("error", err))
		err = send.Data(models.LogsError{
			Message: err.Error(),
		})
		if err != nil {
			slog.ErrorContext(c, "error while sending sse error", slog.Any("error", err))
		}
	}
}

func (s *BoxesServer) sseLogsStreamErr(c context.Context, i *sseLogsStreamInput, send sse.Sender) error {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)
	ncp := global.GetNatsConnPool(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return err
	}

	nc, err := ncp.Connect(s.config.Nats.Url, w.NkeySeed)
	if err != nil {
		return err
	}
	defer ncp.Release(nc)

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	kv, err := js.KeyValue(c, nats_utils.BuildMetadataKVStoreName(c, w.ID))
	if err != nil {
		return err
	}
	logsSteam, err := js.Stream(c, nats_utils.BuildLogsStreamName(c, w.ID))
	if err != nil {
		return err
	}

	hash := util.Sha256Sum([]byte(i.File))
	kve, err := kv.Get(c, fmt.Sprintf("box-%d.%s", box.ID, hash))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return fmt.Errorf("log file %s not found", i.File)
		}
		return err
	}
	var m multitail.LogMetadata
	err = json.Unmarshal(kve.Value(), &m)
	if err != nil {
		return err
	}

	consumerConfig := jetstream.ConsumerConfig{
		DeliverPolicy: jetstream.DeliverAllPolicy,
		FilterSubjects: []string{
			nats_utils.BuildLogsSubjectName(c, w.ID, box.ID, hash),
		},
	}
	if i.Seq > 0 {
		consumerConfig.DeliverPolicy = jetstream.DeliverByStartSequencePolicy
		consumerConfig.OptStartSeq = uint64(i.Seq)
	}
	consumer, err := logsSteam.CreateConsumer(c, consumerConfig)
	if err != nil {
		return err
	}
	consumerInfo, err := consumer.Info(c)
	if err != nil {
		return err
	}
	defer func() {
		_ = logsSteam.DeleteConsumer(c, consumerInfo.Name)
	}()

	err = send.Data(m)
	if err != nil {
		return err
	}

	ch := make(chan jetstream.Msg)
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		ch <- msg
	})
	if err != nil {
		return err
	}
	defer cctx.Stop()

	for {
		select {
		case msg := <-ch:
			err := s.handleNatsLogMsg(msg, send)
			if err != nil {
				slog.ErrorContext(c, "error in handleNatsLogMsg", slog.Any("error", err))
			}
		case <-c.Done():
			return nil
		}
	}
}

func (s *BoxesServer) handleNatsLogMsg(msg jetstream.Msg, send sse.Sender) error {
	msgMeta, err := msg.Metadata()
	if err != nil {
		return err
	}

	batch, err := s.parseNatsLogsMsg(msg.Data())
	if err != nil {
		return err
	}
	batch.Seq = int64(msgMeta.Sequence.Stream)

	err = send.Data(batch)
	if err != nil {
		return err
	}
	err = msg.Ack()
	if err != nil {
		return err
	}
	return nil
}

func (s *BoxesServer) parseNatsLogsMsg(data []byte) (*boxspec.LogsBatch, error) {
	var ret boxspec.LogsBatch
	err := json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
