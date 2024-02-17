package slogx

import (
	"context"
	"log/slog"
)

type payload struct {
	group  string
	attrs  []slog.Attr
	parent *payload
}

type accumulator struct {
	slog.Handler
	last *payload
}

// Accumulator is a wrapper for slog.Handler that accumulates
// attributes and groups and passes them to the underlying handler
// only on Handle call, instead of logging them immediately.
func Accumulator(h slog.Handler) slog.Handler {
	return &accumulator{Handler: h}
}

// Handle accumulates attributes and groups and then calls the wrapped handler.
func (a *accumulator) Handle(ctx context.Context, rec slog.Record) error {
	if a.last != nil {
		rec.AddAttrs(a.assemble()...)
	}
	return a.Handler.Handle(ctx, rec)
}

// WithAttrs returns a new accumulator with the given attributes.
func (a *accumulator) WithAttrs(attrs []slog.Attr) slog.Handler {
	acc := *a // shallow copy
	if acc.last == nil {
		acc.last = &payload{}
	}
	acc.last.attrs = append(acc.last.attrs, attrs...)
	return &acc
}

// WithGroup returns a new accumulator with the given group.
func (a *accumulator) WithGroup(group string) slog.Handler {
	acc := *a // shallow copy
	acc.last = &payload{group: group, parent: acc.last}
	return &acc
}

func (a *accumulator) assemble() (attrs []slog.Attr) {
	for p := a.last; p != nil; p = p.parent {
		attrs = append(attrs, p.attrs...)
		if p.group != "" {
			attrs = []slog.Attr{slog.Group(p.group, listAny(attrs)...)}
		}
	}
	return attrs
}

func listAny(attrs []slog.Attr) []any {
	list := make([]any, len(attrs))
	for i, a := range attrs {
		list[i] = a
	}
	return list
}
