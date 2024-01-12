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
	"github.com/cappuccinotm/slogx/slogm"
	"github.com/google/uuid"
	"log/slog"
)

func main() {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(slogx.NewChain(h,
		slogm.RequestID(),
		slogm.StacktraceOnError(),
		slogm.MaskSecrets("***"),
	))

	ctx := slogm.ContextWithRequestID(context.Background(), uuid.New().String())
	ctx = slogm.AddSecrets(ctx, "secret")
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
    "time": "2023-08-17T02:04:19.281961+06:00",
    "level": "INFO",
    "source": {
        "function": "main.main",
        "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
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
    "time": "2023-08-17T02:04:19.282137+06:00",
    "level": "INFO",
    "source": {
        "function": "main.main",
        "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
        "line": 57
    },
    "msg": "some message",
    "group1": {
        "key": "value"
    },
    "request_id": "bcda1960-fa4d-46b3-9c1b-fec72c7c07a3"
}
```

## Middlewares
- `slogm.RequestID()` - adds a request ID to the context and logs it.
  - `slogm.ContextWithRequestID(ctx context.Context, requestID string) context.Context` - adds a request ID to the context.
- `slogm.StacktraceOnError()` - adds a stacktrace to the log entry if log entry's level is ERROR.
- `slogm.TrimAttrs(limit int)` - trims the length of the attributes to `limit`.
- `slogm.MaskSecrets(replacement string)` - masks secrets in logs, which are stored in the context
  - `slogm.AddSecrets(ctx context.Context, secret ...string) context.Context` - adds a secret value to the context
    - Note: secrets are stored in the context as a pointer to the container object, guarded by a mutex. Child context 
      can safely add secrets to the context, and the secrets will be available for the parent context, but before
      using the secrets container, the container must be initialized in the parent context with this function, e.g.:
      ```go
      ctx = slogm.AddSecrets(ctx)
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

## Status
The code is still under development. Until v1.x released the API may change.
