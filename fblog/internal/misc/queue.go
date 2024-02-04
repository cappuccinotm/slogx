package misc

// Queue is an implementation of the list data structure,
// it is a FIFO (first in, first out) data structure over a slice.
type Queue[T any] struct {
	l   []T
	idx int
	end int
}

// NewQueue returns a new Queue.
func NewQueue[T any](capacity int) *Queue[T] {
	// assume always that the amount of attrs is not less than minCap
	if capacity < minCap {
		capacity = minCap
	}

	return &Queue[T]{l: make([]T, 0, capacity)}
}

// Len returns the length of the queue.
func (q *Queue[T]) Len() int { return q.end - q.idx }

// PushBack adds an element to the end of the queue.
func (q *Queue[T]) PushBack(e T) {
	if q.end < len(q.l) {
		q.l[q.end] = e
	} else {
		q.l = append(q.l, e)
	}
	q.end++
}

// PopFront removes the first element from the queue and returns it.
func (q *Queue[T]) PopFront() T {
	if q.idx > q.end {
		panic("pop from empty queue")
	}
	e := q.l[q.idx]
	q.idx++

	// if the index is too far from the beginning of the slice
	// (half of the slice or more), then we need to copy the
	// remaining elements to the beginning of the slice and reset
	// the index, to avoid memory leaks.
	if q.idx >= cap(q.l)/2 {
		q.shift()
	}

	return e
}

// shift moves the remaining elements to the beginning of the slice
// and resets the index.
func (q *Queue[T]) shift() {
	if q.end-q.idx == 0 {
		q.idx, q.end = 0, 0
		return
	}

	for i := 0; i < q.end-q.idx; i++ {
		q.l[i] = q.l[q.idx+i]
	}
	q.end -= q.idx
	q.idx = 0
}
