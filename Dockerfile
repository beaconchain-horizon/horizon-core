FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /horizon-switch ./cmd/switch

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV GIN_MODE=release
ENV SWITCH_PORT=8080
ENV SWITCH_DB=/data/horizon-switch.db

COPY --from=builder /horizon-switch /app/horizon-switch
COPY --from=builder /app/config /app/config-seed

RUN mkdir -p /data && \
    printf '#!/bin/sh\nif [ ! -f /data/chain.json ]; then cp -r /app/config-seed/. /data/ 2>/dev/null || true; fi\nexec /app/horizon-switch\n' > /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 8080

CMD ["/app/start.sh"]
