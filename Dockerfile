FROM node:20-slim AS app
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app
COPY ./app/calibre-pages/package.json ./package.json
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install
COPY ./app/calibre-pages/ ./
RUN pnpm build

FROM golang:1.24.10-trixie AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /calibre-api

## Deploy
FROM debian:bookworm-slim

# 安装 CA 证书和其他必要工具
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ENV CALIBRE_TEMPLATE_DIR=/app/templates
ENV CALIBRE_STATIC_DIR=/app/static

WORKDIR /app
COPY --from=build /calibre-api ./calibre-api
COPY config.yaml ./config.yaml

COPY --from=app /app/dist/ /app/static


EXPOSE 8080

ENTRYPOINT ["bash","-c","/app/calibre-api"]