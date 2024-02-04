package misc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	t.Run("simple rotation", func(t *testing.T) {
		q := NewQueue[int](9)
		q.idx = 3
		q.end = 8
		q.l = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		assert.Equal(t, 4, q.PopFront())
		assert.Equal(t, 0, q.idx)
		assert.Equal(t, []int{5, 6, 7, 8, 5, 6, 7, 8, 9}, q.l)
		assert.Equal(t, 4, q.end)
	})

	t.Run("from the end to start", func(t *testing.T) {
		q := NewQueue[int](6)
		q.l = []int{0, 1, 2, 3, 4, 5}
		q.idx = 4
		q.end = 6
		g := q.PopFront()
		assert.Equal(t, 4, g)
		assert.Equal(t, []int{5, 1, 2, 3, 4, 5}, q.l)
		assert.Equal(t, 1, q.end)
		assert.Equal(t, 0, q.idx)
	})
}
