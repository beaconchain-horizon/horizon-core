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

# Copy .example files to .json if they don't exist
RUN cp /root/config-seed/chain.json.example /root/config-seed/chain.json 2>/dev/null || true
RUN cp /root/config-seed/customers.json.example /root/config-seed/customers.json 2>/dev/null || true

RUN mkdir -p /root/data && \
    printf '#!/bin/sh\nif [ ! -f /root/data/chain.json ]; then cp -r /root/config-seed/. /root/data/ 2>/dev/null || true; fi\nexec /root/horizon-switch\n' > /root/start.sh && \
    chmod +x /root/start.sh

EXPOSE 8080

CMD ["/root/start.sh"]
