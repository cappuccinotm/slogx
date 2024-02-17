package slogx

import (
	"context"
	"log/slog"
	"testing"

	"github.com/cappuccinotm/slogx/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handlerFunc func(ctx context.Context, rec slog.Record) error

func (f handlerFunc) Handle(ctx context.Context, rec slog.Record) error { return f(ctx, rec) }
func (handlerFunc) WithAttrs([]slog.Attr) slog.Handler                  { return NopHandler() }
func (handlerFunc) WithGroup(string) slog.Handler                       { return NopHandler() }
func (handlerFunc) Enabled(context.Context, slog.Level) bool            { return true }

func TestAccumulator_Handle(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		lg := slog.New(slogt.Handler(t))
		lg.WithGroup("abacaba").Info("test")
	})

	t.Run("accumulate only attributes", func(t *testing.T) {
		acc := Accumulator(handlerFunc(func(ctx context.Context, rec slog.Record) error {
			var attrs []slog.Attr
			rec.Attrs(func(attr slog.Attr) bool {
				attrs = append(attrs, attr)
				return true
			})
			assert.Equal(t, []slog.Attr{
				slog.String("c", "3"),
				slog.String("d", "4"),
				slog.String("a", "1"),
				slog.String("b", "2"),
			}, attrs)
			return nil
		}))

		err := acc.
			WithAttrs([]slog.Attr{
				slog.String("c", "3"),
				slog.String("d", "4"),
			}).
			WithAttrs([]slog.Attr{
				slog.String("a", "1"),
				slog.String("b", "2"),
			}).
			Handle(context.Background(), slog.Record{})
		assert.NoError(t, err)
	})

	t.Run("accumulate groups and attributes", func(t *testing.T) {
		acc := Accumulator(handlerFunc(func(ctx context.Context, rec slog.Record) error {
			var attrs []slog.Attr
			rec.Attrs(func(attr slog.Attr) bool {
				attrs = append(attrs, attr)
				return true
			})
			if !assert.Equal(t, []slog.Attr{
				slog.Group("g1",
					slog.Group("g2",
						slog.String("c", "3"),
						slog.String("d", "4"),
					),
					slog.String("a", "1"),
					slog.String("b", "2"),
				),
			}, attrs) {
				require.NoError(t, slogt.Handler(t).Handle(ctx, rec))
			}
			return nil
		}))
		err := acc.WithGroup("g1").
			WithAttrs([]slog.Attr{
				slog.String("a", "1"),
				slog.String("b", "2"),
			}).
			WithGroup("g2").
			WithAttrs([]slog.Attr{
				slog.String("c", "3"),
				slog.String("d", "4"),
			}).Handle(context.Background(), slog.Record{})
		assert.NoError(t, err)
	})
}
