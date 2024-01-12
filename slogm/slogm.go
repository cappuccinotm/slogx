package slogm

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
)

func stringValue(attr slog.Attr) (str string, ok bool) {
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
		return "", false
	}

	return str, true
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
