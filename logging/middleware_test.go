package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAssignsAndEchoesRequestID(t *testing.T) {
	buf := capture(t)
	var seen string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestID(r.Context())
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("hello"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/fiscal/send-invoice?afm=123456789", nil)
	req.RemoteAddr = "10.0.0.5:5555"
	h.ServeHTTP(rec, req)

	if seen == "" || rec.Header().Get("X-Request-ID") != seen {
		t.Fatalf("request id not propagated: ctx=%q header=%q", seen, rec.Header().Get("X-Request-ID"))
	}
	e := lines(buf)[0]
	if e["component"] != "http" || e["event"] != "access" || e["request_id"] != seen {
		t.Fatalf("access event wrong: %v", e)
	}
	if e["method"] != "POST" || e["path"] != "/api/fiscal/send-invoice" || e["status"] != float64(http.StatusTeapot) || e["bytes"] != float64(5) {
		t.Errorf("access fields wrong: %v", e)
	}
	if _, leaked := e["query"]; leaked {
		t.Error("query strings may carry personal data and must not be logged")
	}
	if e["remote_ip"] != "10.0.0.5" {
		t.Errorf("remote_ip = %v", e["remote_ip"])
	}
	if _, ok := e["duration_ms"]; !ok {
		t.Error("duration_ms missing")
	}
}

func TestMiddlewareKeepsAValidIncomingRequestID(t *testing.T) {
	capture(t)
	var seen string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = RequestID(r.Context()) }))

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("X-Request-ID", "etimologisi-api-7f3a9c21")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if seen != "etimologisi-api-7f3a9c21" {
		t.Fatalf("a caller's request id must be kept so one invoice can be followed across services, got %q", seen)
	}
}

func TestMiddlewareReplacesAnUnsafeRequestID(t *testing.T) {
	capture(t)
	var seen string
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = RequestID(r.Context()) }))

	for _, bad := range []string{"has space", "new\nline", "<script>", "x"} {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		req.Header.Set("X-Request-ID", bad)
		h.ServeHTTP(httptest.NewRecorder(), req)
		if seen == bad || seen == "" {
			t.Errorf("unsafe id %q must be replaced, got %q", bad, seen)
		}
	}
}

func TestMiddlewareUsesForwardedClientIP(t *testing.T) {
	buf := capture(t)
	h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.RemoteAddr = "172.18.0.2:4444"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 172.18.0.2")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got := lines(buf)[0]["remote_ip"]; got != "203.0.113.9" {
		t.Fatalf("behind Traefik the real client ip is the first X-Forwarded-For entry, got %v", got)
	}
}

var _ = context.Background
