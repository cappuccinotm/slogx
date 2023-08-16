package slogx

import (
	"context"

	"log/slog"
)

type requestIDKey struct{}

// ContextWithRequestID returns a new context with the given request ID.
func ContextWithRequestID(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, requestIDKey{}, reqID)
}

// RequestIDFromContext returns request id from context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(requestIDKey{}).(string)
	return v, ok
}

// RequestID returns a middleware that adds request id to record.
func RequestID() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if reqID, ok := RequestIDFromContext(ctx); ok {
				rec.AddAttrs(slog.String(RequestIDKey, reqID))
			}
			return next(ctx, rec)
		}
	}
}
