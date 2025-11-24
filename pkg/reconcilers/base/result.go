package base

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type ReconcileResult struct {
	Status      string
	Error       error
	UserMessage string
}

func StatusWithMessage(status string, message string) ReconcileResult {
	return ReconcileResult{Status: status, UserMessage: message}
}

func InternalError(err error) ReconcileResult {
	return ReconcileResult{Error: err, UserMessage: "internal error"}
}

func ErrorFromMessage(msg string, args ...any) ReconcileResult {
	msg = fmt.Sprintf(msg, args...)
	return ReconcileResult{Error: errors.New(msg), UserMessage: msg}
}

func ErrorWithMessage(err error, msg string, args ...any) ReconcileResult {
	msg = fmt.Sprintf(msg, args...)
	return ReconcileResult{Error: err, UserMessage: msg}
}

func SetReconcileResult[T dmodel.HasReconcileStatus](ctx context.Context, log *slog.Logger, v T, result ReconcileResult) {
	if result.Error != nil {
		log.ErrorContext(ctx, "error in reconcile", "error", result.Error, "userMessage", result.UserMessage)
		err := SetReconcileStatus(ctx, v, "Error", result.UserMessage)
		if err != nil && !querier2.IsSqlNotFoundError(err) {
			log.ErrorContext(ctx, "error in setReconcileStatus", "error", err)
		}
	} else if result.Status != "" {
		err := SetReconcileStatus(ctx, v, result.Status, result.UserMessage)
		if err != nil && !querier2.IsSqlNotFoundError(err) {
			log.ErrorContext(ctx, "error in setReconcileStatus", "error", err)
		}
	} else {
		err := SetReconcileStatus(ctx, v, "Ok", result.UserMessage)
		if err != nil && !querier2.IsSqlNotFoundError(err) {
			log.ErrorContext(ctx, "error in setReconcileStatus", "error", err)
		}
	}
}

func SetReconcileStatus[T dmodel.HasReconcileStatus](ctx context.Context, v T, status string, statusDetails string) error {
	v.SetReconcileStatus(status, statusDetails)
	return dmodel.UpdateReconcileStatus(querier2.GetQuerier(ctx), v)
}
