package fblog

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/cappuccinotm/slogx"
	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	attrs := []interface{}{
		slog.Int("int", 1),
		slog.String("string", "string"),
		slog.Float64("float64", 1.1),
		slog.Bool("bool", true),
		slog.String("some_multi_line_string", "line1\nline2\nline3"),
		slog.Any("multiline_any", "line1\nline2\nline3"),
		slogx.Error(nil),
		slog.Group("group",
			slog.Int("int", 1),
			slog.String("string", "string"),
			slog.Float64("float64", 1.1),
			slog.Bool("bool", false),
			slog.Group("too_wide_group",
				slog.Int("some_very_very_long_key", 1),
			),
		),
	}

	t.Run("simple with source", func(t *testing.T) {
		t.Run("pos", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			lg := slog.New(NewHandler(
				WithLevel(slog.LevelDebug),
				WithSource(SourceFormatPos),
				Out(buf),
			))
			lg.Info("info message", attrs...)
			const expected = `
2006-01-02 15:04:05  [INFO]: info message
                     source: "logger_test.go:43"
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
			assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
		})

		t.Run("pos", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			lg := slog.New(NewHandler(
				WithLevel(slog.LevelDebug),
				WithSource(SourceFormatFunc),
				Out(buf),
			))
			lg.Info("info message", attrs...)
			const expected = `
2006-01-02 15:04:05  [INFO]: info message
                     source: "github.com/cappuccinotm/slogx/fblog.TestSimple.func1.2"
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
			assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
		})

		t.Run("long", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			lg := slog.New(NewHandler(
				WithLevel(slog.LevelDebug),
				WithSource(SourceFormatLong),
				Out(buf),
			))
			lg.Info("info message", attrs...)
			const expected = `
2006-01-02 15:04:05  [INFO]: info message
                     source: "{rootpath}slogx/fblog/logger_test.go:97:fblog.TestSimple.func1.3"
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
			got := buf.String()

			lines := strings.Split(got, "\n")
			pkgPathIdx := strings.Index(lines[1], "slogx/fblog")
			t.Logf("got:\n%s", got)
			t.Logf("pkgPathIdx: %d", pkgPathIdx)
			lines[1] = lines[1][:30] + "{rootpath}" + lines[1][pkgPathIdx:]
			got = strings.Join(lines, "\n")
			assert.Equal(t, expected[1:], correctTimestamps(got))
		})
	})

	t.Run("without fixed key size", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			Out(buf), Err(buf),
		))
		lg.Info("info message", attrs...)
		lg.Warn("warn message", attrs...)
		lg.Error("error message", attrs...)
		lg.Debug("debug message", attrs...)
		const expected = `
2006-01-02 15:04:05  [INFO]: info message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
2006-01-02 15:04:05  [WARN]: warn message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
2006-01-02 15:04:05 [ERROR]: error message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
2006-01-02 15:04:05 [DEBUG]: debug message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})

	t.Run("fixed key size - long", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			WithMaxKeySize(100),
			Out(buf),
		))
		lg.Info("info message", attrs...)
		const expected = `
                                                                          2006-01-02 15:04:05  [INFO]: info message
                                                                                                  int: 1
                                                                                               string: "string"
                                                                                              float64: 1.1
                                                                                                 bool: true
                                                                               some_multi_line_string: "line1\nline2\nline3"
                                                                                        multiline_any: "line1\nline2\nline3"
                                                                                                error: <nil>
                                                                                            group.int: 1
                                                                                         group.string: "string"
                                                                                        group.float64: 1.1
                                                                                           group.bool: false
                                                         group.too_wide_group.some_very_very_long_key: 1
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})

	t.Run("fixed key size - short", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			WithMaxKeySize(10),
			Out(buf),
		))
		lg.Info("info message", attrs...)
		const expected = `
...  [INFO]: info message
        int: 1
     string: "string"
    float64: 1.1
       bool: true
...e_string: "line1\nline2\nline3"
...line_any: "line1\nline2\nline3"
      error: <nil>
  group.int: 1
...p.string: "string"
....float64: 1.1
 group.bool: false
