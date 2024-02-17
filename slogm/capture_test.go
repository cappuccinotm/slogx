package slogm

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/cappuccinotm/slogx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapture(t *testing.T) {
	ch := make(ChannelCapturer, 2)
	buf := bytes.NewBuffer(nil)
	h := slog.Handler(slog.NewTextHandler(buf, &slog.HandlerOptions{}))
	h = slogx.NewChain(h, Capture(ch))
	lg := slog.New(h)
	lg.Info("test",
		slog.String("key", "value"),
		slog.Group("g1",
			slog.String("a", "1"),
		),
	)

	records := ch.Records()
	require.NoError(t, ch.Close())

	assert.Equal(t, []slog.Attr{
		slog.String("key", "value"),
		slog.Group("g1",
			slog.String("a", "1"),
		),
	}, slogx.Attrs(records[0]))
	require.NotEmpty(t, buf.String())
}
