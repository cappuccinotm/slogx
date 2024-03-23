package logger

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/cappuccinotm/slogx"
)

// Option is a function that configures a Logger.
type Option func(*Logger)

// WithUser sets a custom user function.
func WithUser(fn func(*http.Request) (string, error)) Option {
	return func(l *Logger) { l.userFn = fn }
}

// WithBody sets the maximum request & response body length to be logged.
// Zero and negative values mean to not log the body at all.
func WithBody(maxBodySize int) func(l *Logger) {
	return func(l *Logger) { l.maxBodySize = maxBodySize }
}

// WithLogger is a shortcut that sets Log2Slog as the log function
// to log to slog.
func WithLogger(logger *slog.Logger) Option {
	return WithLogFn(func(ctx context.Context, parts *LogParts) {
		Log2Slog(ctx, parts, logger)
	})
}

// WithLogFn sets a custom log function.
func WithLogFn(fn func(context.Context, *LogParts)) Option {
	return func(l *Logger) { l.logFn = fn }
}

// WithSanitizeHeaders sets a custom function to sanitize headers.
func WithSanitizeHeaders(fn func(http.Header) map[string]string) Option {
	return func(l *Logger) { l.sanitizeHeadersFn = fn }
}

// WithSanitizeQuery sets a custom function to sanitize query parameters.
func WithSanitizeQuery(fn func(string) string) Option {
	return func(l *Logger) { l.sanitizeQueryFn = fn }
}

// WithMaskIP sets a custom function to mask IP addresses.
func WithMaskIP(fn func(string) string) Option {
	return func(l *Logger) { l.maskIPFn = fn }
}

// Log2Slog is the default log function that logs to slog.
func Log2Slog(ctx context.Context, parts *LogParts, logger *slog.Logger) {
	msg := "http server request"
	if parts.Client {
		msg = "http client request"
	}

	reqAttrs := []any{
		slog.String("method", parts.Request.Method),
		slog.String("url", parts.Request.URL),
		slog.Any("headers", parts.Request.Headers),
	}
	reqAttrs = appendNotEmpty(reqAttrs, "remote_ip", parts.Request.RemoteIP)
	reqAttrs = appendNotEmpty(reqAttrs, "host", parts.Request.Host)
	reqAttrs = appendNotEmpty(reqAttrs, "user", parts.Request.User)
	reqAttrs = appendNotEmpty(reqAttrs, "body", parts.Request.Body)

	respAttrs := []any{
		slog.Int("status", parts.Response.Status),
		slog.Int64("size", parts.Response.Size),
		slog.Any("headers", parts.Response.Headers),
	}
	respAttrs = appendNotEmpty(respAttrs, "body", parts.Response.Body)
	if parts.Response.Error != nil {
		respAttrs = append(respAttrs, slogx.Error(parts.Response.Error))
	}

	logger.InfoContext(ctx, msg,
		slog.Time("start_at", parts.StartAt),
		slog.Duration("duration", parts.Duration),
		slog.Group("request", reqAttrs...),
		slog.Group("response", respAttrs...),
	)
}

func appendNotEmpty(attrs []any, k, v string) []any {
	if v != "" {
		return append(attrs, slog.String(k, v))
	}
	return attrs
}

func defaultSanitizeHeaders(headers http.Header) map[string]string {
	sanitized := map[string]string{}
	for k := range headers {
		if k == "Authorization" {
			sanitized[k] = "[REDACTED]"
			continue
		}
		sanitized[k] = headers.Get(k)
	}
	return sanitized
}

var keysToHide = []string{"password", "passwd", "secret", "credentials", "token"}

func defaultSanitizeQuery(query string) string {
	u, _ := url.ParseQuery(query)

	for _, key := range keysToHide {
		if _, ok := u[key]; ok {
			u.Set(key, "[REDACTED]")
		}
	}

	return u.Encode()
}
