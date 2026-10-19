package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func capture(t *testing.T) *bytes.Buffer {
	t.Helper()
	t.Setenv("SERVICE_NAME", "test-svc")
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() { SetOutput(nil) })
	return &buf
}

func lines(buf *bytes.Buffer) []map[string]any {
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			panic("log line is not valid single-line JSON: " + l)
		}
		out = append(out, m)
	}
	return out
}

func TestEventIsOneJSONLineWithFixedFields(t *testing.T) {
	buf := capture(t)
	ctx := WithRequestID(context.Background(), "req-123")

	Info(ctx, "fiscal", "accepted", "invoice accepted", String("user_id", "u1"), String("mark", "400001"), Int64("duration_ms", 42))

	got := lines(buf)
	if len(got) != 1 {
		t.Fatalf("want 1 line, got %d: %s", len(got), buf.String())
	}
	e := got[0]
	for k, want := range map[string]any{
		"level": "info", "service": "test-svc", "component": "fiscal", "event": "accepted",
		"request_id": "req-123", "user_id": "u1", "mark": "400001", "msg": "invoice accepted",
	} {
		if e[k] != want {
			t.Errorf("%s = %v, want %v", k, e[k], want)
		}
	}
	if e["duration_ms"] != float64(42) {
		t.Errorf("duration_ms = %v", e["duration_ms"])
	}
	if _, ok := e["timestamp"]; !ok {
		t.Error("timestamp missing")
	}
}

func TestErrorEventCarriesTheError(t *testing.T) {
	buf := capture(t)
	Error(context.Background(), "myDATA", "request_failed", "send failed", errors.New("connection refused"))
	e := lines(buf)[0]
	if e["level"] != "error" || e["error"] != "connection refused" {
		t.Fatalf("got %v", e)
	}
}

func TestPayloadIsKeptWholeAndMultilineBodiesStayOnOneLine(t *testing.T) {
	buf := capture(t)
	body := "<InvoicesDoc>\n  <invoice>\n    <vatNumber>801340031</vatNumber>\n  </invoice>\n</InvoicesDoc>"

	Payload(context.Background(), "myDATA", "request", body, String("url", "https://x/SendInvoices"))

	got := lines(buf)
	if len(got) != 1 {
		t.Fatalf("a small payload must be one event, got %d", len(got))
	}
	if got[0]["payload"] != body {
		t.Errorf("payload altered: %q", got[0]["payload"])
	}
	if got[0]["direction"] != "request" || got[0]["event"] != "payload" || got[0]["payload_bytes"] != float64(len(body)) {
		t.Errorf("payload event fields wrong: %v", got[0])
	}
	if strings.Count(strings.TrimSpace(buf.String()), "\n") != 0 {
		t.Error("an event must be exactly one line")
	}
}

func TestHugePayloadIsSplitAndRejoinsExactly(t *testing.T) {
	buf := capture(t)
	orig := strings.Repeat("0123456789abcdef", 30000) // 480,000 bytes

	Payload(context.Background(), "bratnet", "response", orig)

	got := lines(buf)
	if len(got) < 3 {
		t.Fatalf("480 KB must be split into several events, got %d", len(got))
	}
	var rejoined strings.Builder
	for i, e := range got {
		if e["part"] != float64(i+1) || e["parts"] != float64(len(got)) {
			t.Fatalf("event %d has part=%v parts=%v", i, e["part"], e["parts"])
		}
		p := e["payload"].(string)
		if len(p) > MaxPayloadBytes {
			t.Fatalf("part %d is %d bytes, over the %d limit", i, len(p), MaxPayloadBytes)
		}
		rejoined.WriteString(p)
	}
	if rejoined.String() != orig {
		t.Fatal("payload parts do not rejoin to the original")
	}
}

func TestPayloadsAreRedactedButEverythingElseSurvives(t *testing.T) {
	buf := capture(t)
	Payload(context.Background(), "simply", "request", `{"ApiKey":"s3cr3t","InvoiceDate":"2026-10-19","vat":"801340031"}`)
	e := lines(buf)[0]
	p := e["payload"].(string)
	if strings.Contains(p, "s3cr3t") {
		t.Fatalf("secret leaked: %s", p)
	}
	if !strings.Contains(p, "801340031") || !strings.Contains(p, "2026-10-19") {
		t.Fatalf("business data was lost: %s", p)
	}
}

func TestSplitNeverCutsInsideAMultibyteCharacter(t *testing.T) {
	buf := capture(t)
	orig := strings.Repeat("Τιμολόγιο ", 40000) // Greek, multibyte
	Payload(context.Background(), "myDATA", "request", orig)
	var rejoined strings.Builder
	for _, e := range lines(buf) {
		rejoined.WriteString(e["payload"].(string))
	}
	if rejoined.String() != orig {
		t.Fatal("multibyte payload was corrupted by splitting")
	}
}

func TestRequestID(t *testing.T) {
	if RequestID(context.Background()) != "" {
		t.Error("no id expected on a bare context")
	}
	id := NewRequestID()
	if len(id) < 16 || id == NewRequestID() {
		t.Errorf("ids must be unique and long enough: %q", id)
	}
	if got := RequestID(WithRequestID(context.Background(), id)); got != id {
		t.Errorf("got %q want %q", got, id)
	}
}
