FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY api-linux ./api
EXPOSE 8080
ENV GIN_MODE=release
CMD ["./api"]
