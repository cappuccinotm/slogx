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

func TestMaskSecrets(t *testing.T) {
	t.Run("no attrs", func(t *testing.T) {
		mw := TrimAttrs(10)
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 0, rec.NumAttrs())
			return nil
		})

		err := h(context.Background(), slog.Record{})
		require.NoError(t, err)
	})

	t.Run("not-stringable attr", func(t *testing.T) {
		mw := MaskSecrets("***")
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

	t.Run("attr without secrets", func(t *testing.T) {
		mw := MaskSecrets("***")
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

		ctx := context.Background()
		ctx = AddSecrets(ctx, "secret")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("string attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "value_***_long", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", "value_secret_long")

		ctx := context.Background()
		ctx = AddSecrets(ctx, "secret")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("text marshaler attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "***.***", attr.Value.String())
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

		ctx := context.Background()
		ctx = AddSecrets(ctx, "127.127")

		err = h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("json marshaler attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some *** string", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", json.RawMessage("some very long string"))

		ctx := context.Background()
		ctx = AddSecrets(ctx, "very long")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("byte slice attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some *** string", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", []byte("some very long string"))

		ctx := context.Background()
		ctx = AddSecrets(ctx, "very long")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("byte slice (custom named) attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "some *** string", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		type trickyByteSlice []uint8
		rec.Add("key", trickyByteSlice("some very long string"))

		ctx := context.Background()
		ctx = AddSecrets(ctx, "very long")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("unserializable attr with secret", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "{a:1234567890 b:some***value}", attr.Value.String())
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
		rec.Add("key", testStruct{a: 1234567890, b: "somesecretvalue"})

		ctx := context.Background()
		ctx = AddSecrets(ctx, "secret")

		err := h(ctx, rec)
		require.NoError(t, err)
	})

	t.Run("secret added in the child context", func(t *testing.T) {
		mw := MaskSecrets("***")
		h := mw(func(ctx context.Context, rec slog.Record) error {
			assert.Equal(t, 1, rec.NumAttrs())
			rec.Attrs(func(attr slog.Attr) bool {
				assert.Equal(t, "key", attr.Key)
				assert.Equal(t, "value_***_***", attr.Value.String())
				return true
			})
			return nil
		})

		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		rec := slog.NewRecord(time.Now(), slog.LevelDebug, "test", pcs[0])
		rec.Add("key", "value_secret_long")

		ctx := context.Background()
		ctx = AddSecrets(ctx, "secret")

		AddSecrets(context.WithValue(ctx, "some key", "some value"), "long")

		err := h(ctx, rec)
		require.NoError(t, err)
	})
}
