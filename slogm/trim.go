package slogm

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/cappuccinotm/slogx"
	"log/slog"
	"reflect"
)

// TrimAttrs returns a middleware that trims attributes of String and Any kind
// to the provided limit.
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

	str := ""
	switch attr.Value.Kind() {
	case slog.KindString:
		str = attr.Value.String()
	case slog.KindAny:
		a := attr.Value.Any()

		if tm, ok := a.(encoding.TextMarshaler); ok {
			data, _ := tm.MarshalText()
			str = string(data)
			break
		}

		if jm, ok := a.(json.Marshaler); ok {
			data, _ := jm.MarshalJSON()
			str = string(data)
			break
		}

		if bs, ok := byteSlice(a); ok {
			str = string(bs)
			break
		}

		str = fmt.Sprintf("%+v", a)
	default:
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
