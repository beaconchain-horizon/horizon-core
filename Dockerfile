FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o switch ./cmd/switch

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/switch .
COPY --from=builder /app/config /root/config-seed
RUN mkdir -p /root/data

RUN printf '#!/bin/sh\nif [ ! -f /root/data/chain.json ]; then cp -r /root/config-seed/. /root/data/ 2>/dev/null || true; fi\n./switch\n' > /root/start.sh && chmod +x /root/start.sh

EXPOSE 8080
CMD ["/root/start.sh"]
