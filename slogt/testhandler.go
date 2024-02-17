package slogt

import (
	"context"
	"log/slog"
)

// HandlerFunc is a function that implements Handler interface.
// If consumer uses it with slogx.Accumulator, then it can completely capture log records
// and check them in tests.
type HandlerFunc func(ctx context.Context, rec slog.Record) error

func (f HandlerFunc) Handle(ctx context.Context, rec slog.Record) error { return f(ctx, rec) }
func (f HandlerFunc) WithAttrs([]slog.Attr) slog.Handler                { return f }
func (f HandlerFunc) WithGroup(string) slog.Handler                     { return f }
func (f HandlerFunc) Enabled(context.Context, slog.Level) bool          { return true }
