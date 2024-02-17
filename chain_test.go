package slogx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCtxKey struct{}

func TestChain_Handle(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewChain(slog.NewJSONHandler(buf, nil),
		func(next HandleFunc) HandleFunc {
			return func(ctx context.Context, record slog.Record) error {
				record.AddAttrs(slog.String("a", "1"))
				assert.Equal(t, "val", ctx.Value(testCtxKey{}))
				assert.Equal(t, "test", record.Message)
				return next(ctx, record)
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx context.Context, record slog.Record) error {
				record.AddAttrs(slog.String("b", "2"))
				containsA := false
				record.Attrs(func(attr slog.Attr) bool {
					if attr.Key == "a" {
						containsA = true
						return false
					}
					return true
				})
				assert.True(t, containsA)
				return next(ctx, record)
			}
		},
	)

	ctx := context.WithValue(context.Background(), testCtxKey{}, "val")

	logger := slog.New(h)
	logger.InfoContext(ctx, "test")

	t.Log(buf.String())

	var entry struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		A     string `json:"a"`
		B     string `json:"b"`
	}

	require.NoError(t, json.NewDecoder(buf).Decode(&entry))
	assert.Equal(t, slog.LevelInfo.String(), entry.Level)
	assert.Equal(t, "test", entry.Msg)
	assert.Equal(t, "1", entry.A)
	assert.Equal(t, "2", entry.B)
}

func TestChain_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewChain(slog.NewJSONHandler(buf, nil),
		func(next HandleFunc) HandleFunc {
			return func(ctx context.Context, rec slog.Record) error {
				rec.Add(slog.String("x-request-id", "x-request-id"))
				return next(ctx, rec)
			}
		},
	).WithGroup("test-group")

	logger := slog.New(h)
	logger.Info("test", slog.String("a", "1"))

	t.Log(buf.String())

	var entry struct {
		TestGroup struct {
			A string `json:"a"`
		} `json:"test-group"`
		XRequestID string `json:"x-request-id"` // without accumulating - chain doesn't capture groups and attributes
	}

	require.NoError(t, json.NewDecoder(buf).Decode(&entry))
	assert.Equal(t, "1", entry.TestGroup.A)
	assert.Empty(t, entry.XRequestID, "")
}

func TestChain_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	h := NewChain(slog.NewJSONHandler(buf, nil)).
		WithAttrs([]slog.Attr{
			slog.String("a", "1"),
			slog.String("b", "2"),
		})

	logger := slog.New(h)
	logger.Info("test")

	t.Log(buf.String())

	var entry struct {
		A string `json:"a"`
		B string `json:"b"`
	}

	require.NoError(t, json.NewDecoder(buf).Decode(&entry))
	assert.Equal(t, "1", entry.A)
	assert.Equal(t, "2", entry.B)
}
