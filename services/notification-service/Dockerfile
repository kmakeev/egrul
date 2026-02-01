# Notification Service Dockerfile
# Multi-stage build для оптимизации размера образа

# Stage 1: Builder
FROM golang:1.22-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Рабочая директория
WORKDIR /build

# Копируем весь сервис
COPY services/notification-service/ ./

# Сборка бинарного файла (зависимости загружаются автоматически)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=mod \
    -ldflags="-w -s" \
    -o notification-service \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.20

# Метаданные
LABEL maintainer="egrul-project"
LABEL description="Notification Service for EGRUL/EGRIP change notifications"
LABEL version="1.0.0"

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata curl

# Создание пользователя для запуска сервиса
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Рабочая директория
WORKDIR /app

# Копируем бинарный файл из builder stage
COPY --from=builder /build/notification-service /app/notification-service

# Копируем email шаблоны
COPY --from=builder /build/internal/templates /app/internal/templates

# Копируем сертификаты и timezone данные
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Устанавливаем права доступа
RUN chown -R appuser:appuser /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Expose порт сервиса
EXPOSE 8083

# Health check (проверка запущенного процесса)
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD pgrep -f notification-service || exit 1

# Точка входа
ENTRYPOINT ["/app/notification-service"]
