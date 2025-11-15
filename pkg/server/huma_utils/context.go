package huma_utils

import (
	"context"

	"github.com/gin-gonic/gin"
)

func GetGinContext(ctx context.Context) *gin.Context {
	ginCtx := ctx.Value("ginContext")
	if ginCtx == nil {
		panic("missing ginContext")
	}
	ginCtx2, ok := ginCtx.(*gin.Context)
	if !ok {
		panic("invalid ginContext")
	}
	return ginCtx2
}
