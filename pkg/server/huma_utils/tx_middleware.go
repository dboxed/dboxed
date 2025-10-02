package huma_utils

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const NoTx = "no-tx"

func SetupTxMiddlewares(ginEngine *gin.Engine, humaApi huma.API) {
	humaApi.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		needTx := false
		switch ctx.Method() {
		case "POST", "PATCH", "PUT", "DELETE":
			needTx = true
		}

		if !needTx || HasMetadataTrue(ctx, NoTx) {
			ctx = huma.WithContext(ctx, querier.WithForbidAutoTx(ctx.Context()))
			next(ctx)
			return
		}

		ginCtx := humagin.Unwrap(ctx)

		db := querier.GetDB(ctx.Context())

		err := db.Transaction(ctx.Context(), func(tx *sqlx.Tx) (bool, error) {
			ginCtx.Set("tx", tx)
			ctx = huma.WithValue(ctx, "tx", tx)

			next(ctx)

			if ctx.Status() < 200 || ctx.Status() >= 300 {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			huma.WriteErr(humaApi, ctx, http.StatusInternalServerError, "failed to run transaction", err)
			return
		}
	})
}
