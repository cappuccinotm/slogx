package slogx

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestStacktraceOnError(t *testing.T) {
	t.Run("in chain", func(t *testing.T) {
		buf := &bytes.Buffer{}
		h := NewChain(slog.NewJSONHandler(buf, nil), StacktraceOnError())

		slog.New(h).Error("something bad happened",
			slog.String("detail", "oh my! some error occurred"),
		)
		var entry struct {
			Level   string `json:"level"`
			Message string `json:"msg"`
			Detail  string `json:"detail"`
			Stack   string `json:"stacktrace"`
		}

		require.NoError(t, json.NewDecoder(buf).Decode(&entry))
		assert.Equal(t, slog.LevelError.String(), entry.Level)
		assert.Equal(t, "something bad happened", entry.Message)
		assert.Equal(t, "oh my! some error occurred", entry.Detail)

		t.Log("stacktrace:\n", entry.Stack)
		assert.Contains(t, entry.Stack, "github.com/cappuccinotm/slogx.TestStacktraceOnError")
		assert.NotContains(t, entry.Stack, "slogx/chain.go")
	})

	t.Run("error level", func(t *testing.T) {
		mw := StacktraceOnError()

		found := false
		fn := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, "oh my! some error occurred", rec.Message)
			assert.Equal(t, slog.LevelError, rec.Level)
			rec.Attrs(func(attr slog.Attr) bool {
				if attr.Key != "stacktrace" {
					return true
				}
				found = true

				v := attr.Value.String()
				assert.Contains(t, v, "github.com/cappuccinotm/slogx.TestStacktraceOnError")
				t.Log("stacktrace:\n", v)
				return false
			})
			return nil
		})

		err := fn(context.Background(), slog.Record{
			Level:   slog.LevelError,
			Message: "oh my! some error occurred",
		})
		require.NoError(t, err)

		assert.True(t, found)
	})

	t.Run("info level", func(t *testing.T) {
		mw := StacktraceOnError()

		fn := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, "everything is normal", rec.Message)
			assert.Equal(t, slog.LevelInfo, rec.Level)
			rec.Attrs(func(attr slog.Attr) bool {
				require.NotEqual(t, "stacktrace", attr.Key)
				return true
			})
			return nil
		})

		err := fn(context.Background(), slog.Record{
			Level:   slog.LevelInfo,
			Message: "everything is normal",
		})
		require.NoError(t, err)
	})
}
