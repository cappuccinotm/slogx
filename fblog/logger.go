// Package fblog provides slog handler and its options to print logs in
// the fblog-like style.
package fblog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Handler is a handler that prints logs in the fblog-like style, i.e.
// a new line for each attribute.
type Handler struct {
	out, err io.Writer

	lvl        slog.Level
	srcFormat  SourceFormat
	rep        func([]string, slog.Attr) slog.Attr
	maxKeySize int
	logTimeFmt string
	timeFmt    string

	// only for child handlers
	groups []string
	attrs  []slog.Attr

	// internal use
	lock *sync.Mutex
}

// NewHandler returns a new Handler.
// The format of the logs entries will be:
//
//		<Timestamp>      <Level>: <Message>
//		                 <Attr1>: <Value1>
//	            <Group1>.<Attr2>: <Value2>
//					     ...
func NewHandler(opts ...Option) *Handler {
	h := &Handler{
		out: os.Stdout, err: os.Stderr,
		lock:       &sync.Mutex{},
		lvl:        slog.LevelInfo,
		srcFormat:  SourceFormatNone,
		rep:        func(_ []string, a slog.Attr) slog.Attr { return a },
		logTimeFmt: "2006-01-02 15:04:05",
		timeFmt:    time.RFC3339,
		maxKeySize: HeaderKeySize,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Enabled returns true if the level is enabled.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool { return h.lvl <= level }

// Handle writes the record to the writer.
func (h *Handler) Handle(_ context.Context, rec slog.Record) error {
	e := newEntry(h.timeFmt, h.rep, rec.NumAttrs()+len(h.attrs)+1) // prealloc for source in case
	e.WriteHeader(h.logTimeFmt, h.maxKeySize, rec)
	rec.AddAttrs(h.attrs...)

	if rec.PC != 0 && h.srcFormat != SourceFormatNone {
		frames := runtime.CallersFrames([]uintptr{rec.PC})
		f, _ := frames.Next()

		switch h.srcFormat {
		case SourceFormatPos:
			f.File = f.File[strings.LastIndex(f.File, "/")+1:] // only the file name
			e.WriteAttr([]string{}, slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)))
		case SourceFormatFunc:
			e.WriteAttr([]string{}, slog.String(slog.SourceKey, f.Function))
		case SourceFormatLong:
			f.Function = f.Function[strings.LastIndex(f.Function, "/")+1:] // shorten func name to last pkg
			e.WriteAttr([]string{}, slog.String(slog.SourceKey,
				fmt.Sprintf("%s:%d:%s", f.File, f.Line, f.Function),
			))
		}
	}

	var err error
	rec.Attrs(func(attr slog.Attr) bool {
		e.WriteAttr(h.groups, attr)
		return true
	})
	if err != nil {
		return fmt.Errorf("write attributes: %w", err)
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if rec.Level >= slog.LevelWarn {
		if _, err = e.WriteTo(h.err); err != nil {
			return fmt.Errorf("write entry to the writer: %w", err)
		}
		return nil
	}

	if _, err = e.WriteTo(h.out); err != nil {
		return fmt.Errorf("write entry to the writer: %w", err)
	}

	return nil
}

// WithAttrs returns a new Handler with the given attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hh := *h // shallow copy
	hh.attrs = attrs
	return &hh
}

// WithGroup returns a new Handler with the given group.
func (h *Handler) WithGroup(name string) slog.Handler {
	hh := *h // shallow copy
	hh.groups = make([]string, len(h.groups), len(h.groups)+1)
	copy(hh.groups, h.groups)
	hh.groups = append(hh.groups, name)
	return &hh
}
