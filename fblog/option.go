package fblog

import (
	"io"
	"log/slog"
)

// Option is an option for a Handler.
type Option func(*Handler)

// WithLevel returns an Option that sets the level of the handler.
func WithLevel(lvl slog.Level) Option { return func(h *Handler) { h.lvl = lvl } }

// SourceFormat is the source format of the handler.
type SourceFormat uint8

const (
	// SourceFormatNone is the source format without any source.
	SourceFormatNone = 0

	// SourceFormatPos is the short source format,
	// e.g. "file:line".
	SourceFormatPos = 1

	// SourceFormatFunc is the short source format,
	// e.g. "func".
	SourceFormatFunc = 2

	// SourceFormatLong is the long source format,
	// e.g. "full/file/path:line:func".
	SourceFormatLong = 3
)

// WithSource returns an Option that sets the source of the handler.
func WithSource(srcFormat SourceFormat) Option { return func(h *Handler) { h.srcFormat = srcFormat } }

// WithReplaceAttrs returns an Option that sets the function that will replace the attributes.
func WithReplaceAttrs(r func([]string, slog.Attr) slog.Attr) Option {
	return func(h *Handler) { h.rep = r }
}

// WithLogTimeFormat returns an Option that sets the log's individual time format of the handler.
func WithLogTimeFormat(f string) Option { return func(h *Handler) { h.logTimeFmt = f } }

// WithTimeFormat returns an Option that sets the time format of the handler.
func WithTimeFormat(f string) Option { return func(h *Handler) { h.timeFmt = f } }

// Out sets the output writer of the handler.
func Out(w io.Writer) Option { return func(h *Handler) { h.out = w } }

// Err sets the error writer of the handler.
func Err(w io.Writer) Option { return func(h *Handler) { h.err = w } }

// Predefined key size options.
const (
	// HeaderKeySize looks for the header size of the log entry and
	// trims all the attribute keys to fit the header size.
	HeaderKeySize = 0
	// UnlimitedKeySize seeks for a maximum key size of the log entry
	// and formats the timestamp and level accordingly.
	UnlimitedKeySize = -1
)

// WithMaxKeySize returns an Option that sets the maximum key size of the handler.
// By default, the maximum key size is the "timestamp + level" size, if the key
// is bigger than that it will be trimmed from the left and "..." will be added
// at the beginning.
// Minimum length of the key is 7 (length of the level with braces) and any
// value that falls out of special cases will be set to UnlimitedKeySize.
func WithMaxKeySize(s int) Option {
	return func(h *Handler) {
		if s < -1 || (s > 0 && s < 7) {
			h.maxKeySize = UnlimitedKeySize
		}

		h.maxKeySize = s
	}
}
