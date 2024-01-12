package slogm

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"log/slog"
)

// TrimAttrs returns a middleware that trims attributes to the provided limit.
// Works only with attributes of type String/[]byte or Any.
func TrimAttrs(limit int) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			var nattrs []slog.Attr
			hasOversizedAttrs := false
			rec.Attrs(func(attr slog.Attr) bool {
				nattr, trimmed := trim(limit, attr)
				nattrs = append(nattrs, nattr)
				hasOversizedAttrs = hasOversizedAttrs || trimmed
				return true
			})

			if !hasOversizedAttrs {
				return next(ctx, rec)
			}

			nrec := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)
			nrec.AddAttrs(nattrs...)

			return next(ctx, nrec)
		}
	}
}

func trim(limit int, attr slog.Attr) (res slog.Attr, trimmed bool) {
	attr.Value = attr.Value.Resolve()

	str, ok := stringValue(attr)
	if !ok {
		return attr, false
	}

	if len(str) > limit {
		str = str[:limit] + "..."
	}

	return slog.String(attr.Key, str), true
}
