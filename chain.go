package slogx

import (
	"context"

	"golang.org/x/exp/slog"
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
func (c *Chain) WithGroup(group string) slog.Handler {
	return &Chain{
		mws:     c.mws,
		Handler: c.Handler.WithGroup(group),
	}
}

// WithAttrs returns a new Chain with the given attributes.
func (c *Chain) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Chain{
		mws:     c.mws,
		Handler: c.Handler.WithAttrs(attrs),
	}
}
