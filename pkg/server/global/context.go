package global

import (
	"context"
)

func GetPtr[T any](c context.Context, key string) *T {
	v := c.Value(key)
	if v == nil {
		return nil
	}

	v2, ok := v.(T)
	if !ok {
		panic("invalid value in context")
	}
	return &v2
}

func MustGet[T any](c context.Context, key string) T {
	v := c.Value(key)
	if v == nil {
		panic("missing value in context")
	}

	v2, ok := v.(T)
	if !ok {
		panic("invalid value in context")
	}
	return v2
}
