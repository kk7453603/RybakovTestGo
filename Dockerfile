# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Создаем пользователя для запуска приложения
RUN adduser -D -s /bin/sh -u 1001 appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod files
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download
RUN go mod verify

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Финальный образ
FROM scratch

# Импортируем пользователя и группу из builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Импортируем CA сертификаты
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Импортируем временные зоны
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем бинарный файл
COPY --from=builder /app/server /server

# Используем non-root пользователя
USER appuser

# Открываем порты
EXPOSE 8080 9090

# Запускаем приложение
ENTRYPOINT ["/server"]
