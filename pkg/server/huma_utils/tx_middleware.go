package huma_utils

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const NoTx = "no-tx"

func SetupTxMiddlewares(ginEngine *gin.Engine, humaApi huma.API) {
	ginEngine.Use(func(c *gin.Context) {
		didPanic := true
		defer func() {
			txI, ok := c.Get("tx")
			if !ok {
				return
			}
			tx, ok := txI.(*sqlx.Tx)
			if !ok {
				panic("not a sql.Tx")
			}

			if didPanic {
				slog.ErrorContext(c, "rolling back due to panic")
				_ = tx.Rollback()
				return
			}

			statusI, ok := c.Get("status")
			if !ok {
				slog.ErrorContext(c, "missing status in context, rolling back")
				_ = tx.Rollback()
				return
			}
			status, ok := statusI.(int)
			if !ok {
				panic("status is not an int")
			}
			if status >= 200 && status < 300 {
				_ = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
		}()

		c.Next()
		didPanic = false
	})

	humaApi.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		if HasMetadataTrue(ctx, NoTx) {
			next(ctx)
			return
		}

		ginCtx := humagin.Unwrap(ctx)

		db := querier.GetDB(ctx.Context())

		tx, err := db.Beginx()
		if err != nil {
			huma.WriteErr(humaApi, ctx, http.StatusInternalServerError, "failed to begin transaction", err)
			return
		}
		ginCtx.Set("tx", tx)
		ctx = huma.WithValue(ctx, "tx", tx)

		next(ctx)

		ginCtx.Set("status", ctx.Status())
	})
}
