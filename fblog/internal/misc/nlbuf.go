// Package misc provides miscellaneous data types and functions
// for better processing of log entries.
package misc

import (
	"fmt"
	"unsafe"
)

const (
	defaultAttrLen = 50
	minCap         = 4
)

// NewlineBuffer is a buffer that stores the positions of newlines in a byte slice.
// It also provides a method to write in the middle of the buffer, in order to align
// the tabs before log  keys.
type NewlineBuffer struct {
	buf []byte
	nls []int
}

// NewNewlineBuffer returns a new NewlineBuffer.
func NewNewlineBuffer(capacity int) *NewlineBuffer {
	// assume always that the amount of attrs is not less than minCap
	if capacity < minCap {
		capacity = minCap
	}

	return &NewlineBuffer{
		nls: make([]int, 0, capacity),
		buf: make([]byte, 0, defaultAttrLen*capacity),
	}
}

// Len returns the length of the buffer.
func (b *NewlineBuffer) Len() int { return len(b.buf) }

// Nls returns the positions of newlines in the buffer.
func (b *NewlineBuffer) Nls() []int { return b.nls }

// String returns the buffer as a string.
func (b *NewlineBuffer) String() string { return unsafe.String(unsafe.SliceData(b.buf), len(b.buf)) }

// Bytes returns the buffer as a byte slice.
func (b *NewlineBuffer) Bytes() []byte { return b.buf }

// Write writes p to the buffer.
func (b *NewlineBuffer) Write(p []byte) (n int, err error) {
	for i, c := range p {
		if c == '\n' {
			b.nls = append(b.nls, len(b.buf)+i)
		}
	}
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteAt writes p to the buffer at the given position.
func (b *NewlineBuffer) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off > int64(len(b.buf)) {
		return 0, fmt.Errorf("write at: invalid offset %d", off)
	}
	if off+int64(len(p)) > int64(len(b.buf)) {
		p = p[:len(b.buf)-int(off)]
	}
	return copy(b.buf[off:], p), nil
}

// WriteRune writes r to the buffer.
func (b *NewlineBuffer) WriteRune(r rune) (n int, err error) {
	if r == '\n' {
		b.nls = append(b.nls, len(b.buf))
	}
	b.buf = append(b.buf, byte(r))
	return 1, nil
}

// WriteString writes s to the buffer.
func (b *NewlineBuffer) WriteString(s string) (n int, err error) {
	for i, c := range s {
		if c == '\n' {
			b.nls = append(b.nls, len(b.buf)+i)
		}
	}
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// Grow grows the buffer to guarantee space for n bytes.
func (b *NewlineBuffer) Grow(n int) {
	if n < 0 {
		panic("negative count")
	}
	if cap(b.buf)-len(b.buf) < n {
		buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
		copy(buf, b.buf)
		b.buf = buf
	}
}
