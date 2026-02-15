package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("round tripper", func(t *testing.T) {
		t.Run("body present", func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, "hello", string(b), "request body was changed")

				// manually set date to make test deterministic
				w.Header().Add("Date", "Fri, 01 Jan 2021 00:00:01 GMT")
				w.Header().Add("X-Test-Server", "test")
				w.WriteHeader(http.StatusTeapot)
				_, _ = w.Write([]byte("hi"))
			}))
			defer ts.Close()

			buf := &bytes.Buffer{}
			l := New(
				WithLogger(slog.New(slog.NewJSONHandler(buf, nil))),
				WithBody(1024),
				WithUser(func(*http.Request) (string, error) { return "username", nil }),
			)

			nowCalled := 0
			st := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			l.now = func() time.Time {
				nowCalled++
				switch nowCalled {
				case 1:
					return st
				case 2:
					return st.Add(time.Second)
				default:
					assert.Failf(t, "unexpected call to now()", "called %d times", nowCalled)
					return time.Time{}
				}
			}

			cl := ts.Client()
			cl.Transport = l.HTTPClientRoundTripper(cl.Transport)

			req, err := http.NewRequest(http.MethodGet, ts.URL+"/foo/bar", strings.NewReader("hello"))
			require.NoError(t, err)

			req.Header.Set("X-Test-Client", "test")

			resp, err := cl.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, "hi", string(b), "response body was changed")

			t.Logf("log: %s", buf.String())

			var entry jsonLogEntry
			require.NoError(t, json.NewDecoder(buf).Decode(&entry))
			assert.Equal(t, jsonLogEntry{
				Msg:   "http client request",
				Level: "INFO",
				LogParts: LogParts{
					Client:   false, // false as this field doesn't fall into logs
					Duration: time.Second,
					StartAt:  st,
					Request: &RequestInfo{
						Method:  http.MethodGet,
						URL:     ts.URL + "/foo/bar",
						Host:    "127.0.0.1",
						User:    "username",
						Headers: map[string]string{"X-Test-Client": "test"},
						Body:    "hello",
					},
					Response: &ResponseInfo{
						Status: http.StatusTeapot,
						Size:   2,
						Headers: map[string]string{
							"Date":           "Fri, 01 Jan 2021 00:00:01 GMT",
							"X-Test-Server":  "test",
							"Content-Length": "2",
							"Content-Type":   "text/plain; charset=utf-8",
						},
						Body: "hi",
					},
				},
			}, entry)
		})

		t.Run("no body present - nil", func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, "", string(b), "request body was changed")

				// manually set date to make test deterministic
				w.Header().Add("Date", "Fri, 01 Jan 2021 00:00:01 GMT")
				w.Header().Add("X-Test-Server", "test")
				w.WriteHeader(http.StatusTeapot)
				_, _ = w.Write([]byte("hi"))
			}))
			defer ts.Close()

			buf := &bytes.Buffer{}
			l := New(
				WithLogger(slog.New(slog.NewJSONHandler(buf, nil))),
				WithBody(1024),
				WithUser(func(*http.Request) (string, error) { return "username", nil }),
			)

			nowCalled := 0
			st := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			l.now = func() time.Time {
				nowCalled++
				switch nowCalled {
				case 1:
					return st
				case 2:
					return st.Add(time.Second)
				default:
					assert.Fail(t, "unexpected call to now(), called %d times", nowCalled)
					return time.Time{}
				}
			}

			cl := ts.Client()
			cl.Transport = l.HTTPClientRoundTripper(cl.Transport)

			req, err := http.NewRequest(http.MethodGet, ts.URL+"/foo/bar", nil) //nolint:gocritic // nil is intentional
			require.NoError(t, err)

			req.Header.Set("X-Test-Client", "test")

			resp, err := cl.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, "hi", string(b), "response body was changed")

			t.Logf("log: %s", buf.String())

			var entry jsonLogEntry
			require.NoError(t, json.NewDecoder(buf).Decode(&entry))
			assert.Equal(t, jsonLogEntry{
				Msg:   "http client request",
				Level: "INFO",
				LogParts: LogParts{
					Client:   false, // false as this field doesn't fall into logs
					Duration: time.Second,
					StartAt:  st,
					Request: &RequestInfo{
						Method:  http.MethodGet,
						URL:     ts.URL + "/foo/bar",
						Host:    "127.0.0.1",
						User:    "username",
						Headers: map[string]string{"X-Test-Client": "test"},
						Body:    "",
					},
					Response: &ResponseInfo{
						Status: http.StatusTeapot,
						Size:   2,
						Headers: map[string]string{
							"Date":           "Fri, 01 Jan 2021 00:00:01 GMT",
							"X-Test-Server":  "test",
							"Content-Length": "2",
							"Content-Type":   "text/plain; charset=utf-8",
						},
						Body: "hi",
					},
				},
			}, entry)
		})

		t.Run("no body present - http.NoBody", func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, "", string(b), "request body was changed")

				// manually set date to make test deterministic
				w.Header().Add("Date", "Fri, 01 Jan 2021 00:00:01 GMT")
				w.Header().Add("X-Test-Server", "test")
				w.WriteHeader(http.StatusTeapot)
				_, _ = w.Write([]byte("hi"))
			}))
			defer ts.Close()

			buf := &bytes.Buffer{}
			l := New(
				WithLogger(slog.New(slog.NewJSONHandler(buf, nil))),
				WithBody(1024),
				WithUser(func(*http.Request) (string, error) { return "username", nil }),
			)

			nowCalled := 0
			st := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
			l.now = func() time.Time {
				nowCalled++
				switch nowCalled {
				case 1:
					return st
				case 2:
					return st.Add(time.Second)
				default:
					assert.Failf(t, "unexpected call to now()", "called %d times", nowCalled)
					return time.Time{}
				}
			}

			cl := ts.Client()
			cl.Transport = l.HTTPClientRoundTripper(cl.Transport)

			req, err := http.NewRequest(http.MethodGet, ts.URL+"/foo/bar", http.NoBody)
			require.NoError(t, err)

			req.Header.Set("X-Test-Client", "test")

			resp, err := cl.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, "hi", string(b), "response body was changed")

			t.Logf("log: %s", buf.String())

			var entry jsonLogEntry
			require.NoError(t, json.NewDecoder(buf).Decode(&entry))
			assert.Equal(t, jsonLogEntry{
				Msg:   "http client request",
				Level: "INFO",
				LogParts: LogParts{
					Client:   false, // false as this field doesn't fall into logs
					Duration: time.Second,
					StartAt:  st,
					Request: &RequestInfo{
						Method:  http.MethodGet,
						URL:     ts.URL + "/foo/bar",
						Host:    "127.0.0.1",
						User:    "username",
						Headers: map[string]string{"X-Test-Client": "test"},
						Body:    "",
					},
					Response: &ResponseInfo{
						Status: http.StatusTeapot,
						Size:   2,
						Headers: map[string]string{
							"Date":           "Fri, 01 Jan 2021 00:00:01 GMT",
							"X-Test-Server":  "test",
							"Content-Length": "2",
							"Content-Type":   "text/plain; charset=utf-8",
						},
						Body: "hi",
					},
				},
			}, entry)
		})
	})

	t.Run("server middleware", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l := New(
			WithLogger(slog.New(slog.NewJSONHandler(buf, nil))),
			WithBody(1024),
			WithUser(func(*http.Request) (string, error) { return "username", nil }),
		)

		ts := httptest.NewServer(l.HTTPServerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			assert.Equal(t, "hello", string(b), "request body was changed")

			// manually set date to make test deterministic
			w.Header().Add("Date", "Fri, 01 Jan 2021 00:00:01 GMT")
			w.Header().Add("X-Test-Server", "test")
			w.WriteHeader(http.StatusTeapot)
			_, _ = w.Write([]byte("hi"))
		})))
		defer ts.Close()

		nowCalled := 0
		st := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		l.now = func() time.Time {
			nowCalled++
			switch nowCalled {
			case 1:
				return st
			case 2:
				return st.Add(time.Second)
			default:
				assert.Failf(t, "unexpected call to now()", "called %d times", nowCalled)
				return time.Time{}
			}
		}

		req, err := http.NewRequest(http.MethodGet, ts.URL+"/foo/bar", strings.NewReader("hello"))
		require.NoError(t, err)

		req.Header.Set("X-Test-Client", "test")

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, "hi", string(b), "response body was changed")

		t.Logf("log: %s", buf.String())

		var entry jsonLogEntry
		require.NoError(t, json.NewDecoder(buf).Decode(&entry))
		assert.Equal(t, jsonLogEntry{
			Msg:   "http server request",
			Level: "INFO",
			LogParts: LogParts{
				Client:   false, // false as this field doesn't fall into logs
				Duration: time.Second,
				StartAt:  st,
				Request: &RequestInfo{
					Method:   http.MethodGet,
					URL:      "/foo/bar",
					RemoteIP: "127.0.0.1",
					Host:     "127.0.0.1",
					User:     "username",
					Headers: map[string]string{
						"Accept-Encoding": "gzip",
						"Content-Length":  "5",
						"User-Agent":      "Go-http-client/1.1",
						"X-Test-Client":   "test",
					},
					Body: "hello",
				},
				Response: &ResponseInfo{
					Status: http.StatusTeapot,
					Size:   2,
					Headers: map[string]string{
						"Date":          "Fri, 01 Jan 2021 00:00:01 GMT",
						"X-Test-Server": "test",
					},
					Body: "hi",
				},
			},
		}, entry)
	})
}

type jsonLogEntry struct {
	Msg   string `json:"msg"`
	Level string `json:"level"`
	LogParts
}
