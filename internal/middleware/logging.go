package middleware

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// wrappedResponseWriter implements http.ResponseWriter, http.Hijacker,
// http.Flusher, and http.Pusher so that middleware like gorilla/websocket
// (which needs Hijacker for the upgrade) works correctly even when the
// logging layer is active.
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	return &wrappedResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *wrappedResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return // prevent double WriteHeader
	}
	rw.wroteHeader = true
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *wrappedResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack lets gorilla/websocket take over the TCP connection for WebSocket.
// Without this method, upgrader.Upgrade() fails with HTTP 500 because it
// cannot find the http.Hijacker interface on the wrapped writer.
func (rw *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	// Should never happen with standard net/http, but guard anyway.
	panic("underlying ResponseWriter does not implement http.Hijacker")
}

// Flush implements http.Flusher — needed by SSE, streaming, and some proxies.
func (rw *wrappedResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Push implements http.Pusher — needed for HTTP/2 server push.
func (rw *wrappedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := rw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Unwrap lets Go's net/http access the original writer when needed.
func (rw *wrappedResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// ReadFrom implements io.ReaderFrom — enables sendfile(2) optimization
// for large responses that pass through an io.Reader.
func (rw *wrappedResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	if rf, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(rw.ResponseWriter, r)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newWrappedResponseWriter(w)

		next.ServeHTTP(rw, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", time.Since(start).String(),
			"ip", r.RemoteAddr,
		)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
				)
				// Only write error response if nothing was written yet.
				if rw, ok := w.(*wrappedResponseWriter); ok {
					if !rw.wroteHeader {
						http.Error(w, `{"success":false,"error":"internal server error"}`, http.StatusInternalServerError)
					}
				} else {
					http.Error(w, `{"success":false,"error":"internal server error"}`, http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
