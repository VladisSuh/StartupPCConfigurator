FROM golang:1.23.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o aggregator ./cmd/aggregator-service

FROM alpine:3.17
WORKDIR /app

COPY --from=builder /app/aggregator /app/aggregator

EXPOSE 8082
CMD ["/app/aggregator"]
