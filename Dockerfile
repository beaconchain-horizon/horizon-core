FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /horizon-switch ./cmd/switch

FROM alpine:3.20
WORKDIR /root
RUN apk add --no-cache ca-certificates tzdata

ENV GIN_MODE=release
ENV SWITCH_PORT=8080
ENV SWITCH_DB=/root/data/horizon-switch.db
ENV CHAIN_CONFIG=/root/data/chain.json
ENV CUSTOMERS_CONFIG=/root/data/customers.json

COPY --from=builder /horizon-switch /root/horizon-switch
COPY --from=builder /app/config /root/config-seed

RUN mkdir -p /root/config-seed && \
    cp /root/config-seed/chain.json.example /root/config-seed/chain.json 2>/dev/null || true && \
    cp /root/config-seed/customers.json.example /root/config-seed/customers.json 2>/dev/null || true

# start.sh: هر فایل رو جدا چک می‌کنه
RUN printf '%s\n' \
    '#!/bin/sh' \
    'set -e' \
    'mkdir -p /root/data' \
    'for f in chain.json customers.json; do' \
    '  if [ ! -f "/root/data/$f" ]; then' \
    '    echo "[start.sh] copying $f from config-seed"' \
    '    cp "/root/config-seed/$f" "/root/data/$f" 2>/dev/null || true' \
    '  fi' \
    'done' \
    'echo "[start.sh] data dir contents:"' \
    'ls -la /root/data/' \
    'exec /root/horizon-switch' \
    > /root/start.sh && chmod +x /root/start.sh

EXPOSE 8080
CMD ["/root/start.sh"]
