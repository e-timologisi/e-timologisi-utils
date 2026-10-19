package logging

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const requestIDHeader = "X-Request-ID"

// A caller's id is kept only if it is a plain token, so a client cannot inject
// newlines or markup into the logs.
var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Flush and Unwrap keep streaming and http.ResponseController working.
func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Middleware gives every request an id (kept from X-Request-ID when the caller
// sent a safe one, so an invoice can be followed across services), returns it
// in the response header and logs one structured access event.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if !safeRequestID.MatchString(id) {
			id = NewRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		ctx := WithRequestID(r.Context(), id)

		started := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r.WithContext(ctx))
		if sw.status == 0 {
			sw.status = http.StatusOK
		}

		// The query string is never logged: it can carry personal data (AFMs).
		Info(ctx, "http", "access", "request",
			String("method", r.Method),
			String("path", r.URL.Path),
			Int("status", sw.status),
			Int("bytes", sw.bytes),
			Int64("duration_ms", time.Since(started).Milliseconds()),
			String("remote_ip", clientIP(r)),
			String("user_agent", r.UserAgent()),
		)
	})
}

// clientIP is the first X-Forwarded-For entry (behind Traefik), else the peer.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
