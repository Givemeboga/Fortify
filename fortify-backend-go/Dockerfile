# Fortify Go backend — multi-stage build, single static binary.
# Pure-Go SQLite (modernc.org/sqlite): no cgo/gcc needed, builds anywhere.
FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY fortify-backend-go/go.mod fortify-backend-go/go.sum ./
RUN go mod download

COPY fortify-backend-go/ ./
RUN go build -trimpath -o /fortify-go ./cmd/server

FROM debian:12-slim

WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 10001 fortify \
    && mkdir -p /app/data && chown fortify:fortify /app/data

COPY --from=builder /fortify-go ./fortify-go

USER fortify

EXPOSE 8500

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8500/healthz || exit 1

CMD ["./fortify-go"]
