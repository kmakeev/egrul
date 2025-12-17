# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем shared модуль
COPY services/shared ../shared

# Копируем go.mod и (если есть) go.sum
COPY services/search-service/go.mod services/search-service/go.sum* ./

# Скачиваем зависимости по go.mod
RUN go mod download

# Копируем исходный код
COPY services/search-service/ .

# Обновляем go.sum с учётом фактических импортов и собираем приложение
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /search-service .

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /search-service .

EXPOSE 8081

CMD ["./search-service"]

