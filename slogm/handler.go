package slogm

import (
	"context"
	"log/slog"

	"github.com/cappuccinotm/slogx"
)

// ApplyHandler wraps slog.Handler as Middleware.
// Error from the handler is ignored, but the handler is called only if it is enabled for the record's level.
func ApplyHandler(handler slog.Handler) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if handler.Enabled(ctx, rec.Level) {
				_ = handler.Handle(ctx, rec)
			}

			return next(ctx, rec)
		}
	}
}
