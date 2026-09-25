# gopherit.dev 🐹

The source code behind [gopherit.dev](https://gopherit.dev) — a high-performance, zero-dependency personal portfolio and API server authored entirely in vanilla Go and standard web standards.

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/Lighthouse_Performance-100%25-success?logo=lighthouse&logoColor=fff)](#)
[![Accessibility](https://img.shields.io/badge/Lighthouse_Accessibility-100%25-success?logo=lighthouse&logoColor=fff)](#)
[![Best Practices](https://img.shields.io/badge/Lighthouse_Best_Practices-100%25-success?logo=lighthouse&logoColor=fff)](#)
[![SEO](https://img.shields.io/badge/Lighthouse_SEO-100%25-success?logo=lighthouse&logoColor=fff)](#)



---

## 🏛️ Architecture & Philosophy

- **Zero External Dependencies:** Built 100% with the Go standard library (`net/http`, `html/template`, `log/slog`, `embed`).
- **Single Static Binary:** All HTML templates, stylesheets, and client scripts are embedded directly into the Go executable via `embed.FS`.
- **Pre-Compiled Template Cache:** Template trees are validated and parsed at startup into memory buffers to avoid per-request I/O and prevent partial render leaks.
- **Defense in Depth:** Hardened HTTP security headers configured on every response (Strict CSP, HSTS preload, X-Frame-Options DENY, Permissions-Policy).
- **Graceful Lifecycle Management:** Traps OS termination signals (`SIGINT`, `SIGTERM`) with a 10-second drain window for in-flight requests.
- **Structured Observability:** Native JSON logs via `log/slog` reporting request latency, status codes, and real client IPs through reverse proxies.

---

## 📂 Project Structure

```text
gopherit/
├── cmd/
│   └── server/
│       ├── main.go            # Application wiring & graceful shutdown
│       ├── handlers.go        # HTTP route handlers & template rendering
│       ├── handlers_test.go   # Table-driven handler tests
│       ├── middleware.go      # Security, logging, & panic recovery
│       └── middleware_test.go # Concurrent unit tests
├── internal/
│   └── ui/
│       ├── embed.go           # Embedded asset filesystem
│       ├── static/            # CSS, JS, and static assets
│       └── templates/         # Composed HTML templates (base, partials, pages)
├── Dockerfile                 # Multi-stage Alpine container build
├── go.mod                     # Module definition (0 dependencies)
└── .gitignore
