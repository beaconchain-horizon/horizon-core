FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /horizon-switch ./cmd/switch

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl su-exec && \
    addgroup -g 10001 horizon && \
    adduser -D -u 10001 -G horizon horizon

WORKDIR /app

ENV GIN_MODE=release
ENV SWITCH_PORT=8080
ENV SWITCH_DB=/root/data/horizon-switch.db
ENV CHAIN_CONFIG=/root/data/chain.json
ENV CUSTOMERS_CONFIG=/root/data/customers.json

COPY --from=builder /horizon-switch /app/horizon-switch
COPY --from=builder /app/config /app/config-seed

RUN mkdir -p /root/data && chown -R horizon:horizon /app

RUN printf '%s\n' \
    '#!/bin/sh' \
    'set -e' \
    'mkdir -p /root/data' \
    'chown -R horizon:horizon /root/data 2>/dev/null || true' \
    'for f in chain.json customers.json; do' \
    '  if [ ! -f /root/data/$f ]; then' \
    '    cp /app/config-seed/$f /root/data/$f 2>/dev/null || true' \
    '  fi' \
    'done' \
    'echo "[start.sh] data dir:"' \
    'ls -la /root/data/' \
    'exec su-exec horizon /app/horizon-switch' \
    > /app/start.sh && chmod +x /app/start.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD curl -f http://127.0.0.1:8080/api/v1/health || exit 1

CMD ["/app/start.sh"]
