// Package logging gives the HTTP clients and handlers one log style, the same
// tagged "[tag] key=value" lines e-timologisi-api uses, and makes sure
// credentials never reach the logs.
package logging

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

const maxBody = 20000

var redactions = []struct {
	re   *regexp.Regexp
	repl string
}{
	// JSON fields: "ApiKey":"...", "password": "..."
	{regexp.MustCompile(`(?i)("(?:api[_-]?key|password|secret|token|subscription[_-]?key)"\s*:\s*")[^"]*(")`), `${1}***${2}`},
	// HTTP header lines
	{regexp.MustCompile(`(?im)^((?:ocp-apim-subscription-key|x-eparoxos-api-key|apikey|authorization|x-api-key)\s*:\s*).*$`), `${1}***`},
}

// Redact masks credentials in a log string.
func Redact(s string) string {
	for _, r := range redactions {
		s = r.re.ReplaceAllString(s, r.repl)
	}
	return s
}

// Truncate shortens long bodies so one request cannot flood the log.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return fmt.Sprintf("%s...(truncated %d bytes)", s[:max], len(s)-max)
}

// Outbound logs a request about to be sent to an external API. extra is an
// optional "key=value" string (never credentials).
func Outbound(ctx context.Context, tag, method, url string, bodyBytes int, extra string) {
	f := []Field{String("method", method), String("url", url), Int("bytes", bodyBytes)}
	if extra != "" {
		f = append(f, String("detail", extra))
	}
	Info(ctx, tag, "request", "request sent", f...)
}

// Body logs a request or response body whole (credentials masked).
func Body(ctx context.Context, tag, direction, body string) {
	Payload(ctx, tag, direction, body)
}

// Inbound logs the response of an external API.
func Inbound(ctx context.Context, tag string, status int, bodyBytes int, took time.Duration) {
	Info(ctx, tag, "response", "response received", Int("status", status), Int("bytes", bodyBytes), Int64("duration_ms", took.Milliseconds()))
}

// Failed logs a request that got no response.
func Failed(ctx context.Context, tag, method, path string, took time.Duration, err error) {
	Error(ctx, tag, "request_failed", "request failed", err, String("method", method), String("path", path), Int64("duration_ms", took.Milliseconds()))
}