...long_key: 1
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})

	t.Run("fixed key size - minimum", func(t *testing.T) {
		t.Run("with empty ts", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			lg := slog.New(NewHandler(
				WithLevel(slog.LevelDebug),
				WithMaxKeySize(7),
				WithLogTimeFormat(""),
				Out(buf),
			))
			lg.Debug("info message", attrs...)
			const expected = `
[DEBUG]: info message
    int: 1
 string: "string"
float64: 1.1
   bool: true
...ring: "line1\nline2\nline3"
..._any: "line1\nline2\nline3"
  error: <nil>
....int: 1
...ring: "string"
...at64: 1.1
...bool: false
..._key: 1
`
			assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
		})

		t.Run("with default ts", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			lg := slog.New(NewHandler(
				WithLevel(slog.LevelDebug),
				WithMaxKeySize(0),
				Out(buf),
			))
			lg.Debug("info message", attrs...)
			const expected = `
2006-01-02 15:04:05 [DEBUG]: info message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
			assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
		})
	})

	t.Run("with child handler", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			Out(buf),
		))
		lg = lg.WithGroup("group")
		lg = lg.With(slog.Int("grouped_int", 1),
			slog.String("grouped_string", "string"))
		lg.Info("info message",
			slog.Int("ungrouped_int", 1),
			slog.String("ungrouped_string", "string"),
		)

		const expected = `
2006-01-02 15:04:05  [INFO]: info message
        group.ungrouped_int: 1
     group.ungrouped_string: "string"
          group.grouped_int: 1
       group.grouped_string: "string"
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})

	t.Run("with replacer", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			WithReplaceAttrs(func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "int" {
					return slog.Int("int", 2)
				}
				return a
			}),
			Out(buf),
		))
		lg.Info("info message", attrs...)
		const expected = `
2006-01-02 15:04:05  [INFO]: info message
                        int: 2
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 2
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})

	t.Run("unlimited key size", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		lg := slog.New(NewHandler(
			WithLevel(slog.LevelDebug),
			WithMaxKeySize(UnlimitedKeySize),
			Out(buf),
		))
		lg.Info("some very long message", attrs...)
		const expected = `
2006-01-02 15:04:05  [INFO]: some very long message
                        int: 1
                     string: "string"
                    float64: 1.1
                       bool: true
     some_multi_line_string: "line1\nline2\nline3"
              multiline_any: "line1\nline2\nline3"
                      error: <nil>
                  group.int: 1
               group.string: "string"
              group.float64: 1.1
                 group.bool: false
....some_very_very_long_key: 1
`
		assert.Equal(t, expected[1:], correctTimestamps(buf.String()))
	})
}

func correctTimestamps(s string) string {
	trimSpaces := func(s string) (result string, spacesCount int) {
		for _, c := range s {
			if c != ' ' {
				break
			}
			spacesCount++
		}
		return s[spacesCount:], spacesCount
	}

	// find all timestamps in the s
	lines := strings.Split(s, "\n")
	for i := range lines {
		line, spacesNum := trimSpaces(lines[i])
		if len(line) < 19 {
			continue
		}
		if _, err := time.Parse("2006-01-02 15:04:05", line[:19]); err != nil {
			continue
		}
		lines[i] = `2006-01-02 15:04:05` + line[19:]
		lines[i] = strings.Repeat(" ", spacesNum) + lines[i]
	}
	return strings.Join(lines, "\n")
}

func BenchmarkFBlog(b *testing.B) {
	b.ReportAllocs()
	h := NewHandler(Out(io.Discard))
	lg := slog.New(h)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info("message", slog.Int("int", 1))
	}
	b.StopTimer()
}

func BenchmarkSlogJSON(b *testing.B) {
	b.ReportAllocs()
	h := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})
	lg := slog.New(h)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info("message", slog.Int("int", 1))
	}
	b.StopTimer()
}

func BenchmarkSlogText(b *testing.B) {
	b.ReportAllocs()
	h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})
	lg := slog.New(h)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info("message", slog.Int("int", 1))
	}
	b.StopTimer()
}
