# Change Detection Service Dockerfile
# Multi-stage build для оптимизации размера образа

# Stage 1: Builder
FROM golang:1.23-alpine AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git ca-certificates tzdata

# Рабочая директория
WORKDIR /build

# Копируем shared модуль
COPY services/shared/ /shared/

# Копируем весь сервис
COPY services/change-detection-service/ ./

# Сборка бинарного файла (зависимости загружаются автоматически)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=mod \
    -ldflags="-w -s" \
    -o change-detection-service \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.20

# Метаданные
LABEL maintainer="egrul-project"
LABEL description="Change Detection Service for EGRUL/EGRIP data"
LABEL version="1.0.0"

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata curl

# Создание пользователя для запуска сервиса
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Рабочая директория
WORKDIR /app

# Копируем бинарный файл из builder stage
COPY --from=builder /build/change-detection-service /app/change-detection-service

# Копируем сертификаты и timezone данные
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Устанавливаем права доступа
RUN chown -R appuser:appuser /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Expose порт сервиса
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

# Точка входа
ENTRYPOINT ["/app/change-detection-service"]
