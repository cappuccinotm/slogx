package slogm

import (
	"encoding"
	"encoding/json"
	"fmt"
	"log/slog"
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
