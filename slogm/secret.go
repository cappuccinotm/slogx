package slogm

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"log/slog"
	"strings"
)

type secretsKey struct{}

// ContextWithSecrets returns a new context with the given secret.
func ContextWithSecrets(parent context.Context, secret ...string) context.Context {
	var secrets []string
	if v := parent.Value(secretsKey{}); v != nil {
		secrets = v.([]string)
	}
	secrets = append(secrets, secret...)
	return context.WithValue(parent, secretsKey{}, secrets)
}

// SecretsFromContext returns secrets from context.
func SecretsFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(secretsKey{}).([]string)
	return v, ok
}

// MaskSecrets is a middleware that masks secrets (retrieved from context) in logs.
// Works only with attributes of type String/[]byte or Any.
func MaskSecrets(replacement string) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			secrets, ok := SecretsFromContext(ctx)
			if !ok {
				return next(ctx, rec)
			}

			var nattrs []slog.Attr
			hasMaskedAttrs := false
			rec.Attrs(func(attr slog.Attr) bool {
				nattr, trimmed := mask(replacement, secrets, attr)
				nattrs = append(nattrs, nattr)
				hasMaskedAttrs = hasMaskedAttrs || trimmed
				return true
			})

			if !hasMaskedAttrs {
				return next(ctx, rec)
			}

			nrec := slog.NewRecord(rec.Time, rec.Level, rec.Message, rec.PC)
			nrec.AddAttrs(nattrs...)

			return next(ctx, nrec)
		}
	}
}

func mask(replacement string, secrets []string, attr slog.Attr) (res slog.Attr, masked bool) {
	attr.Value = attr.Value.Resolve()

	str, ok := stringValue(attr)
	if !ok {
		return attr, false
	}

	for _, secret := range secrets {
		masked = masked || strings.Contains(str, secret)
		str = strings.ReplaceAll(str, secret, replacement)
	}

	return slog.String(attr.Key, str), masked
}
