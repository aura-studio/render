# 基于官方golang镜像进行构建和运行
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && \
    CGO_ENABLED=0 go build -o render main.go

FROM alpine:3.18 AS runtime
WORKDIR /app
COPY --from=builder /app/render /usr/local/bin/render
ENTRYPOINT ["/usr/local/bin/render"]