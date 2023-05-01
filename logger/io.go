package logger

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
)

// ErrNotHijacker is returned when the underlying ResponseWriter does not
// implement http.Hijacker interface.
var ErrNotHijacker = errors.New("ResponseWriter is not a Hijacker")

type closerFn struct {
	io.Reader
	close func() error
}

func (c *closerFn) Close() error { return c.close() }

func peek(src io.Reader, limit int64) (rd io.Reader, s string, full bool, err error) {
	if limit < 0 {
		limit = 0
	}

	buf := &bytes.Buffer{}
	if _, err = io.CopyN(buf, src, limit+1); err == io.EOF {
		str := buf.String()
		return buf, str, false, nil
	}
	if err != nil {
		return src, "", false, err
	}

	s = buf.String()
	s = s[:len(s)-1]

	return io.MultiReader(buf, src), s, true, nil
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   string

	limit int
}

// WriteHeader implements http.ResponseWriter and saves status
func (c *responseWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

// Write implements http.ResponseWriter and tracks number of bytes written
func (c *responseWriter) Write(b []byte) (int, error) {
	if c.status == 0 {
		c.status = 200
	}

	if c.limit > 0 {
		part := b
		if len(b) > c.limit {
			part = b[:c.limit]
		}
		c.body += string(part)
		c.limit -= len(part)
	}

	n, err := c.ResponseWriter.Write(b)
	c.size += n
	return n, err
}

// Flush implements http.Flusher
func (c *responseWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker
func (c *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := c.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, ErrNotHijacker
}
