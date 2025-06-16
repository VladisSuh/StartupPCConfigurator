FROM golang:1.23.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o config ./cmd/config-service

FROM alpine:3.17
WORKDIR /app

COPY --from=builder /app/config /app/config

EXPOSE 8002
EXPOSE 6060
CMD ["/app/config"]
