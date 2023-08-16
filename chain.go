package slogx

import (
	"context"

	"log/slog"
)

// Chain is a chain of middleware.
type Chain struct {
	mws []Middleware
	slog.Handler
}

// NewChain returns a new Chain with the given middleware.
func NewChain(base slog.Handler, mws ...Middleware) *Chain {
	return &Chain{mws: mws, Handler: base}
}

// Handle runs the chain of middleware and the handler.
func (c *Chain) Handle(ctx context.Context, rec slog.Record) error {
	h := c.Handler.Handle
	for i := len(c.mws) - 1; i >= 0; i-- {
		h = c.mws[i](h)
	}
	return h(ctx, rec)
}

// WithGroup returns a new Chain with the given group.
// It applies middlewares on the top-level handler.
func (c *Chain) WithGroup(group string) slog.Handler {
	return &groupHandler{
		group:   group,
		Handler: c,
	}
}

// WithAttrs returns a new Chain with the given attributes.
func (c *Chain) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Chain{
		mws:     c.mws,
		Handler: c.Handler.WithAttrs(attrs),
	}
}

type groupHandler struct {
	group string
	slog.Handler
}

func (h *groupHandler) WithGroup(group string) slog.Handler {
	return &groupHandler{
		group:   group,
		Handler: h,
	}
}

func (h *groupHandler) Handle(ctx context.Context, rec slog.Record) error {
	var attrs []any
	rec.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	r := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)
	r.AddAttrs(slog.Group(h.group, attrs...))

	return h.Handler.Handle(ctx, r)
}
