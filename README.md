# slogx [![Go Reference](https://pkg.go.dev/badge/github.com/cappuccinotm/slogx.svg)](https://pkg.go.dev/github.com/cappuccinotm/slogx) [![Go](https://github.com/cappuccinotm/slogx/actions/workflows/go.yaml/badge.svg)](https://github.com/cappuccinotm/slogx/actions/workflows/go.yaml) [![codecov](https://codecov.io/gh/cappuccinotm/slogx/branch/master/graph/badge.svg?token=ueQqCRqxxS)](https://codecov.io/gh/cappuccinotm/slogx)
Package slogx contains extensions for standard library's slog package.

## Install
```bash
go get github.com/cappuccinotm/slogx
```

## Handlers
- `slogx.Accumulator(slog.Handler) slog.Handler` - returns a handler that accumulates attributes and groups from the `WithGroup` and `WithAttrs` calls, to pass them to the underlying handler only on `Handle` call. Allows middlewares to capture the handler-level attributes and groups, but may be consuming.
- `slogx.NopHandler() slog.Handler` - returns a handler that does nothing. Can be used in tests, to disable logging.
- `slog.Chain` - chains the multiple "middlewares" - handlers, which can modify the log entry.
- `slogt.TestHandler` - returns a handler that logs the log entry through `testing.T`'s `Log` function. It will shorten attributes, so the output will be more readable.
- `fblog.Handler` - a handler that logs the log entry in the [fblog-like](https://github.com/brocode/fblog) format, like:
  ```
  2024-02-05 09:11:37  [INFO]: info message
                          key: 1
       some_multi_line_string: "line1\nline2\nline3"
                multiline_any: "line1\nline2\nline3"
                    group.int: 1
                 group.string: "string"
                group.float64: 1.1
                   group.bool: false
   ...some_very_very_long_key: 1
  ```

  Some benchmarks (though this handler wasn't designed for performance, but for comfortable reading of the logs in debug/local mode):
  ```
  BenchmarkHandler
  BenchmarkHandler/fblog.NewHandler
  BenchmarkHandler/fblog.NewHandler-8         	 1479525	       800.9 ns/op	     624 B/op	       8 allocs/op
  BenchmarkHandler/slog.NewJSONHandler
  BenchmarkHandler/slog.NewJSONHandler-8      	 2407322	       500.0 ns/op	      48 B/op	       1 allocs/op
  BenchmarkHandler/slog.NewTextHandler
  BenchmarkHandler/slog.NewTextHandler-8      	 2404581	       490.0 ns/op	      48 B/op	       1 allocs/op
  ```
  
  All the benchmarks were run on a MacBook Pro (14-inch, 2021) with Apple M1 processor, the benchmark contains the only log `lg.Info("message", slog.Int("int", 1))`

## Middlewares
- `slogm.RequestID()` - adds a request ID to the context and logs it.
  - `slogm.ContextWithRequestID(ctx context.Context, requestID string) context.Context` - adds a request ID to the context.
- `slogm.StacktraceOnError()` - adds a stacktrace to the log entry if log entry's level is ERROR.
- `slogm.TrimAttrs(limit int)` - trims the length of the attributes to `limit`.
- `slogm.ApplyHandler` - adds `slog.Handler` as a `Middleware`, error from the handler will be ignored.
- `slogm.MaskSecrets(replacement string)` - masks secrets in logs, which are stored in the context
  - `slogm.AddSecrets(ctx context.Context, secret ...string) context.Context` - adds a secret value to the context
    - Note: secrets are stored in the context as a pointer to the container object, guarded by a mutex. Child context 
      can safely add secrets to the context, and the secrets will be available for the parent context, but before
      using the secrets container, the container must be initialized in the parent context with this function, e.g.:
      ```go
      ctx = slogm.AddSecrets(ctx)
      ```

## Helpers
- `slogx.Error(err error)` - adds an error to the log entry under "error" key.

## Example

```go
package main

import (
	"context"
	"errors"
	"os"

	"github.com/cappuccinotm/slogx"
	"github.com/cappuccinotm/slogx/slogm"
	"github.com/google/uuid"
	"log/slog"
)

func main() {
  h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    AddSource: true,
    Level:     slog.LevelInfo,
  })

  logger := slog.New(slogx.Accumulator(slogx.NewChain(h,
    slogm.RequestID(),
    slogm.StacktraceOnError(),
    slogm.MaskSecrets("***"),
  )))

  ctx := slogm.ContextWithRequestID(context.Background(), uuid.New().String())
  ctx = slogm.AddSecrets(ctx, "secret")
  logger.InfoContext(ctx,
    "some message",
    slog.String("key", "value"),
  )

  logger.ErrorContext(ctx, "oh no, an error occurred",
    slog.String("details", "some important secret error details"),
    slogx.Error(errors.New("some error")),
  )

  logger.WithGroup("group1").
    With(slog.String("omg", "the previous example was wrong")).
    WithGroup("group2").
    With(slog.String("omg", "this is the right example")).
    With(slog.String("key", "value")).
          InfoContext(ctx, "some message",
            slog.String("key", "value"))
}
```

Produces:
```json
{
  "time": "2023-08-17T02:04:19.281961+06:00",
  "level": "INFO",
  "source": {
    "function": "main.main",
    "file": "/.../github.com/cappuccinotm/slogx/_example/main.go",
    "line": 25
  },
  "msg": "some message",
  "key": "value",
  "request_id": "bcda1960-fa4d-46b3-9c1b-fec72c7c07a3"
}
```
``` json
{
   "time": "2023-08-17T03:35:21.251385+06:00",
   "level": "ERROR",
   "source": {
       "function": "main.main",
       "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
       "line": 47
   },
   "msg": "oh no, an error occurred",
   "details": "some important *** error details",
   "error": "some error",
   "request_id": "8ba29407-5d58-4dca-99e9-54528b1ae3f0",
   "stacktrace": "main.main()\n\t/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go:47 +0x4a4\n"
}
```
```json
{
  "time": "2024-02-18T05:02:13.030604+06:00",
  "level": "INFO",
  "source": {
    "function": "main.main",
    "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
    "line": 74
  },
  "msg": "some message",
  "key": "value",
  "group1": {
    "omg": "the previous example was wrong",
    "group2": {
      "omg": "this is the right example",
      "key": "value"
    }
  },
  "request_id": "1a34889f-a5b4-464e-9a86-0a30b50376cc"
}
```

## Client/Server logger
Package slogx also contains a `logger` package, which provides a `Logger` service, that could be used
as an HTTP server middleware and a `http.RoundTripper`, that logs HTTP requests and responses.

### Usage
```go
l := logger.New(
    logger.WithLogger(slog.Default()),
    logger.WithBody(1024),
    logger.WithUser(func(*http.Request) (string, error) { return "username", nil }),
)
```

### Options
- `logger.WithLogger(logger slog.Logger)` - sets the slog logger.
- `logger.WithLogFn(fn func(context.Context, *LogParts))` - sets a custom function to log request and response.
- `logger.WithBody(maxBodySize int)` - logs the request and response body, maximum size of the logged body is set by `maxBodySize`.
- `logger.WithUser(fn func(*http.Request) (string, error))` - sets a function to get the user data from the request.

## Testing handler
Library provides a `slogt.TestHandler` function to build a test handler, which will print out the log entries through `testing.T`'s `Log` function. It will shorten attributes, so the output will be more readable.

### Usage
```go
func TestSomething(t *testing.T) {
    h := slogt.Handler(t, slogt.SplitMultiline)
    logger := slog.New(h)
    logger.Info("some\nmultiline\nmessage", slog.String("key", "value"))
}

// Output:
// === RUN   TestSomething
//     handler.go:306: t=11:36:28.649 l=DEBUG s=main_test.go:12 msg="some single-line message" key=value group.groupKey=groupValue
//     
//     testing.go:52: some
//     testing.go:52: multiline
//     testing.go:52: message
//     handler.go:306: t=11:36:28.649 l=INFO s=main_test.go:17 msg="message with newlines has been printed to t.Log" key=value
```
