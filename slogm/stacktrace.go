package slogm

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"regexp"
	"runtime"

	"log/slog"
)

var reTrace = regexp.MustCompile(`.*/slog/logger\.go.*\n`)

// StacktraceOnError returns a middleware that adds stacktrace to record if level is error.
func StacktraceOnError() slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			if rec.Level != slog.LevelError {
				return next(ctx, rec)
			}

			stackInfo := make([]byte, 1024*1024)
			if stackSize := runtime.Stack(stackInfo, false); stackSize > 0 {
				traceLines := reTrace.Split(string(stackInfo[:stackSize]), -1)
				if len(traceLines) == 0 {
					return next(ctx, rec)
				}
				rec.AddAttrs(slog.String("stacktrace", traceLines[len(traceLines)-1]))
			}

			return next(ctx, rec)
		}
	}
}
