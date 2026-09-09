FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o switch ./cmd/switch
EXPOSE 8080
CMD ["./switch"]
