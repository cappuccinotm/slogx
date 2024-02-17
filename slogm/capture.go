package slogm

import (
	"context"
	"log/slog"

	"github.com/cappuccinotm/slogx"
)

// Capturer is an interface for capturing log records.
type Capturer interface {
	Push(slog.Record)
	Records() []slog.Record
	Close() error
}

// ChannelCapturer is a channel that captures log records.
type ChannelCapturer chan slog.Record

// Push adds a record to the channel.
func (c ChannelCapturer) Push(rec slog.Record) { c <- rec }

// Close closes the channel.
func (c ChannelCapturer) Close() error {
	close(c)
	return nil
}

// Records returns all captured records for this moment.
func (c ChannelCapturer) Records() []slog.Record {
	var records []slog.Record
	for len(c) > 0 {
		records = append(records, <-c)
	}
	return records
}

// Capture returns a middleware that captures log records.
func Capture(capt Capturer) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			capt.Push(rec)
			return next(ctx, rec)
		}
	}
}
