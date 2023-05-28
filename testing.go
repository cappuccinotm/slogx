package slogx

import (
	"context"
	"strings"

	"golang.org/x/exp/slog"
)

// NopHandler returns a slog.Handler, that does nothing.
func NopHandler() slog.Handler { return nopHandler{} }

type nopHandler struct{}

func (nopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (nopHandler) Handle(context.Context, slog.Record) error { return nil }
func (n nopHandler) WithAttrs([]slog.Attr) slog.Handler      { return n }
func (n nopHandler) WithGroup(string) slog.Handler           { return n }

// TestHandler returns a slog.Handler, that directs all log messages to the
// t.Logf function with the "[slog]" prefix.
// It also shortens some common attributes, like "time" and "level" to "t" and "l"
// and truncates the time to "15:04:05.000" format.
func TestHandler(t testingT) slog.Handler {
	t.Helper()

	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				tt := a.Value.Time()
				return slog.String("t", tt.Format("15:04:05.000"))
			case slog.LevelKey:
				return slog.String("l", a.Value.String())
			case slog.SourceKey:
				// trim source to last 2 components
				s := a.Value.String()
				return slog.String("s", s[strings.LastIndex(s, "/")+1:])
			default:
				return a
			}
		},
	}
	return opts.NewTextHandler(tWriter{t})
}

type testingT interface {
	Log(args ...interface{})
	Helper()
}

type tWriter struct{ t testingT }

// Write directs the provided bytes to the t.Logf function with the "[slog]"
func (w tWriter) Write(p []byte) (n int, err error) {
	w.t.Helper()

	w.t.Log(string(p))
	return len(p), nil
}
