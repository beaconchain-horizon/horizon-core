FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /horizon-switch ./cmd/switch

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl su-exec && \
    addgroup -g 10001 horizon && \
    adduser -D -u 10001 -G horizon horizon
WORKDIR /app
ENV GIN_MODE=release
ENV SWITCH_PORT=8080
ENV SWITCH_DB=/data/horizon-switch.db
ENV CHAIN_CONFIG=/data/chain.json
ENV CUSTOMERS_CONFIG=/data/customers.json
COPY --from=builder /horizon-switch /app/horizon-switch
COPY --from=builder /app/config /app/config-seed
COPY scripts/start-switch.sh /app/start.sh
RUN chmod 0755 /app/start.sh && mkdir -p /data && chown -R horizon:horizon /app /data
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD curl -f http://127.0.0.1:8080/health || exit 1
CMD ["/app/start.sh"]
