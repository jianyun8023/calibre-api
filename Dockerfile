FROM golang:1.25-bookworm AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /calibre-api

## Deploy
FROM debian:bookworm-slim

# 安装 CA 证书、curl（用于健康检查）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 复制 Go 后端二进制文件
COPY --from=build /calibre-api ./calibre-api
COPY config.yaml ./config.yaml

EXPOSE 8080

ENTRYPOINT ["./calibre-api"]