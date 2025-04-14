# Stage 1: Build
FROM golang:1.23.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/auth-service

# Stage 2: Runtime
FROM alpine:3.17
WORKDIR /app

COPY --from=builder /app/auth /app/auth

EXPOSE 8080
CMD ["/app/auth"]
