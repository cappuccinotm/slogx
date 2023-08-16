package slogm

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/netip"
	"runtime"
	"testing"
	"time"
)

func TestTrimAttrs(t *testing.T) {
	t.Run("no attrs", func(t *testing.T) {
		mw := TrimAttrs(10)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 0, rec.NumAttrs())
			return nil
		})

		err := h(context.Background(), slog.Record{})
		require.NoError(t, err)
	})

	t.Run("not limitable attr", func(t *testing.T) {
		mw := TrimAttrs(10)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, 12345678912.3456789, attr.Value.Float64())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", 12345678912.3456789)
		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("not oversized attr", func(t *testing.T) {
		mw := TrimAttrs(10)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "value", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", "value")

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized string attr", func(t *testing.T) {
		mw := TrimAttrs(5)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "value...", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", "value_very_long")

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized text marshaler attr", func(t *testing.T) {
		mw := TrimAttrs(11)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "127.127.127...", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		addr, err := netip.ParseAddr("127.127.127.127")
		require.NoError(t, err)
		rec.Add("key", addr)

		err = h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized json marshaler attr", func(t *testing.T) {
		mw := TrimAttrs(11)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some very l...", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", json.RawMessage("some very long string"))

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized byte slice attr", func(t *testing.T) {
		mw := TrimAttrs(11)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some very l...", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", []byte("some very long string"))

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized byte slice (custom named) attr", func(t *testing.T) {
		mw := TrimAttrs(11)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some very l...", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		type trickyByteSlice []uint8
		rec.Add("key", trickyByteSlice("some very long string"))

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})

	t.Run("oversized unserializable attr", func(t *testing.T) {
		mw := TrimAttrs(5)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "{a:12...", attr.Value.String())
				return true
			})
			return nil
		})

		type testStruct struct {
			a int
			b string
		}

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", testStruct{a: 1234567890, b: "abacaba"})

		err := h(context.Background(), rec)
		require.NoError(t, err)
	})
}
