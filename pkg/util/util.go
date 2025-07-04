package util

import (
	"context"
	"time"
)

func Ptr[T any](v T) *T {
	return &v
}

func SleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
