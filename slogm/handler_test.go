package slogm

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cappuccinotm/slogx/slogt"
)

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
