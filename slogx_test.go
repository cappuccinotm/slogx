package slogx

import (
	"errors"
	"log/slog"
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
		t.Run("LogAttrNone", func(t *testing.T) {
			ErrAttrStrategy = LogAttrNone
			defer func() {
				ErrAttrStrategy = LogAttrAsIs
			}()

			attr := Error(nil)
			assert.Equal(t, slog.Attr{}, attr)
		})

		t.Run("LogAttrAsIs", func(t *testing.T) {
			ErrAttrStrategy = LogAttrAsIs
			defer func() {
				ErrAttrStrategy = LogAttrAsIs
			}()

			attr := Error(nil)
			assert.Equal(t, attr.Key, ErrorKey)
			assert.Nil(t, attr.Value.Any())
		})
	})
}
