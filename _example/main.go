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

	// produces:
	// {
	//    "time": "2023-08-17T02:04:19.281961+06:00",
	//    "level": "INFO",
	//    "source": {
	//        "function": "main.main",
	//        "file": "/.../github.com/cappuccinotm/slogx/_example/main.go",
	//        "line": 25
	//    },
	//    "msg": "some message",
	//    "key": "value",
	//    "request_id": "bcda1960-fa4d-46b3-9c1b-fec72c7c07a3"
	// }

	logger.ErrorContext(ctx, "oh no, an error occurred",
		slog.String("details", "some important error details"),
		slogx.Error(errors.New("some error")),
	)

	// produces:
	// {
	//    "time": "2023-08-17T02:04:19.282077+06:00",
	//    "level": "ERROR",
	//    "source": {
	//        "function": "main.main",
	//        "file": "/.../github.com/cappuccinotm/slogx/_example/main.go",
	//        "line": 40
	//    },
	//    "msg": "oh no, an error occurred",
	//    "details": "some important error details",
	//    "error": "some error",
	//    "request_id": "bcda1960-fa4d-46b3-9c1b-fec72c7c07a3",
	//    "stacktrace": "main.main()\n\t/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go:40 +0x41c\n"
	//}

	logger.WithGroup("group1").InfoContext(ctx, "some message",
		slog.String("key", "value"))

	// produces:
	// {
	//    "time": "2023-08-17T02:04:19.282137+06:00",
	//    "level": "INFO",
	//    "source": {
	//        "function": "main.main",
	//        "file": "/.../github.com/cappuccinotm/slogx/_example/main.go",
	//        "line": 57
	//    },
	//    "msg": "some message",
	//    "group1": {
	//        "key": "value"
	//    },
	//    "request_id": "bcda1960-fa4d-46b3-9c1b-fec72c7c07a3"
	// }
}
