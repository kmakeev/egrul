.PHONY: all build run test clean generate lint fmt help

# Переменные
BINARY_NAME=api-gateway
GO=go
GOFLAGS=-v
BUILD_DIR=./bin
CMD_DIR=./cmd/server

# Версия (из git tag или commit)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

all: lint test build

## build: Сборка приложения
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

## run: Запуск приложения
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(CMD_DIR)/main.go

## test: Запуск тестов
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

## test-coverage: Запуск тестов с отчетом о покрытии
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## generate: Генерация GraphQL кода
generate:
	@echo "Generating GraphQL code..."
	$(GO) run github.com/99designs/gqlgen generate

## lint: Проверка кода линтером
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

## fmt: Форматирование кода
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

## clean: Очистка артефактов сборки
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## deps: Установка зависимостей
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## docker-build: Сборка Docker образа
docker-build:
	@echo "Building Docker image..."
	cd ../.. && docker build -f services/api-gateway/Dockerfile -t egrul-api-gateway:$(VERSION) .

## docker-run: Запуск в Docker
docker-run: docker-stop
	@echo "Running in Docker..."
	docker run --name egrul-api-gateway -p 8080:8080 \
		-e CLICKHOUSE_HOST=host.docker.internal \
		-e CLICKHOUSE_PORT=9000 \
		-e CLICKHOUSE_DATABASE=egrul \
		-e CLICKHOUSE_USER=egrul_api \
		-e CLICKHOUSE_PASSWORD=password123 \
		egrul-api-gateway:$(VERSION)

## docker-stop: Остановка и удаление Docker контейнера
docker-stop:
	@echo "Stopping Docker container..."
	-@docker rm -f egrul-api-gateway 2>/dev/null || true

## help: Показать справку
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

