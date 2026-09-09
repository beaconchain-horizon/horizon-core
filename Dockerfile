FROM golang:1.24-alpine
WORKDIR /app
COPY cmd/switch/switch.go .
RUN go mod init horizon-switch && \
    go mod tidy && \
    go build -o switch switch.go
EXPOSE 8080
CMD ["./switch"]
