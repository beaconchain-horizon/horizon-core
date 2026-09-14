FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git ca-certificates

ENV GOPROXY=direct
ENV GOSUMDB=off
ENV GIN_MODE=release

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o api ./cmd/api

EXPOSE 8080

CMD ["./api"]
