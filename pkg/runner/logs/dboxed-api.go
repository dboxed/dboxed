package logs

import (
	"context"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/clients"
	multitail2 "github.com/dboxed/dboxed/pkg/runner/logs/multitail"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dustin/go-humanize"
)

type TailToApi struct {
	ctx context.Context

	client *baseclient.Client
	boxId  string

	MultiTail *multitail2.MultiTail
}

func NewTailToApi(ctx context.Context, c *baseclient.Client, tailDbFile string, boxId string) (*TailToApi, error) {
	ttn := &TailToApi{
		ctx:    ctx,
		client: c,
		boxId:  boxId,
	}

	var err error
	ttn.MultiTail, err = multitail2.NewMultiTail(ctx, tailDbFile, multitail2.MultiTailOptions{
		LineBatchBytesCount: 256 * humanize.KiByte,
		LineBatchLinger:     time.Millisecond * 100,
		LineBatchHandler:    ttn.handleLineBatch,
	})
	if err != nil {
		return nil, err
	}
	return ttn, nil
}

func (ttn *TailToApi) handleLineBatch(metadata boxspec.LogMetadata, lines []*multitail2.Line) error {
	req := models.PostLogs{
		Metadata: metadata,
	}

	req.Lines = make([]boxspec.LogsLine, 0, len(lines))
	for _, l := range lines {
		req.Lines = append(req.Lines, boxspec.LogsLine{
			Time: l.Time,
			Line: l.Line,
		})
	}

	c2 := clients.BoxClient{Client: ttn.client}
	err := c2.PostLogLines(ttn.ctx, ttn.boxId, req)
	if err != nil {
		return err
	}
	return nil
}
