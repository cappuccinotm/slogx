// Package slogx contains extensions for standard library's slog package.
package slogx

import (
	"context"

	"log/slog"
)

// Common used keys.
const (
	ErrorKey     = "error"
	RequestIDKey = "request_id"
)

// HandleFunc is a function that handles a record.
type HandleFunc func(context.Context, slog.Record) error

// Middleware is a middleware for logging handler.
type Middleware func(HandleFunc) HandleFunc

// Error returns an attribute with error key.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String(ErrorKey, "nil")
	}
	return slog.String(ErrorKey, err.Error())
}
