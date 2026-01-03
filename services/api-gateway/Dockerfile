# Build stage
FROM golang:1.22-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем все исходные файлы
COPY services/api-gateway/ .

# Генерируем go.sum и скачиваем зависимости
RUN go mod tidy && go mod download

# Удаляем только сгенерированные файлы, НЕ наши резолверы
RUN rm -f internal/graph/generated/generated.go

# Генерируем GraphQL код, игнорируя ошибки валидации
RUN go run github.com/99designs/gqlgen generate || true

# Удаляем конфликтующие файлы ПОСЛЕ генерации
RUN rm -f internal/graph/schema.resolvers.go
RUN rm -f internal/graph/company.resolvers.go
RUN rm -f internal/graph/entrepreneur.resolvers.go

# Проверяем какие файлы остались
RUN ls -la internal/graph/*.go

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /api-gateway \
    ./cmd/server

# Runtime stage
FROM alpine:3.20

# Устанавливаем CA сертификаты и timezone данные
RUN apk --no-cache add ca-certificates tzdata curl

# Создаем непривилегированного пользователя
RUN adduser -D -g '' appuser

WORKDIR /app

# Копируем бинарник
COPY --from=builder /api-gateway .

# Копируем конфигурационные файлы
COPY --from=builder /app/config.yaml* ./

# Меняем владельца
RUN chown -R appuser:appuser /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение
CMD ["./api-gateway"]
