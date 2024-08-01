FROM node:20-slim AS app
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app
COPY ./app/calibre-pages/package.json ./package.json
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install
COPY ./app/calibre-pages/ ./
RUN pnpm build

FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /calibre-api

## Deploy
FROM debian:bookworm-slim

ENV CALIBRE_TEMPLATE_DIR=/app/templates
ENV CALIBRE_STATIC_DIR=/app/static

WORKDIR /app
COPY --from=build /calibre-api ./calibre-api
COPY config.yaml ./

COPY --from=app /app/dist/ ./templates


EXPOSE 8080

ENTRYPOINT ["bash","-c","/app/calibre-api"]