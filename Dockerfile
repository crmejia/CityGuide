# syntax=docker/dockerfile:1
FROM golang:1.21.3-alpine as build-stage
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY templates ./templates
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o /cityguide cmd/server/main.go

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /cityguide /cityguide

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/cityguide"]
