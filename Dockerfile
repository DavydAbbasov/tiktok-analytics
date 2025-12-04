FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/tiktok ./cmd/tiktok
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/migrator ./cmd/migrator
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bin ./bin
COPY internal/config/config.yaml ./internal/config/config.yaml
COPY migrations ./migrations


