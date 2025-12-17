# Build stage
FROM rust:1.83-alpine AS builder

RUN apk add --no-cache musl-dev

WORKDIR /app

# Копируем файлы манифестов
COPY Cargo.toml Cargo.lock ./
COPY parser/Cargo.toml ./parser/

# Создаем пустой main.rs для кэширования зависимостей
RUN mkdir -p parser/src && echo "fn main() {}" > parser/src/main.rs

# Собираем зависимости
RUN cargo build --release --package egrul-parser

# Копируем реальный исходный код
COPY parser/src ./parser/src

# Пересобираем с реальным кодом
RUN touch parser/src/main.rs && cargo build --release --package egrul-parser

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/target/release/egrul-parser .

ENTRYPOINT ["./egrul-parser"]

