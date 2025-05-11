# Стадия сборки
FROM  golang:1.24-alpine  AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod ./
COPY go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

# RUN apk add --no-cache build-base librdkafka-dev 

RUN CGO_ENABLED=1 go build -tags musl -o /kafka-reader kafka-reader/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированные бинарники
COPY --from=builder /kafka-reader .

# Запускаем kafka-reader
CMD ["./kafka-reader"]