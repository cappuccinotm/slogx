package slogx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("actual error", func(t *testing.T) {
		err := errors.New("test")
		attr := Error(err)
		assert.Equal(t, attr.Key, ErrorKey)
		assert.Equal(t, attr.Value.String(), err.Error())
	})

	t.Run("nil error", func(t *testing.T) {
		attr := Error(nil)
		assert.Equal(t, attr.Key, ErrorKey)
		assert.Equal(t, attr.Value.String(), "nil")
	})
}
