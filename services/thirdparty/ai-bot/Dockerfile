FROM golang:1.26.2-alpine3.23 AS builder

RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    make

WORKDIR /build

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

COPY go.work go.work.sum ./
COPY services/ ./services/

WORKDIR /build/services/thirdparty/ai-bot
RUN go mod download
RUN go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o /app/ai-bot .

FROM alpine:3.23.4

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/ai-bot /app/ai-bot
COPY services/thirdparty/ai-bot/etc /app/etc

EXPOSE 8889

CMD ["./ai-bot", "-f", "etc/ai-bot.yaml"]