# syntax=docker/dockerfile:1
FROM golang:1.19.4-alpine
#need gcc for the SQLite library https://wiki.alpinelinux.org/wiki/GCC
RUN apk add build-base
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY templates ./templates
COPY cmd ./cmd
RUN go build -o /cityguide cmd/server/main.go

EXPOSE 8080

CMD ["/cityguide"]
