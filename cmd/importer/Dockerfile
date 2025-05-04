# -------- Builder --------
FROM golang:1.23.3-alpine AS builder
WORKDIR /go/src/app

# Копируем только модули, чтобы кешировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходник
COPY . .

# Переходим в каталог импортер-клиента и собираем
WORKDIR /go/src/app/cmd/importer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/importer .

# -------- Final --------
FROM alpine:3.17
RUN apk add --no-cache ca-certificates
WORKDIR /data

# Бинарник скомпилированный копируем из builder
COPY --from=builder /bin/importer /usr/local/bin/importer

# Монтируем папку с прайс-листами внутрь контейнера
VOLUME ["/data"]

ENTRYPOINT ["importer"]
CMD ["-dir=/data"]
