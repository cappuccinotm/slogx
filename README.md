# slogx [![Go Reference](https://pkg.go.dev/badge/github.com/cappuccinotm/slogx.svg)](https://pkg.go.dev/github.com/cappuccinotm/slogx) [![Go](https://github.com/cappuccinotm/slogx/actions/workflows/go.yaml/badge.svg)](https://github.com/cappuccinotm/slogx/actions/workflows/go.yaml) [![codecov](https://codecov.io/gh/cappuccinotm/slogx/branch/master/graph/badge.svg?token=ueQqCRqxxS)](https://codecov.io/gh/cappuccinotm/slogx)
Package slogx contains extensions for standard library's slog package.

## Install
```bash
go get github.com/cappuccinotm/slogx
```

## Usage

```go
package main

import (
	"context"
	"errors"
	"os"

	"github.com/cappuccinotm/slogx"
	"github.com/google/uuid"
	"log/slog"
)

func main() {
    h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
        AddSource: true,
        Level:     slog.LevelInfo,
    })

	logger := slog.New(slogx.NewChain(h,
		slogx.RequestID(),
		slogx.StacktraceOnError(),
	))

	ctx := slogx.ContextWithRequestID(context.Background(), uuid.New().String())
	logger.InfoContext(ctx,
		"some message",
		slog.String("key", "value"),
	)
	
	logger.ErrorContext(ctx, "oh no, an error occurred",
		slog.String("details", "some important error details"),
		slogx.Error(errors.New("some error")),
	)

	logger.WithGroup("group1").InfoContext(ctx, "some message",
		slog.String("key", "value"),
	)
}
```

Produces:
```json
{
   "time": "2023-05-02T02:59:05.108479+03:00",
   "level": "INFO",
   "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:23",
   "msg": "some message",
   "key": "value",
   "request_id": "36f90947-cf6e-49be-9cf2-c59a124a6dcb"
}
```
``` json
{
  "time": "2023-05-02T02:59:05.108786+03:00",
  "level": "ERROR",
  "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:30",
  "msg": "oh no, an error occurred",
  "details": "some important error details",
  "error": "some error",
  "request_id": "36f90947-cf6e-49be-9cf2-c59a124a6dcb",
  "stacktrace": "main.main()\n\t/.../github.com/cappuccinotm/slogx/_example/main.go:30 +0x3e4\n"
}
```
```json
{
   "time": "2023-05-02T04:59:43.50776+03:00",
   "level": "INFO",
   "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:55",
   "msg": "some message",
   "group1": {
       "key": "value",
       "request_id": "97222728-485c-44ad-8142-0ef46c70d52b"
   }
}
```

## Middlewares
- `slogx.RequestID()` - adds a request ID to the context and logs it.
  - `slogx.ContextWithRequestID(ctx context.Context, requestID string) context.Context` - adds a request ID to the context.
- `slogx.StacktraceOnError()` - adds a stacktrace to the log entry if log entry's level is ERROR.

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

## Status
The code is still under development. Until v1.x released (as well as the `slog` package reaches the stdlib) the API may change.
