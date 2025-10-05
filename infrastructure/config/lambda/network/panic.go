package network

import (
	"context"
	"log/slog"
)

// HandlePanic handles a Lambda panic, canceling the current context
func HandlePanic(ctx context.Context) {
	if err := recover(); err != nil {
		slog.ErrorContext(ctx, "Lambda handler panicked", "err", err)

		switch err := err.(type) {
		case error:
			_, cancelCtx := context.WithCancelCause(ctx)
			cancelCtx(err)
		default:
			_, cancelCtx := context.WithCancel(ctx)
			cancelCtx()
		}
	}
}
