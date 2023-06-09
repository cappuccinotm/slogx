package slogx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func Test_TestHandler(t *testing.T) {
	tm := &testTMock{t: t}
	l := slog.New(TestHandler(tm))
	l.Debug("test", slog.String("key", "value"))

	assert.Len(t, tm.rows, 1, "should be 1 row")
	assert.Contains(t, tm.rows[0], "t=")
	assert.Contains(t, tm.rows[0], fmt.Sprintf(" l=%s", slog.LevelDebug.String()))
	assert.Contains(t, tm.rows[0], " s=testing_test.go:15")
	assert.Contains(t, tm.rows[0], fmt.Sprintf(" %s=test", slog.MessageKey))
	assert.Contains(t, tm.rows[0], " key=value")

	// show how it prints log
	l = slog.New(TestHandler(t))
	l.Debug("test", slog.String("key", "value"))
}

type testTMock struct {
	t    *testing.T
	rows []string
}

func (t *testTMock) Log(args ...any) {
	t.t.Helper()

	require.Equal(t.t, 1, len(args), "must be only 1 argument")
	row, ok := args[0].(string)
	require.True(t.t, ok, "must be string argument")
	t.rows = append(t.rows, row)
}

func (t *testTMock) Helper() {}
