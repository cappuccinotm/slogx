package main

import (
	"github.com/cappuccinotm/slogx/slogt"
	"log/slog"
	"testing"
)

func TestSomething(t *testing.T) {
	h := slogt.Handler(t, slogt.SplitMultiline)
	logger := slog.New(h)
	logger.Debug("some single-line message",
		slog.String("key", "value"),
		slog.Group("group",
			slog.String("groupKey", "groupValue"),
		))
	logger.Info("some\nmultiline\nmessage", slog.String("key", "value"))
}
