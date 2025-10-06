package global

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
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

func GetWorkspace(c context.Context) *dmodel.Workspace {
	return MustGet[*dmodel.Workspace](c, "workspace")
}
