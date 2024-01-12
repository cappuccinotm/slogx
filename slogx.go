// Package slogx contains extensions for standard library's slog package.
package slogx

import (
	"context"

	"log/slog"
)

// LogAttrStrategy specifies what to do with the attribute that
// is about to be logged.
type LogAttrStrategy uint8

const (
	// LogAttrNone means that the attribute should not be logged.
	LogAttrNone LogAttrStrategy = iota
	// LogAttrAsIs means that the attribute should be logged as is.
	LogAttrAsIs
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

// ErrAttrStrategy specifies how to log errors.
// "AsIs" logs nils, when the error is nil, if you want to not
// log nils, use "None".
var ErrAttrStrategy = LogAttrAsIs

// Error returns an attribute with error key.
func Error(err error) slog.Attr {
	switch {
	case err == nil && ErrAttrStrategy == LogAttrNone:
		return slog.Attr{}
	case err == nil && ErrAttrStrategy == LogAttrAsIs:
		return slog.String(ErrorKey, "nil")
	default:
		return slog.String(ErrorKey, err.Error())
	}
}

// NopHandler returns a slog.Handler, that does nothing.
func NopHandler() slog.Handler { return nopHandler{} }

type nopHandler struct{}

func (nopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (nopHandler) Handle(context.Context, slog.Record) error { return nil }
func (n nopHandler) WithAttrs([]slog.Attr) slog.Handler      { return n }
func (n nopHandler) WithGroup(string) slog.Handler           { return n }
