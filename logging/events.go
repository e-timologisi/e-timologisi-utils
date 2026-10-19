package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Every event is ONE line of JSON with fixed fields, so CloudWatch Logs
// Insights can filter and aggregate on them:
//
//	timestamp, level, service, component, event, request_id, msg, + event fields
//
// Payloads (requests and responses sent to or received from AADE and the
// providers) are logged whole as a `payload` field; anything above
// MaxPayloadBytes is split into numbered parts (`part`, `parts`) that rejoin
// exactly. Credentials are masked, nothing else is dropped.

// MaxPayloadBytes keeps one event safely under CloudWatch's 256 KB limit.
const MaxPayloadBytes = 200000

type Field = zap.Field

var (
	String  = zap.String
	Int     = zap.Int
	Int64   = zap.Int64
	Bool    = zap.Bool
	Float64 = zap.Float64
	Strings = zap.Strings
)

type ctxKey struct{}

var (
	mu     sync.RWMutex
	events = build(os.Stdout)
)

func build(w io.Writer) *zap.Logger {
	enc := zap.NewProductionEncoderConfig()
	enc.TimeKey = "timestamp"
	enc.LevelKey = "level"
	enc.MessageKey = "msg"
	enc.CallerKey = ""
	enc.StacktraceKey = ""
	enc.EncodeTime = zapcore.ISO8601TimeEncoder
	enc.EncodeLevel = zapcore.LowercaseLevelEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(enc), zapcore.AddSync(w), zap.NewAtomicLevelAt(zap.InfoLevel))
	service := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if service == "" {
		service = "unknown"
	}
	return zap.New(core, zap.Fields(zap.String("service", service)))
}

// SetOutput redirects events (tests). nil restores stdout.
func SetOutput(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	mu.Lock()
	events = build(w)
	mu.Unlock()
}

func l() *zap.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return events
}

// NewRequestID returns a random 128-bit id.
func NewRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func base(ctx context.Context, component, event string, fields []Field) []Field {
	out := make([]Field, 0, len(fields)+3)
	out = append(out, zap.String("component", component), zap.String("event", event))
	if id := RequestID(ctx); id != "" {
		out = append(out, zap.String("request_id", id))
	}
	return append(out, fields...)
}

// Info logs one event.
func Info(ctx context.Context, component, event, msg string, fields ...Field) {
	l().Info(msg, base(ctx, component, event, fields)...)
}

// Error logs one error event.
func Error(ctx context.Context, component, event, msg string, err error, fields ...Field) {
	if err != nil {
		fields = append(fields, zap.String("error", err.Error()))
	}
	l().Error(msg, base(ctx, component, event, fields)...)
}

// Payload logs a request or response body whole (credentials masked).
// direction is "request", "response", "inbound" or "outbound".
func Payload(ctx context.Context, component, direction, body string, fields ...Field) {
	body = Redact(body)
	parts := split(body, MaxPayloadBytes)

	for i, p := range parts {
		f := append([]Field{
			zap.String("direction", direction),
			zap.Int("payload_bytes", len(body)),
			zap.String("payload", p),
		}, fields...)
		if len(parts) > 1 {
			f = append(f, zap.Int("part", i+1), zap.Int("parts", len(parts)))
		}
		l().Info("payload", base(ctx, component, "payload", f)...)
	}
}

// split cuts s into chunks of at most max bytes without breaking a character.
func split(s string, max int) []string {
	if len(s) <= max {
		return []string{s}
	}
	var out []string
	for len(s) > max {
		cut := max
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		out = append(out, s[:cut])
		s = s[cut:]
	}
	return append(out, s)
}
