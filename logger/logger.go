// Package logger contains a service that provides methods to log HTTP requests
// for both server and client sides.
package logger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/tomasen/realip"
)

// Logger provides methods to log HTTP requests for both server and client sides.
type Logger struct {
	logFn             func(context.Context, *LogParts)
	userFn            func(*http.Request) (string, error)
	maskIPFn          func(string) string
	sanitizeHeadersFn func(http.Header) map[string]string
	sanitizeQueryFn   func(string) string

	maxBodySize int

	// mock functions for testing
	now func() time.Time
}

// New returns a new Logger.
func New(opts ...Option) *Logger {
	l := &Logger{
		logFn:             func(ctx context.Context, parts *LogParts) { Log2Slog(ctx, parts, slog.Default()) },
		userFn:            func(*http.Request) (string, error) { return "", nil },
		maskIPFn:          func(ip string) string { return ip },
		sanitizeHeadersFn: defaultSanitizeHeaders,
		sanitizeQueryFn:   defaultSanitizeQuery,
		maxBodySize:       0,

		now: time.Now,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// HTTPClientRoundTripper returns a RoundTripper that logs HTTP requests.
func (l *Logger) HTTPClientRoundTripper(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		reqInfo := l.obtainRequestInfo(req)
		start := l.now()

		defer func() {
			end := l.now()

			p := &LogParts{
				StartAt:  start,
				Duration: end.Sub(start),
				Request:  reqInfo,
				Response: &ResponseInfo{},
				Client:   true,
			}

			p.Response.Error = err
			if resp != nil {
				resp.Body, p.Response.Body = l.readBody(resp.Body, nil)
				p.Response.Status = resp.StatusCode
				p.Response.Size = resp.ContentLength
				p.Response.Headers = l.sanitizeHeadersFn(resp.Header)
			}

			l.logFn(req.Context(), p)
		}()

		return next.RoundTrip(req)
	})
}

// HTTPServerMiddleware returns a middleware that logs HTTP requests.
func (l *Logger) HTTPServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := &responseWriter{ResponseWriter: w, limit: l.maxBodySize}

		reqInfo := l.obtainRequestInfo(r)
		start := l.now()

		defer func() {
			end := l.now()

			p := &LogParts{
				StartAt:  start,
				Duration: end.Sub(start),
				Request:  reqInfo,
				Response: &ResponseInfo{},
			}

			p.Response.Status = wr.status
			p.Response.Size = int64(wr.size)
			p.Response.Headers = l.sanitizeHeadersFn(wr.Header())
			p.Response.Body = wr.body

			l.logFn(r.Context(), p)
		}()

		next.ServeHTTP(wr, r)
	})
}

func (l *Logger) obtainRequestInfo(req *http.Request) *RequestInfo {
	var reqBody string
	req.Body, reqBody = l.readBody(req.Body, req.GetBody)

	u := *req.URL
	u.RawQuery = l.sanitizeQueryFn(u.RawQuery)
	rawurl := u.String()
	if unescURL, err := url.QueryUnescape(rawurl); err == nil {
		rawurl = unescURL
	}

	ip := l.maskIPFn(realip.FromRequest(req))

	server := req.URL.Hostname()
	if server == "" {
		server = strings.Split(req.Host, ":")[0]
	}

	user, err := l.userFn(req)
	if err != nil {
		user = fmt.Sprintf("can't get user: %v", err)
	}

	return &RequestInfo{
		Method:   req.Method,
		URL:      rawurl,
		RemoteIP: ip,
		Host:     server,
		User:     user,
		Headers:  l.sanitizeHeadersFn(req.Header),
		Body:     reqBody,
	}
}

var reMultWhtsp = regexp.MustCompile(`[\s\p{Zs}]{2,}`)

func (l *Logger) readBody(src io.ReadCloser, getBodyFn func() (io.ReadCloser, error)) (r io.ReadCloser, bodyPart string) {
	if l.maxBodySize <= 0 {
		return src, ""
	}

	rd, body, hasMore, err := peek(src, int64(l.maxBodySize))
	if err != nil {
		return src, ""
	}

	if body != "" {
		body = strings.ReplaceAll(body, "\n", " ")
		body = reMultWhtsp.ReplaceAllString(body, " ")
	}

	if hasMore {
		body += "..."
	}

	if getBodyFn != nil {
		if rd, err := getBodyFn(); err == nil {
			return rd, body
		}
	}

	return &closerFn{Reader: rd, close: src.Close}, body
}

// LogParts contains the information to be logged.
type LogParts struct {
	// Client is true if the logger is used as round tripper.
	Client bool `json:"-"`

	Duration time.Duration `json:"duration"`
	StartAt  time.Time     `json:"start_at"`
	Request  *RequestInfo  `json:"request"`
	Response *ResponseInfo `json:"response"`
}

// RequestInfo contains the request information to be logged.
type RequestInfo struct {
	Method   string `json:"method"`
	URL      string `json:"url"`
	RemoteIP string `json:"remote_ip"`
	Host     string `json:"host"`
	User     string `json:"user"`

	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// ResponseInfo contains the response information to be logged.
type ResponseInfo struct {
	Status int   `json:"status"`
	Size   int64 `json:"size"`
	Error  error `json:"error"`

	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}
