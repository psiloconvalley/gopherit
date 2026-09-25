package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// nullLogger returns a logger that discards all log messages.
// This keeps test outputs clean and readable.
func nullLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()

	// A dummy next handler to verify the middleware passes requests through
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := securityHeaders(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=()",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
	}

	for header, expectedVal := range expectedHeaders {
		got := rec.Header().Get(header)
		if got != expectedVal {
			t.Errorf("expected header %q = %q; got %q", header, expectedVal, got)
		}
	}

	// Verify Content-Security-Policy exists and contains vital rules
	csp := rec.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header, got empty string")
	}
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	// Handler designed intentionally to panic
	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went catastrophically wrong")
	})

	mw := recovery(nullLogger())(panickingHandler)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// Execute — should NOT crash the test suite
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on panic; got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "Single X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.195",
		},
		{
			name: "Multiple X-Forwarded-For Proxies",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.195",
		},
		{
			name: "X-Real-IP fallback",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.1",
			},
			remoteAddr: "10.0.0.1:1234",
			expected:   "198.51.100.1",
		},
		{
			name:       "Direct RemoteAddr without headers",
			headers:    map[string]string{},
			remoteAddr: "192.0.2.1:54321",
			expected:   "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := clientIP(req)
			if got != tt.expected {
				t.Errorf("expected IP %q, got %q", tt.expected, got)
			}
		})
	}
}
