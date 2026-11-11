# ─── مرحله ۱: کامپایل (Builder) ───
FROM golang:1.24-alpine AS builder
WORKDIR /app

# کپی کردن فایل‌های ماژول
COPY go.mod go.sum ./
RUN go mod download

# کپی کردن کل سورس کد
COPY . .

# کامپایل تمام سرویس‌ها برای لینوکس (بدون نیاز به CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/horizon-switch ./cmd/switch
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/horizon-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/horizon-angel ./cmd/angel-node
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/horizon-agent ./cmd/agent-node
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/brain ./cmd/brain

# ─── مرحله ۲: محیط اجرا (Runner) ───
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata curl

# کپی کردن باینری‌های کامپایل شده از مرحله قبل
COPY --from=builder /app/bin/ /app/

# تنظیم منطقه زمانی
ENV TZ=Asia/Tehran
