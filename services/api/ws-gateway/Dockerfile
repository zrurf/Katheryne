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

WORKDIR /build/services/api/ws-gateway
RUN go mod download
RUN go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o /app/ws-gateway .

FROM alpine:3.23.4

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat

WORKDIR /app
COPY --from=builder /app/ws-gateway .
COPY --from=builder /build/services/api/ws-gateway/etc/ ./etc/

EXPOSE 8080

CMD ["./ws-gateway"]