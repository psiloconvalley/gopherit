# gopherit.dev

Source code for [gopherit.dev](https://gopherit.dev). Written in Go using only the standard library.

## Architecture

- **Zero dependencies:** Built entirely with `net/http`, `html/template`, `log/slog`, and `embed`.
- **Single binary:** HTML, CSS, and JS assets embedded via `embed.FS`.
- **Template caching:** Templates parsed and held in memory at startup.
- **Security:** Strict CSP, HSTS, X-Frame-Options, and Permissions-Policy headers.
- **Lifecycle:** Signal handling (`SIGINT`, `SIGTERM`) with connection draining.
- **Logs:** Structured JSON logging via `log/slog`.

## Project Structure

```text
gopherit/
├── cmd/
│   └── server/
│       ├── main.go            # Entrypoint, routing, shutdown
│       ├── handlers.go        # HTTP handlers
│       ├── handlers_test.go   # Handler tests
│       ├── middleware.go      # Security headers, logging, recovery
│       └── middleware_test.go # Middleware tests
├── internal/
│   └── ui/
│       ├── embed.go           # Embedded filesystem
│       ├── static/            # CSS and client JS
│       └── templates/         # HTML templates
├── Dockerfile
├── go.mod
└── .gitignore
