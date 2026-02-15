package slogm

import (
	"context"
	"log/slog"

	"github.com/cappuccinotm/slogx"
)

// ApplyHandler wraps slog.Handler as Middleware.
func ApplyHandler(handler slog.Handler) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if handler.Enabled(ctx, rec.Level) {
				return handler.Handle(ctx, rec)
			}

			return next(ctx, rec)
		}
	}
}
