// Package slogt provides functions for comfortable using of slog in tests.
package slogt

import (
	"fmt"
	"log/slog"
	"strings"
)

type testingOpts struct {
	splitMultiline bool
}

// TestingOpt is an option for Handler.
type TestingOpt func(*testingOpts)

// SplitMultiline enables splitting multiline messages into multiple log lines.
func SplitMultiline(opts *testingOpts) { opts.splitMultiline = true }

// Handler returns a slog.Handler, that directs all log messages to the
// t.Logf function with the "[slog]" prefix.
// It also shortens some common attributes, like "time" and "level" to "t" and "l"
// and truncates the time to "15:04:05.000" format.
func Handler(t testingT, topts ...TestingOpt) slog.Handler {
	t.Helper()

	options := testingOpts{}
	for _, opt := range topts {
		opt(&options)
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch {
			case a.Key == slog.TimeKey: // shorten full time to "15:04:05.000"
				tt := a.Value.Time()
				return slog.String("t", tt.Format("15:04:05.000"))
			case a.Key == slog.LevelKey: // shorten "level":"debug" to "l":"debug"
				return slog.String("l", a.Value.String())
			case a.Key == slog.SourceKey: // shorten "source":"full/path/to/file.go:123" to "s":"file.go:123"
				src := a.Value.Any().(*slog.Source)
				file := src.File[strings.LastIndex(src.File, "/")+1:]
				return slog.String("s", fmt.Sprintf("%s:%d", file, src.Line))
			case a.Key == slog.MessageKey && options.splitMultiline &&
				strings.Contains(a.Value.String(), "\n"): // print the multiline message to t.Log, instead of slog
				msg := a.Value.String()
				lines := strings.Split(msg, "\n")
				for _, line := range lines {
					t.Log(line)
				}

				return slog.String(slog.MessageKey, "message with newlines has been printed to t.Log")
			default:
				return a
			}
		},
	}
	return slog.NewTextHandler(tWriter{t: t}, handlerOpts)
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
