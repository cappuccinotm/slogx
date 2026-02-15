package slogm

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/cappuccinotm/slogx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cappuccinotm/slogx/slogt"
)

func TestApplyHandler(t *testing.T) {
	t.Run("intermediate handler raised an error - log it", func(t *testing.T) {
		rch := make(ChannelCapturer, 10)
		defer rch.Close()

		lg := slog.New(slogx.NewChain(slogt.Handler(t, slogt.SplitMultiline),
			ApplyHandler(slogt.HandlerFunc(func(ctx context.Context, rec slog.Record) error {
				return errors.New("handler failed")
			}), LogIntermediateError),
			Capture(rch),
		))
		lg.Info("test", slog.String("key", "value"))

		recs := rch.Records()
		require.Len(t, recs, 2)

		errRec := recs[0]
		assert.Equal(t, slog.LevelWarn, errRec.Level)
		assert.Equal(t, "[slogm.ApplyHandler] intermediate handler error", errRec.Message)
		assert.Equal(t, 1, errRec.NumAttrs())
		var errAttr slog.Attr
		errRec.Attrs(func(attr slog.Attr) bool { errAttr = attr; return true })
		assert.Equal(t, slogx.ErrorKey, errAttr.Key)
		assert.EqualError(t, errAttr.Value.Any().(error), "handler failed")
		// check that ptr is valid
		src := errRec.Source()
		assert.Equal(t, "github.com/cappuccinotm/slogx/slogm.ApplyHandler.func1.1", src.Function)

		mainRec := recs[1]
		assert.Equal(t, slog.LevelInfo, mainRec.Level)
		assert.Equal(t, "test", mainRec.Message)
		assert.Equal(t, 1, mainRec.NumAttrs())
		var keyAttr slog.Attr
		mainRec.Attrs(func(attr slog.Attr) bool { keyAttr = attr; return true })
		assert.Equal(t, "key", keyAttr.Key)
		assert.Equal(t, "value", keyAttr.Value.String())
	})

	t.Run("intermediate handler called", func(t *testing.T) {
		mw := ApplyHandler(slogt.HandlerFunc(func(ctx context.Context, rec slog.Record) error { return nil }))

		called := false
		err := mw(func(ctx context.Context, record slog.Record) error {
			called = true
			return nil
		})(context.Background(), slog.Record{})
		require.NoError(t, err)
		assert.True(t, called, "next middleware must be called")
	})
}
