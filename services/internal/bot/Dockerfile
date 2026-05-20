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

WORKDIR /build/services/internal/bot
RUN go mod download
RUN go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o /app/bot .

FROM alpine:3.23.4

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat

WORKDIR /app
COPY --from=builder /app/bot /app/bot
COPY --from=builder /build/services/internal/bot/etc /app/etc

EXPOSE 8010

CMD ["./bot", "-f", "etc/bot.yaml"]