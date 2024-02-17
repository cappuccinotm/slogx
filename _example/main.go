package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/cappuccinotm/slogx/slogm"

	"github.com/cappuccinotm/slogx"
	"github.com/google/uuid"
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
		slog.String("details", "some important secret error details"),
		slogx.Error(errors.New("some error")),
	)

	// produces:
	// {
	//    "time": "2023-08-17T03:35:21.251385+06:00",
	//    "level": "ERROR",
	//    "source": {
	//        "function": "main.main",
	//        "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
	//        "line": 47
	//    },
	//    "msg": "oh no, an error occurred",
	//    "details": "some important *** error details",
	//    "error": "some error",
	//    "request_id": "8ba29407-5d58-4dca-99e9-54528b1ae3f0",
	//    "stacktrace": "main.main()\n\t/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go:47 +0x4a4\n"
	// }

	logger.WithGroup("group1").
		With(slog.String("omg", "the previous example was wrong")).
		WithGroup("group2").
		With(slog.String("omg", "this is the right example")).
		With(slog.String("key", "value")).
		InfoContext(ctx, "some message",
			slog.String("key", "value"))

	// produces:
	// {
	//     "time": "2024-02-18T05:02:13.030604+06:00",
	//     "level": "INFO",
	//     "source": {
	//         "function": "main.main",
	//         "file": "/Users/semior/go/src/github.com/cappuccinotm/slogx/_example/main.go",
	//         "line": 74
	//     },
	//     "msg": "some message",
	//     "key": "value",
	//     "group1": {
	//         "omg": "the previous example was wrong",
	//         "group2": {
	//             "omg": "this is the right example",
	//             "key": "value"
	//         }
	//     },
	//     "request_id": "1a34889f-a5b4-464e-9a86-0a30b50376cc"
	// }
}
