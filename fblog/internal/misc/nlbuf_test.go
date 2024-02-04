package misc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewlineBuffer_String(t *testing.T) {
	buf := NewNewlineBuffer(3)
	buf.WriteString("hello\nworld\n")
	require.Equal(t, "hello\nworld\n", buf.String())
	require.Equal(t, []int{5, 11}, buf.nls)
}
