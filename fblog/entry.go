package fblog

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cappuccinotm/slogx/fblog/internal/misc"
)

type entry struct {
	timeFmt string
	rep     func([]string, slog.Attr) slog.Attr

	// internal use
	headerLen int
	buf       *bytes.Buffer
	q         *misc.Queue[grouped]
}

const lvlSize = 7 // length of the level with braces, e.g. "[DEBUG]"

func newEntry(timeFmt string, rep func([]string, slog.Attr) slog.Attr, numAttrs int) *entry {
	return &entry{
		buf:     bytes.NewBuffer(nil),
		q:       misc.NewQueue[grouped](numAttrs),
		timeFmt: timeFmt,
		rep:     rep,
	}
}

func (e *entry) WriteHeader(logTimeFmt string, maxKeySize int, rec slog.Record) {
	if logTimeFmt != "" {
		tmFmts := rec.Time.Format(logTimeFmt)
		maxKeySize -= lvlSize
		switch {
		case maxKeySize == UnlimitedKeySize-lvlSize:
			e.buf.WriteString(tmFmts) // TODO handle the case with expanding space prefixes for a long
		case maxKeySize == HeaderKeySize-lvlSize:
			e.buf.WriteString(tmFmts)
		case len(tmFmts) > maxKeySize:
			// trim the timestamp from the left, add "..." at the beginning
			e.buf.WriteString("...")
			e.buf.WriteString(tmFmts[len(tmFmts)-maxKeySize+3:])
		case maxKeySize > len(tmFmts):
			e.spaces(maxKeySize-len(tmFmts), false)
			e.buf.WriteString(tmFmts)
		default:
			e.buf.WriteString(tmFmts)
		}
		e.buf.WriteString(" ")
	}

	switch rec.Level {
	case slog.LevelDebug:
		e.buf.WriteString("[DEBUG]")
	case slog.LevelInfo:
		e.buf.WriteString(" [INFO]")
	case slog.LevelWarn:
		e.buf.WriteString(" [WARN]")
	case slog.LevelError:
		e.buf.WriteString("[ERROR]")
	default:
		e.buf.WriteString("[UNKNW]")
	}

	e.headerLen = e.buf.Len()
	e.buf.WriteString(": ")
	e.buf.WriteString(rec.Message)
	e.buf.WriteString("\n")
}

type grouped struct {
	group []string
	attr  slog.Attr
}

func (e *entry) WriteAttr(group []string, a slog.Attr) {
	e.q.PushBack(grouped{group: group, attr: a})

	for e.q.Len() > 0 {
		g := e.q.PopFront()

		groups, attr := g.group, g.attr
		attr = e.rep(groups, attr)
		attr.Value = attr.Value.Resolve() // resolve the value before writing

		if attr.Value.Kind() == slog.KindGroup {
			for _, a := range attr.Value.Group() {
				e.q.PushBack(grouped{group: append(groups, attr.Key), attr: a})
			}
			continue
		}

		e.WriteKey(groups, attr)
		e.buf.WriteString(": ")
		e.WriteTextValue(attr)
		e.buf.WriteString("\n")
	}
}

func (e *entry) WriteTextValue(attr slog.Attr) {
	switch attr.Value.Kind() {
	case slog.KindString:
		_, _ = fmt.Fprintf(e.buf, "%q", attr.Value.String()) // escape the string
	case slog.KindTime:
		e.buf.WriteString(attr.Value.Time().Format(e.timeFmt))
	case slog.KindDuration:
		e.buf.WriteString(attr.Value.Duration().String())
	case slog.KindGroup:
		panic("impossible case, group should be resolved to this point, please, file an issue")
	default:
		_, _ = fmt.Fprintf(e.buf, "%+v", attr.Value.Any())
	}
}

func (e *entry) WriteKey(groups []string, attr slog.Attr) {
	key := &strings.Builder{}
	for _, g := range groups {
		key.WriteString(g)
		key.WriteString(".")
	}
	key.WriteString(attr.Key)
	s := key.String()

	e.buf.Grow(e.headerLen) // preallocate the space for the key
	if e.headerLen-key.Len() < 0 {
		// trim the key from the left, add "..." at the beginning
		e.buf.WriteString("...")
		s = s[key.Len()-e.headerLen+3:]
	}

	if e.headerLen-key.Len() > 0 {
		e.spaces(e.headerLen-key.Len(), true)
	}

	e.buf.WriteString(s)
}

func (e *entry) WriteTo(wr io.Writer) (int64, error) {
	// TODO: append offset in case of UnlimitedKeySize
	n, err := wr.Write(e.buf.Bytes())
	return int64(n), err
}

func (e *entry) spaces(n int, alreadyGrown bool) {
	if !alreadyGrown {
		e.buf.Grow(n)
	}
	for i := 0; i < n; i++ {
		e.buf.WriteByte(' ')
	}
}
