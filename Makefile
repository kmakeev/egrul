.PHONY: help setup dev build test clean docker-up docker-down docker-logs \
        parser-build parser-run parser-test \
        services-build services-run services-test \
        frontend-dev frontend-build frontend-test \
        db-migrate db-seed lint format

# Переменные
DOCKER_COMPOSE = docker compose
CARGO = cargo
GO = go
PNPM = pnpm

# Цвета
CYAN = \033[0;36m
GREEN = \033[0;32m
YELLOW = \033[1;33m
NC = \033[0m

help: ## Показать справку
	@echo "$(CYAN)ЕГРЮЛ/ЕГРИП Система - Доступные команды:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ==================== Общие команды ====================

setup: ## Начальная настройка проекта
	@echo "$(CYAN)🚀 Настройка проекта...$(NC)"
	@chmod +x infrastructure/scripts/*.sh
	@./infrastructure/scripts/setup.sh

dev: docker-up ## Запуск в режиме разработки
	@echo "$(CYAN)🔧 Запуск сервисов разработки...$(NC)"
	@$(PNPM) dev

build: parser-build services-build frontend-build ## Сборка всех компонентов
	@echo "$(GREEN)✅ Сборка завершена$(NC)"

test: parser-test services-test frontend-test ## Запуск всех тестов
	@echo "$(GREEN)✅ Тесты пройдены$(NC)"

clean: ## Очистка артефактов сборки
	@echo "$(YELLOW)🧹 Очистка...$(NC)"
	@rm -rf target/
	@rm -rf frontend/.next/
	@rm -rf frontend/node_modules/
	@rm -rf node_modules/
	@echo "$(GREEN)✅ Очистка завершена$(NC)"

lint: ## Проверка кода линтерами
	@echo "$(CYAN)🔍 Проверка кода...$(NC)"
	@cd parser && $(CARGO) clippy -- -D warnings
	@cd services/api-gateway && $(GO) vet ./...
	@cd services/search-service && $(GO) vet ./...
	@cd frontend && $(PNPM) lint

format: ## Форматирование кода
	@echo "$(CYAN)✨ Форматирование кода...$(NC)"
	@cd parser && $(CARGO) fmt
	@cd services/api-gateway && $(GO) fmt ./...
	@cd services/search-service && $(GO) fmt ./...

# ==================== Docker команды ====================

docker-up: ## Запуск Docker контейнеров
	@echo "$(CYAN)🐳 Запуск Docker контейнеров...$(NC)"
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Остановка Docker контейнеров
	@echo "$(YELLOW)🛑 Остановка Docker контейнеров...$(NC)"
	@$(DOCKER_COMPOSE) down

docker-logs: ## Просмотр логов Docker
	@$(DOCKER_COMPOSE) logs -f

docker-build: ## Сборка Docker образов
	@echo "$(CYAN)🔨 Сборка Docker образов...$(NC)"
	@$(DOCKER_COMPOSE) build

docker-clean: docker-down ## Полная очистка Docker
	@echo "$(YELLOW)🧹 Очистка Docker...$(NC)"
	@$(DOCKER_COMPOSE) down -v --rmi local

# ==================== Parser (Rust) ====================

parser-build: ## Сборка Rust парсера
	@echo "$(CYAN)🦀 Сборка парсера...$(NC)"
	@cd parser && $(CARGO) build --release

parser-run: ## Запуск парсера
	@echo "$(CYAN)▶️  Запуск парсера...$(NC)"
	@./infrastructure/scripts/parse-data.sh $(INPUT) $(OUTPUT)

parser-test: ## Тестирование парсера
	@echo "$(CYAN)🧪 Тестирование парсера...$(NC)"
	@cd parser && $(CARGO) test

parser-check: ## Проверка парсера
	@cd parser && $(CARGO) check

# ==================== Services (Go) ====================

services-build: ## Сборка Go сервисов
	@echo "$(CYAN)🐹 Сборка сервисов...$(NC)"
	@cd services/api-gateway && $(GO) build -o ../../bin/api-gateway .
	@cd services/search-service && $(GO) build -o ../../bin/search-service .

services-run-api: ## Запуск API Gateway
	@echo "$(CYAN)▶️  Запуск API Gateway...$(NC)"
	@cd services/api-gateway && $(GO) run .

services-run-search: ## Запуск Search Service
	@echo "$(CYAN)▶️  Запуск Search Service...$(NC)"
	@cd services/search-service && $(GO) run .

services-test: ## Тестирование Go сервисов
	@echo "$(CYAN)🧪 Тестирование сервисов...$(NC)"
	@cd services/api-gateway && $(GO) test ./...
	@cd services/search-service && $(GO) test ./...

# ==================== Frontend (Next.js) ====================

frontend-dev: ## Запуск frontend в режиме разработки
	@echo "$(CYAN)⚛️  Запуск frontend...$(NC)"
	@cd frontend && $(PNPM) dev

frontend-build: ## Сборка frontend
	@echo "$(CYAN)📦 Сборка frontend...$(NC)"
	@cd frontend && $(PNPM) build

frontend-test: ## Тестирование frontend
	@echo "$(CYAN)🧪 Тестирование frontend...$(NC)"
	@cd frontend && $(PNPM) test || true

frontend-install: ## Установка зависимостей frontend
	@echo "$(CYAN)📥 Установка зависимостей frontend...$(NC)"
	@$(PNPM) install

# ==================== База данных ====================

db-migrate: ## Применение миграций БД
	@echo "$(CYAN)📊 Применение миграций...$(NC)"
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d egrul -f /docker-entrypoint-initdb.d/init.sql

db-reset: ## Сброс базы данных
	@echo "$(YELLOW)⚠️  Сброс базы данных...$(NC)"
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS egrul;"
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -c "CREATE DATABASE egrul;"
	@make db-migrate

db-shell: ## Открыть psql консоль
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d egrul

# ==================== ClickHouse ====================

ch-migrate: ## Применение миграций ClickHouse
	@echo "$(CYAN)📊 Применение миграций ClickHouse...$(NC)"
	@$(DOCKER_COMPOSE) --profile setup up --force-recreate --remove-orphans clickhouse-migrations

ch-shell: ## Открыть ClickHouse консоль
	@$(DOCKER_COMPOSE) exec clickhouse clickhouse-client --user admin --password admin

ch-stats: ## Показать статистику ClickHouse
	@echo "$(CYAN)📊 Статистика ClickHouse...$(NC)"
	@./infrastructure/scripts/import-data.sh --stats

ch-truncate: ## Очистить все таблицы ClickHouse
	@echo "$(YELLOW)⚠️  Очистка всех таблиц ClickHouse...$(NC)"
	@$(DOCKER_COMPOSE) exec clickhouse clickhouse-client --user admin --password admin --multiquery -q "\
		TRUNCATE TABLE IF EXISTS egrul.companies; \
		TRUNCATE TABLE IF EXISTS egrul.entrepreneurs; \
		TRUNCATE TABLE IF EXISTS egrul.founders; \
		TRUNCATE TABLE IF EXISTS egrul.company_history; \
		TRUNCATE TABLE IF EXISTS egrul.licenses; \
		TRUNCATE TABLE IF EXISTS egrul.branches; \
		TRUNCATE TABLE IF EXISTS egrul.ownership_graph; \
		TRUNCATE TABLE IF EXISTS egrul.import_log;"
	@echo "$(GREEN)✅ Таблицы очищены$(NC)"

ch-reset: ## Полное пересоздание всех таблиц ClickHouse (удаление и применение миграций)
	@echo "$(YELLOW)⚠️  Полное пересоздание таблиц ClickHouse...$(NC)"
	@echo "$(YELLOW)⚠️  ВНИМАНИЕ: Все данные будут удалены!$(NC)"
	@$(DOCKER_COMPOSE) exec clickhouse clickhouse-client --user admin --password admin --multiquery -q "\
		DROP DATABASE IF EXISTS egrul; \
		CREATE DATABASE egrul ENGINE = Atomic;"
	@echo "$(GREEN)✅ База данных пересоздана$(NC)"
	@echo "$(CYAN)📊 Применение миграций...$(NC)"
	@make ch-migrate
	@echo "$(GREEN)✅ Таблицы пересозданы$(NC)"

# ==================== Импорт данных ====================

import: ## Импорт данных из Parquet в ClickHouse
	@echo "$(CYAN)📥 Импорт данных в ClickHouse...$(NC)"
	@./infrastructure/scripts/import-data.sh

import-docker: ## Импорт данных через Docker
	@echo "$(CYAN)🐳 Импорт данных через Docker...$(NC)"
	@$(DOCKER_COMPOSE) --profile import up data-import

import-egrul: ## Импорт только ЕГРЮЛ
	@echo "$(CYAN)📥 Импорт ЕГРЮЛ...$(NC)"
	@./infrastructure/scripts/import-data.sh --egrul ./output/egrul_egrul.parquet

import-egrip: ## Импорт только ЕГРИП
	@echo "$(CYAN)📥 Импорт ЕГРИП...$(NC)"
	@./infrastructure/scripts/import-data.sh --egrip ./output/egrip_egrip.parquet

okved-extra: ## Батч-выгрузка дополнительных ОКВЭД в отдельные таблицы
	@echo "$(CYAN)📊 Выгрузка дополнительных ОКВЭД...$(NC)"
	@chmod +x infrastructure/scripts/import-okved-extra.sh
	@./infrastructure/scripts/import-okved-extra.sh

# ==================== Полный пайплайн ====================

pipeline: ## Полный пайплайн: парсинг -> импорт
	@echo "$(CYAN)🚀 Запуск полного пайплайна...$(NC)"
	@make parser-run INPUT=$(INPUT)
	@make import
	@echo "$(GREEN)✅ Пайплайн завершен$(NC)"

# ==================== Утилиты ====================

install-deps: ## Установка всех зависимостей
	@echo "$(CYAN)📥 Установка зависимостей...$(NC)"
	@$(PNPM) install
	@cd services/api-gateway && $(GO) mod download
	@cd services/search-service && $(GO) mod download

update-deps: ## Обновление зависимостей
	@echo "$(CYAN)🔄 Обновление зависимостей...$(NC)"
	@$(PNPM) update
	@cd parser && $(CARGO) update
	@cd services/api-gateway && $(GO) get -u ./...
	@cd services/search-service && $(GO) get -u ./...

