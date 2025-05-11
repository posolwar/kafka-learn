FROM golang:1.24.3 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app


COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Копируем весь исходный код проекта
COPY . .

# Компилируем kafka-writer
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka-writer kafka-writer/main.go

# Компилируем kafka-reader
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka-reader kafka-reader/main.go

# Создаем финальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем скомпилированные бинарные файлы из builder
COPY --from=builder /kafka-writer /app/kafka-writer
COPY --from=builder /kafka-reader /app/kafka-reader

# Устанавливаем зависимости для работы бинарных файлов
RUN apk --no-cache add ca-certificates

# Указываем точку входа 
CMD ["/bin/sh"]