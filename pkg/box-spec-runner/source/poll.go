package source

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/types"
)

type PollFunc func(ctx context.Context) ([]byte, error)

func NewPollSource(ctx context.Context, poll PollFunc, interval time.Duration) (*BoxSpecSource, error) {
	raw, err := poll(ctx)
	if err != nil {
		return nil, err
	}

	s := &BoxSpecSource{
		Chan:     make(chan *types.BoxFile),
		stopChan: make(chan struct{}),
	}
	err = s.trySetNewSpec(ctx, raw, true, false)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-s.stopChan:
				close(s.Chan)
				return
			case <-ctx.Done():
				close(s.Chan)
				return
			case <-time.After(interval):
				break
			}

			newRaw, err := poll(ctx)
			if err != nil {
				slog.WarnContext(ctx, "retrieving box spec failed", slog.Any("error", err))
				continue
			}

			_ = s.trySetNewSpec(ctx, newRaw, false, true)
		}
	}()
	return s, nil
}
