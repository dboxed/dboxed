package source

import (
	"context"
	"os"
	"time"
)

func NewFileSource(ctx context.Context, path string, interval time.Duration) (*BoxSpecSource, error) {
	return NewPollSource(ctx, func(ctx context.Context) ([]byte, error) {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return b, nil
	}, interval)
}
