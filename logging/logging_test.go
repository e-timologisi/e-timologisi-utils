package logging

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name string
		in   string
		gone []string // secrets that must not survive
		kept []string // context that must survive
	}{
		{"simply api key in json", `{"ApiKey":"s3cr3t","InvoiceDate":"2026-10-19"}`, []string{"s3cr3t"}, []string{"InvoiceDate", "2026-10-19"}},
		{"lower-case api key", `{"apiKey": "s3cr3t"}`, []string{"s3cr3t"}, nil},
		{"password field", `{"username":"u","password":"hunter2"}`, []string{"hunter2"}, []string{`"username":"u"`}},
		{"subscription key header", "Ocp-Apim-Subscription-Key: abcdef123456\r\nContent-Type: text/xml", []string{"abcdef123456"}, []string{"Content-Type"}},
		{"eparoxos header", "X-EPAROXOS-API-KEY: key-999", []string{"key-999"}, nil},
		{"basic auth header", "Authorization: Basic dXNlcjpwYXNz", []string{"dXNlcjpwYXNz"}, nil},
		{"nothing sensitive", `<invoice><vatNumber>801340031</vatNumber></invoice>`, nil, []string{"801340031"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.in)
			for _, s := range tt.gone {
				if strings.Contains(got, s) {
					t.Errorf("secret %q survived: %s", s, got)
				}
			}
			for _, s := range tt.kept {
				if !strings.Contains(got, s) {
					t.Errorf("context %q was lost: %s", s, got)
				}
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("short", 10); got != "short" {
		t.Errorf("short string changed: %q", got)
	}
	got := Truncate(strings.Repeat("a", 50), 10)
	if !strings.HasPrefix(got, strings.Repeat("a", 10)) || !strings.Contains(got, "truncated 40") {
		t.Errorf("got %q", got)
	}
}
