FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /calibre-api

## Deploy
FROM alpine

ENV CALIBRE_TEMPLATE_DIR=/app/templates
ENV CALIBRE_STATIC_DIR=/app/static

WORKDIR /app
COPY --from=build /calibre-api ./calibre-api
COPY config.yaml ./
COPY pages/ ./
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/calibre-api"]