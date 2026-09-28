FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD 2>/dev/null || echo '1.0.0')" \
    -o server \
    ./cmd/server

FROM alpine:3.20

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
