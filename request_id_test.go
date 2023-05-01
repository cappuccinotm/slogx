package slogx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey{}, "test")
	mw := RequestID()
	found := false
	h := mw(func(ctx context.Context, rec slog.Record) error {
		rec.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "request_id" && attr.Value.String() == "test" {
				found = true
				return false
			}
			return true
		})
		return nil
	})

	err := h(ctx, slog.Record{})
	require.NoError(t, err)
	assert.True(t, found)
}

func TestContextWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithRequestID(ctx, "test")
	v, ok := ctx.Value(requestIDKey{}).(string)
	require.True(t, ok)
	assert.Equal(t, "test", v)
}

func TestRequestIDFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey{}, "test")
	v, ok := RequestIDFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "test", v)
}
