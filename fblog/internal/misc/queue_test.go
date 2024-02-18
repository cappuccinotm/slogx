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

	t.Run("push-pop", func(t *testing.T) {
		q := NewQueue[int](6)
		q.PushBack(1)
		q.PushBack(2)
		q.PushBack(3)
		assert.Equal(t, 3, q.Len())
		assert.Equal(t, 1, q.PopFront())
		assert.Equal(t, 2, q.PopFront())
		assert.Equal(t, 3, q.PopFront())
	})

	t.Run("push to rotated", func(t *testing.T) {
		q := NewQueue[int](2)
		q.PushBack(1)
		q.PushBack(2)
		assert.Equal(t, 2, q.Len())
		assert.Equal(t, 1, q.PopFront())
		q.PushBack(3)
		assert.Equal(t, 2, q.PopFront())
		assert.Equal(t, 3, q.PopFront())
		q.PushBack(4)
		assert.Equal(t, 4, q.PopFront())
	})

	t.Run("pop from empty queue", func(t *testing.T) {
		q := NewQueue[int](6)
		assert.Panics(t, func() { q.PopFront() })
	})

	t.Run("min cap always preset", func(t *testing.T) {
		q := NewQueue[int](-10)
		assert.Equal(t, minCap, cap(q.l))
	})
}
