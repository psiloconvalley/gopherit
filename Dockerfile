# syntax=docker/dockerfile:1.4

# ── Stage 1: Build Environment ────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for version injection and ca-certificates for TLS
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod ./
RUN go mod download

# Copy source + embedded assets
COPY . .

# Fast Compile with BuildKit Cache Mounts
RUN --mount=type=cache,id=go-build,target=/root/.cache/go-build \
    --mount=type=cache,id=go-mod,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD 2>/dev/null || echo 'production')" \
    -o server \
    ./cmd/server

# ── Stage 2: Hardened Runtime ─────────────────────────
FROM alpine:3.20

# Copy SSL root certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run as unprivileged user (Principle of Least Privilege)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

# Copy ONLY the compiled binary
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
