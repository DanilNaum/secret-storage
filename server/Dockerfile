# Используем официальный образ Go в качестве базового образа
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модуля go и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/secret-storage-server ./cmd/main.go

# Используем Alpine Linux в качестве базового образа для финального образа
FROM alpine:latest

# Устанавливаем часовой пояс
RUN apk --no-cache add tzdata
ENV TZ=Europe/Moscow

# Создаем директорию для приложения
WORKDIR /root/

# Копируем скомпилированный бинарный файл из образа builder
COPY --from=builder /app/secret-storage-server .
COPY --from=builder /app/.env .
COPY --from=builder ./app/migrations ./migrations

# Запускаем приложение
CMD ["./secret-storage-server"]
