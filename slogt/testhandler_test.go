package slogt

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerFunc_Handle(t *testing.T) {
	called := 0
	f := HandlerFunc(func(ctx context.Context, rec slog.Record) error {
		called++
		assert.Empty(t, rec)
		return nil
	})
	assert.True(t, f.Enabled(context.Background(), slog.LevelInfo))
	require.NoError(t, f.Handle(context.Background(), slog.Record{}))
	require.NoError(t, f.WithGroup("group").Handle(context.Background(), slog.Record{}))
	require.NoError(t, f.WithAttrs([]slog.Attr{}).Handle(context.Background(), slog.Record{}))
	assert.Equal(t, 3, called)
}
