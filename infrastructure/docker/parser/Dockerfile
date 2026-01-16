# ==============================================================================
# Rust Parser Dockerfile - Optimized Multi-stage Build
# ==============================================================================

# Build stage
FROM rust:1.83-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache musl-dev

WORKDIR /app

# Копируем манифесты для кэширования зависимостей
COPY Cargo.toml Cargo.lock ./
COPY parser/Cargo.toml ./parser/

# Создаем пустой main.rs для pre-build зависимостей
RUN mkdir -p parser/src && echo "fn main() {}" > parser/src/main.rs

# Собираем зависимости (это закэшируется)
RUN cargo build --release --package egrul-parser

# Копируем реальный исходный код
COPY parser/src ./parser/src

# Пересобираем с реальным кодом
RUN touch parser/src/main.rs && cargo build --release --package egrul-parser

# Strip binary для уменьшения размера
RUN strip target/release/egrul-parser

# Runtime stage
FROM alpine:3.20

# Установка runtime зависимостей
RUN apk --no-cache add ca-certificates tini

WORKDIR /app

# Создаем непривилегированного пользователя
RUN addgroup -S parser && adduser -S parser -G parser

# Копируем бинарник из builder
COPY --from=builder /app/target/release/egrul-parser .

# Создаем volume mount points с правильными permissions
RUN mkdir -p /data/input /data/output && \
    chown -R parser:parser /data /app

# Переключаемся на непривилегированного пользователя
USER parser

# Tini как init process для правильной обработки сигналов
ENTRYPOINT ["/sbin/tini", "--"]

# Команда по умолчанию
CMD ["./egrul-parser", "--input", "/data/input", "--output", "/data/output"]

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./egrul-parser --version || exit 1

# Labels
LABEL maintainer="EGRUL/EGRIP System"
LABEL description="Rust XML parser for EGRUL/EGRIP government registries"
LABEL version="1.0"
