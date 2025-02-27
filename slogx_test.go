package slogx

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cappuccinotm/slogx/slogt"
)

func TestError(t *testing.T) {
	t.Run("actual error", func(t *testing.T) {
		err := errors.New("test")
		attr := Error(err)
		assert.Equal(t, attr.Key, ErrorKey)
		assert.Equal(t, attr.Value.String(), err.Error())
	})

	t.Run("nil error", func(t *testing.T) {
		t.Run("LogAttrNone", func(t *testing.T) {
			ErrAttrStrategy = LogAttrNone
			defer func() {
				ErrAttrStrategy = LogAttrAsIs
			}()

			attr := Error(nil)
			assert.Equal(t, slog.Attr{}, attr)
		})

		t.Run("LogAttrAsIs", func(t *testing.T) {
			ErrAttrStrategy = LogAttrAsIs
			defer func() {
				ErrAttrStrategy = LogAttrAsIs
			}()

			attr := Error(nil)
			assert.Equal(t, attr.Key, ErrorKey)
			assert.Nil(t, attr.Value.Any())
		})
	})
}

func TestAttrs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		rec := slog.Record{}
		attrs := Attrs(rec)
		assert.Empty(t, attrs)
	})

	t.Run("non-empty", func(t *testing.T) {
		rec := slog.Record{}
		rec.AddAttrs(
			slog.String("a", "1"),
			slog.String("b", "2"),
		)
		attrs := Attrs(rec)
		assert.Equal(t, []slog.Attr{
			slog.String("a", "1"),
			slog.String("b", "2"),
		}, attrs)
	})
}

func TestApplyHandler(t *testing.T) {
	t.Run("error when handler failed", func(t *testing.T) {
		mw := ApplyHandler(slogt.HandlerFunc(func(ctx context.Context, rec slog.Record) error {
			return errors.New("handler failed")
		}))

		err := mw(func(ctx context.Context, record slog.Record) error {
			return nil
		})(context.Background(), slog.Record{})
		require.Error(t, errors.New("handler failed"), err)
	})

	t.Run("run next middleware", func(t *testing.T) {
		mw := ApplyHandler(slogt.HandlerFunc(func(ctx context.Context, rec slog.Record) error {
			return nil
		}))

		called := false

		err := mw(func(ctx context.Context, record slog.Record) error {
			called = true
			return nil
		})(context.Background(), slog.Record{})
		require.NoError(t, err)
		assert.True(t, called, "next middleware must be called")
	})
}
