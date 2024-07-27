FROM golang:1.19-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /calibre-api

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /app
COPY --from=build /calibre-api ./calibre-api
COPY config.yaml ./
COPY pages/ ./
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/calibre-api"]