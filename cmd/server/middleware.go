package main

import (
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// securityHeaders attaches defense-in-depth HTTP headers to every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"style-src 'self'; "+
				"script-src 'self'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'",
		)
		next.ServeHTTP(w, r)
	})
}

// statusRecorder wraps http.ResponseWriter to intercept the HTTP status code
// for structured request logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// clientIP extracts the real client IP, accounting for reverse proxies
// like Railway, Cloudflare, or Nginx.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; the first entry is the client.
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// requestLogger logs every inbound request with method, path, status,
// latency (in milliseconds), and client IP using slog.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK, // Default status if WriteHeader is not explicitly called
			}

			next.ServeHTTP(recorder, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", clientIP(r),
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// recovery intercepts panics anywhere downstream, logs the stack trace,
// and returns a safe 500 Internal Server Error response.
func recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
						"ip", clientIP(r),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
