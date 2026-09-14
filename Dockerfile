FROM golang:1.24-alpine

WORKDIR /app

ENV GIN_MODE=release

COPY . .

RUN go build -mod=vendor -o api ./cmd/api

EXPOSE 8080

CMD ["./api"]
