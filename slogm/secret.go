package slogm

import (
	"context"
	"github.com/cappuccinotm/slogx"
	"log/slog"
	"strings"
	"sync"
)

type secretsKey struct{}

type secretsContainer struct {
	values []string
	mu     sync.RWMutex
}

func (c *secretsContainer) Add(secrets ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = append(c.values, secrets...)
}

func (c *secretsContainer) Get() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.values
}

// AddSecrets adds secrets to the context secrets container.
func AddSecrets(ctx context.Context, secret ...string) context.Context {
	v, ok := ctx.Value(secretsKey{}).(*secretsContainer)
	if !ok {
		v = &secretsContainer{}
		ctx = context.WithValue(ctx, secretsKey{}, v)
	}

	v.Add(secret...)
	return ctx
}

// SecretsFromContext returns secrets from context.
func SecretsFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(secretsKey{}).(*secretsContainer)
	if !ok {
		return nil, false
	}

	return v.Get(), true
}

// MaskSecrets is a middleware that masks secrets (retrieved from context) in logs.
//
// Works only with attributes of type String/[]byte or Any.
// If attribute is of type Any, there will be attempt to match it to:
// - encoding.TextMarshaler
// - json.Marshaler
// - []byte
// if it didn't match to any of these, it will be formatted with %+v and then masked.
//
// Child calls can add secrets to the container only if it's already present in the context,
// so before any call, user should initialize it with the first "AddSecrets" call, e.g.:
//
//	ctx = AddSecrets(ctx)
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
