FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o switch ./cmd/switch
RUN go build -o angel ./cmd/angel-node
RUN go build -o backend ./cmd/api
RUN go build -o brain ./cmd/brain

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/switch /app/angel /app/backend /app/brain ./
CMD ["/app/switch"]
