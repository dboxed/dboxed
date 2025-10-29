package huma_utils

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func InitHumaErrorOverride() {
	orig := huma.NewError
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		if status == http.StatusInternalServerError {
			for _, err := range errs {
				if querier.IsSqlNotFoundError(err) {
					status = http.StatusNotFound
					msg = err.Error()
				} else if querier.IsSqlConstraintViolationError(err) {
					status = http.StatusConflict
					msg = err.Error()
				} else {
					slog.Error("internal server error", slog.Any("error", err))
				}
			}
		}
		return orig(status, msg, errs...)
	}
}
