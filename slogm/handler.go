package slogm

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/cappuccinotm/slogx"
)

type applyHandlerOptions struct{ logError bool }

// ApplyHandlerOption is a functional option for ApplyHandler.
type ApplyHandlerOption func(*applyHandlerOptions)

// LogIntermediateError configures ApplyHandler to log errors from the handler instead of ignoring them.
func LogIntermediateError(opts *applyHandlerOptions) { opts.logError = true }

// ApplyHandler wraps slog.Handler as Middleware.
// Error from the handler is ignored, but the handler is called only if it is enabled for the record's level.
func ApplyHandler(handler slog.Handler, opts ...ApplyHandlerOption) slogx.Middleware {
	o := &applyHandlerOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if !handler.Enabled(ctx, rec.Level) {
				return next(ctx, rec)
			}

			if err := handler.Handle(ctx, rec); err != nil && o.logError {
				var pcs [1]uintptr
				runtime.Callers(1, pcs[:])

				errRec := slog.NewRecord(rec.Time, slog.LevelWarn,
					"[slogm.ApplyHandler] intermediate handler error", pcs[0])

				errRec.AddAttrs(slogx.Error(err))
				_ = next(ctx, errRec)
			}

			return next(ctx, rec)
		}
	}
}
