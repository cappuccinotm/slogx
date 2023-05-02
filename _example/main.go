package main

import (
	"context"
	"errors"
	"os"

	"github.com/cappuccinotm/slogx"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

func main() {
	h := slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo}.
		NewJSONHandler(os.Stderr)

	logger := slog.New(slogx.NewChain(h,
		slogx.RequestID(),
		slogx.StacktraceOnError(),
	))

	ctx := slogx.ContextWithRequestID(context.Background(), uuid.New().String())
	logger.InfoCtx(ctx,
		"some message",
		slog.String("key", "value"),
	)

	// produces:
	// {
	//    "time": "2023-05-02T02:59:05.108479+03:00",
	//    "level": "INFO",
	//    "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:23",
	//    "msg": "some message",
	//    "key": "value",
	//    "request_id": "36f90947-cf6e-49be-9cf2-c59a124a6dcb"
	// }

	logger.ErrorCtx(ctx, "oh no, an error occurred",
		slog.String("details", "some important error details"),
		slogx.Error(errors.New("some error")),
	)

	// produces:
	// {
	//    "time": "2023-05-02T02:59:05.108786+03:00",
	//    "level": "ERROR",
	//    "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:30",
	//    "msg": "oh no, an error occurred",
	//    "details": "some important error details",
	//    "error": "some error",
	//    "request_id": "36f90947-cf6e-49be-9cf2-c59a124a6dcb",
	//    "stacktrace": "main.main()\n\t/.../github.com/cappuccinotm/slogx/_example/main.go:30 +0x3e4\n"
	// }

	logger.WithGroup("group1").InfoCtx(ctx, "some message",
		slog.String("key", "value"))

	// produces:
	// {
	//    "time": "2023-05-02T04:59:43.50776+03:00",
	//    "level": "INFO",
	//    "source": "/.../github.com/cappuccinotm/slogx/_example/main.go:55",
	//    "msg": "some message",
	//    "group1": {
	//        "key": "value",
	//        "request_id": "97222728-485c-44ad-8142-0ef46c70d52b"
	//    }
	// }
}
