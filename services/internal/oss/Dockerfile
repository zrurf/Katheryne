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

WORKDIR /build/services/internal/oss
RUN go mod download
RUN go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o /app/oss .

FROM alpine:3.23.4

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/oss .
COPY services/internal/oss/etc/ ./etc/

CMD ["./oss", "-f", "etc/oss.yaml"]