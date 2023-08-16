package slogm

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"log/slog"
	"reflect"
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

// byteSlice returns its argument as a []byte if the argument's
// underlying type is []byte, along with a second return value of true.
// Otherwise, it returns nil, false.
func byteSlice(a any) ([]byte, bool) {
	if bs, ok := a.([]byte); ok {
		return bs, true
	}
	// Like Printf's %s, we allow both the slice type and the byte element type to be named.
	t := reflect.TypeOf(a)
	if t != nil && t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return reflect.ValueOf(a).Bytes(), true
	}
	return nil, false
}
