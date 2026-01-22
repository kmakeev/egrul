.PHONY: help setup dev build test test-coverage clean docker-up docker-down docker-logs \
        parser-build parser-run parser-test \
        services-build services-run services-test services-test-local services-test-coverage services-generate \
        frontend-dev frontend-build frontend-test \
        db-migrate db-seed lint format

# Переменные
DOCKER_COMPOSE = docker compose
CARGO = cargo
GO = go
PNPM = pnpm
HISTORY_MAX_MEMORY=10000000000 
HISTORY_BUCKETS=10

# Переменные окружения для импорта данных:
# HISTORY_MAX_MEMORY - лимит памяти для батчей истории в байтах 
# HISTORY_BUCKETS - количество батчей для обработки истории 
# 
# Примеры использования:
# make import-basic HISTORY_MAX_MEMORY=4000000000 HISTORY_BUCKETS=200


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
	@echo "$(GREEN)✅ Все тесты пройдены$(NC)"

test-coverage: ## Запуск всех тестов с покрытием кода
	@echo "$(CYAN)📊 Запуск тестов с coverage...$(NC)"
	@echo ""
	@echo "$(CYAN)=== Rust Parser ===$(NC)"
	@cd parser && cargo tarpaulin --out Stdout 2>/dev/null || cargo test
	@echo ""
	@echo "$(CYAN)=== Go Services ===$(NC)"
	@make services-test-coverage
	@echo ""
	@echo "$(CYAN)=== Frontend ===$(NC)"
	@cd frontend && $(PNPM) test:unit:coverage || true
	@echo ""
	@echo "$(GREEN)✅ Coverage тесты завершены$(NC)"

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

services-test: ## Тестирование Go сервисов (в Docker)
	@echo "$(CYAN)🧪 Тестирование сервисов...$(NC)"
	@docker run --rm \
		-v "$(PWD)/services/api-gateway:/app" \
		-w /app \
		golang:1.22 \
		sh -c "go mod tidy && go test -v -short ./..."
	@docker run --rm \
		-v "$(PWD)/services/search-service:/app" \
		-w /app \
		golang:1.22 \
		sh -c "go mod tidy && go test -v -short ./..."

services-test-local: ## Тестирование Go сервисов (локально, требуется Go)
	@echo "$(CYAN)🧪 Тестирование сервисов (локально)...$(NC)"
	@cd services/api-gateway && $(GO) test -v -short ./...
	@cd services/search-service && $(GO) test -v -short ./...

services-test-coverage: ## Тестирование с покрытием кода
	@echo "$(CYAN)📊 Тестирование с coverage...$(NC)"
	@docker run --rm \
		-v "$(PWD)/services/api-gateway:/app" \
		-w /app \
		golang:1.22 \
		sh -c "go mod tidy && go test -v -short -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total"

services-generate: ## Генерация GraphQL кода для API Gateway
	@echo "$(CYAN)🔧 Генерация GraphQL кода...$(NC)"
	@docker run --rm \
		-v "$(PWD)/services/api-gateway:/app" \
		-w /app \
		golang:1.22-alpine \
		sh -c "apk add --no-cache git && go mod tidy && go mod download && go run github.com/99designs/gqlgen generate || true"
	@if [ -f services/api-gateway/internal/graph/schema.resolvers.go ]; then \
		echo "$(YELLOW)⚠️  Удаление дублирующего schema.resolvers.go...$(NC)"; \
		rm -f services/api-gateway/internal/graph/schema.resolvers.go; \
	fi
	@echo "$(GREEN)✅ GraphQL код сгенерирован$(NC)"

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

import: ## Импорт данных из Parquet в ClickHouse с дополнительными ОКВЭД
	@echo "$(CYAN)📥 Полный импорт данных в ClickHouse...$(NC)"
	@HISTORY_MAX_MEMORY=${HISTORY_MAX_MEMORY} HISTORY_BUCKETS=${HISTORY_BUCKETS} ./infrastructure/scripts/import-data.sh
	@echo ""
	@echo "$(CYAN)📊 Выгрузка дополнительных ОКВЭД...$(NC)"
	@chmod +x infrastructure/scripts/import-okved-extra.sh
	@./infrastructure/scripts/import-okved-extra.sh
	@echo ""
	@echo "$(GREEN)✅ Полный импорт данных завершен!$(NC)"

import-basic: ## Базовый импорт данных из Parquet в ClickHouse (без дополнительных ОКВЭД)
	@echo "$(CYAN)📥 Базовый импорт данных в ClickHouse...$(NC)"
	@HISTORY_MAX_MEMORY=${HISTORY_MAX_MEMORY} HISTORY_BUCKETS=${HISTORY_BUCKETS} ./infrastructure/scripts/import-data.sh

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

pipeline: ## Полный пайплайн: парсинг -> импорт с ОКВЭД
	@echo "$(CYAN)🚀 Запуск полного пайплайна...$(NC)"
	@make parser-run INPUT=$(INPUT)
	@make import
	@echo "$(GREEN)✅ Пайплайн завершен$(NC)"

pipeline-basic: ## Базовый пайплайн: парсинг -> импорт без ОКВЭД
	@echo "$(CYAN)🚀 Запуск базового пайплайна...$(NC)"
	@make parser-run INPUT=$(INPUT)
	@make import-basic
	@echo "$(GREEN)✅ Базовый пайплайн завершен$(NC)"

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

# ==================== Docker Profiles ====================

docker-up-full: ## Запуск всех сервисов (profile: full)
	@echo "$(CYAN)🐳 Запуск всех сервисов (full profile)...$(NC)"
	@$(DOCKER_COMPOSE) --profile full up -d

docker-up-tools: ## Запуск с UI инструментами (profile: tools)
	@echo "$(CYAN)🔧 Запуск с UI инструментами...$(NC)"
	@$(DOCKER_COMPOSE) --profile tools up -d

docker-up-monitoring: ## Запуск с мониторингом (profile: monitoring)
	@echo "$(CYAN)📊 Запуск с мониторингом...$(NC)"
	@$(DOCKER_COMPOSE) --profile monitoring up -d

docker-up-dev: ## Dev mode с hot reload
	@echo "$(CYAN)🔧 Запуск в dev режиме (hot reload)...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.override.yml up -d

docker-up-prod: ## Production mode
	@echo "$(CYAN)🚀 Запуск в production режиме...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml up -d

# ==================== MinIO ====================

minio-console: ## Открыть MinIO Console
	@echo "$(CYAN)📦 Открытие MinIO Console...$(NC)"
	@open http://localhost:9001 || xdg-open http://localhost:9001 || echo "Откройте http://localhost:9001 в браузере"

minio-upload: ## Загрузить файлы в MinIO (OUTPUT=./output)
	@echo "$(CYAN)📤 Загрузка файлов в MinIO...$(NC)"
	@$(DOCKER_COMPOSE) exec minio mc cp $(OUTPUT:-./output)/* egrul/parquet-files/ --recursive

# ==================== Kafka ====================

kafka-topics: ## Список Kafka топиков
	@echo "$(CYAN)📋 Список Kafka топиков:$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-create-topic: ## Создать Kafka топик (TOPIC=name)
	@echo "$(CYAN)➕ Создание топика: $(TOPIC)$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --bootstrap-server localhost:9092 --create --topic $(TOPIC) --partitions 3 --replication-factor 1

kafka-console: ## Kafka console consumer (TOPIC=name)
	@echo "$(CYAN)🎧 Консоль Kafka для топика: $(TOPIC)$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic $(TOPIC) --from-beginning

# ==================== Elasticsearch ====================

es-create-indices: ## Создать индексы Elasticsearch с русской морфологией
	@echo "$(CYAN)📊 Создание индексов Elasticsearch...$(NC)"
	@chmod +x infrastructure/scripts/es-create-indices.sh
	@./infrastructure/scripts/es-create-indices.sh

es-delete-indices: ## Удалить индексы Elasticsearch
	@echo "$(YELLOW)⚠️  Удаление индексов Elasticsearch...$(NC)"
	@curl -X DELETE "http://localhost:9200/egrul_*"
	@echo ""

es-reindex: ## Полная переиндексация (удаление + создание + initial sync)
	@echo "$(CYAN)🔄 Полная переиндексация Elasticsearch...$(NC)"
	@chmod +x infrastructure/scripts/es-reindex.sh
	@./infrastructure/scripts/es-reindex.sh

es-sync-initial: ## Первичная синхронизация данных ClickHouse → Elasticsearch
	@echo "$(CYAN)📥 Первичная синхронизация (initial mode)...$(NC)"
	@$(DOCKER_COMPOSE) run --rm sync-service ./sync-service --mode=initial

es-sync-incremental: ## Инкрементальная синхронизация (только изменения)
	@echo "$(CYAN)🔄 Инкрементальная синхронизация...$(NC)"
	@$(DOCKER_COMPOSE) run --rm sync-service ./sync-service --mode=incremental

es-sync-daemon: ## Запуск sync-service в daemon mode (периодическая синхронизация)
	@echo "$(CYAN)🔁 Запуск sync-service в daemon mode...$(NC)"
	@$(DOCKER_COMPOSE) --profile full up -d sync-service

es-sync-stop: ## Остановка sync-service daemon
	@echo "$(YELLOW)⏹  Остановка sync-service...$(NC)"
	@$(DOCKER_COMPOSE) stop sync-service

es-stats: ## Статистика индексов Elasticsearch
	@echo "$(CYAN)📊 Статистика индексов Elasticsearch:$(NC)"
	@curl -s "http://localhost:9200/egrul_*/_stats?pretty" | grep -A5 "\"docs\"\|\"store\"\|\"indexing\"\|\"search\""
	@echo ""
	@echo "$(CYAN)📋 Список индексов:$(NC)"
	@curl -s "http://localhost:9200/_cat/indices/egrul_*?v&h=index,docs.count,store.size,health,status"

es-search-test: ## Тестовый поиск в Elasticsearch (QUERY=текст)
	@echo "$(CYAN)🔍 Тестовый поиск: $(QUERY)$(NC)"
	@curl -X POST "http://localhost:9200/egrul_companies/_search?pretty" \
		-H 'Content-Type: application/json' \
		-d '{"query": {"match": {"full_name": "$(QUERY)"}}}'

es-health: ## Проверка состояния Elasticsearch
	@echo "$(CYAN)❤️  Проверка Elasticsearch:$(NC)"
	@curl -s "http://localhost:9200/_cluster/health?pretty"

# ==================== UI Tools ====================

adminer: ## Открыть Adminer
	@echo "$(CYAN)🗄️  Открытие Adminer...$(NC)"
	@open http://localhost:8090 || xdg-open http://localhost:8090 || echo "Откройте http://localhost:8090 в браузере"

redisinsight: ## Открыть RedisInsight
	@echo "$(CYAN)🔴 Открытие RedisInsight...$(NC)"
	@open http://localhost:8091 || xdg-open http://localhost:8091 || echo "Откройте http://localhost:8091 в браузере"

# ==================== Seed Data ====================

seed-data: ## Загрузка тестовых данных из test/
	@echo "$(CYAN)🌱 Загрузка тестовых данных...$(NC)"
	@chmod +x infrastructure/scripts/seed-data.sh
	@./infrastructure/scripts/seed-data.sh

# ==================== Init Scripts ====================

init-db: ## Инициализация PostgreSQL метаданных
	@echo "$(CYAN)🔧 Инициализация PostgreSQL...$(NC)"
	@chmod +x infrastructure/scripts/init-db.sh
	@$(DOCKER_COMPOSE) exec -T postgres bash < infrastructure/scripts/init-db.sh

# ==================== ClickHouse Cluster ====================

cluster-up: ## Запуск ClickHouse кластера (6 нод + 3 Keeper)
	@echo "$(CYAN)🚀 Запуск ClickHouse кластера...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml --profile cluster up -d
	@echo "$(GREEN)✅ Кластер запущен$(NC)"
	@echo "$(CYAN)Keeper ноды: keeper-01, keeper-02, keeper-03$(NC)"
	@echo "$(CYAN)ClickHouse ноды: clickhouse-01..06$(NC)"

cluster-up-full: ## Запуск кластера с мониторингом (+ Prometheus)
	@echo "$(CYAN)🚀 Запуск кластера с мониторингом...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml --profile full up -d

cluster-down: ## Остановка ClickHouse кластера
	@echo "$(YELLOW)⏹  Остановка кластера...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml --profile cluster down

cluster-restart: ## Перезапуск кластера
	@echo "$(CYAN)🔄 Перезапуск кластера...$(NC)"
	@make cluster-down
	@make cluster-up

cluster-verify: ## Проверка состояния кластера
	@echo "$(CYAN)🔍 Проверка кластера...$(NC)"
	@chmod +x infrastructure/scripts/verify-cluster.sh
	@./infrastructure/scripts/verify-cluster.sh

cluster-test: ## Тестирование кластера (full test suite)
	@echo "$(CYAN)🧪 Тестирование кластера...$(NC)"
	@chmod +x infrastructure/scripts/test-cluster.sh
	@./infrastructure/scripts/test-cluster.sh

cluster-reset: ## Полное пересоздание БД кластера (удаление и применение миграций)
	@echo "$(YELLOW)⚠️  Полное пересоздание БД кластера...$(NC)"
	@echo "$(YELLOW)⚠️  ВНИМАНИЕ: Все данные будут удалены!$(NC)"
	@echo "$(CYAN)🛑 Остановка кластера...$(NC)"
	@docker compose -f docker-compose.cluster.yml --profile cluster down -v 2>/dev/null || true
	@echo "$(CYAN)🚀 Запуск чистого кластера...$(NC)"
	@docker compose -f docker-compose.cluster.yml --profile cluster up -d
	@echo "$(CYAN)⏳ Ожидание готовности кластера (60 сек)...$(NC)"
	@sleep 60
	@echo "$(CYAN)📊 Создание базы данных...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "\
		CREATE DATABASE IF NOT EXISTS egrul ON CLUSTER egrul_cluster ENGINE = Atomic" 2>&1 | tail -1
	@echo "$(GREEN)✅ База данных создана на всех нодах$(NC)"
	@echo "$(CYAN)📊 Применение миграции 011...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция применена, таблицы созданы$(NC)"
	@echo "$(CYAN)🔍 Проверка кластера...$(NC)"
	@make cluster-verify

cluster-truncate: ## Очистить все таблицы кластера
	@echo "$(YELLOW)⚠️  Очистка всех таблиц кластера...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery -q "\
		TRUNCATE TABLE IF EXISTS egrul.companies_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.entrepreneurs_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.founders_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.company_history_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.licenses_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.branches_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.ownership_graph_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.companies_okved_additional_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.entrepreneurs_okved_additional_local ON CLUSTER egrul_cluster; \
		TRUNCATE TABLE IF EXISTS egrul.import_log_local ON CLUSTER egrul_cluster;"
	@echo "$(GREEN)✅ Таблицы очищены на всех нодах$(NC)"

cluster-import: ## Импорт данных в кластер (использует make import)
	@echo "$(CYAN)📥 Импорт данных в кластер...$(NC)"
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 make import

cluster-import-okved: ## Импорт только дополнительных ОКВЭД в кластер
	@echo "$(CYAN)📊 Импорт дополнительных ОКВЭД в кластер...$(NC)"
	@chmod +x infrastructure/scripts/import-okved-extra.sh
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 CLICKHOUSE_USER=egrul_import CLICKHOUSE_PASSWORD=123 \
		./infrastructure/scripts/import-okved-extra.sh
	@echo "$(GREEN)✅ Импорт ОКВЭД завершен$(NC)"

cluster-frontend: ## Запуск frontend и API Gateway для работы с кластером
	@echo "$(CYAN)🌐 Запуск frontend и API Gateway для кластера...$(NC)"
	@echo "$(CYAN)API Gateway подключится к кластеру (clickhouse-01)$(NC)"
	@echo "$(CYAN)Frontend: http://localhost:3000$(NC)"
	@echo "$(CYAN)GraphQL Playground: http://localhost:8080/playground$(NC)"
	@echo ""
	@echo "$(YELLOW)⚠️  Проверьте .env файл:$(NC)"
	@echo "$(YELLOW)   NEXT_PUBLIC_GRAPHQL_URL должен быть http://localhost:8080/graphql$(NC)"
	@echo "$(YELLOW)   NEXT_PUBLIC_API_URL должен быть http://localhost:8080/api/v1$(NC)"
	@echo ""
	@$(DOCKER_COMPOSE) stop api-gateway frontend 2>/dev/null || true
	@echo "$(CYAN)🚀 Пересоздание контейнеров с новыми переменными...$(NC)"
	@CLICKHOUSE_HOST=clickhouse-01 $(DOCKER_COMPOSE) up -d --force-recreate --no-deps api-gateway frontend
	@sleep 2
	@echo "$(CYAN)🔗 Подключение к кластерной сети...$(NC)"
	@docker network connect egrul_egrul-cluster-network egrul-api-gateway 2>/dev/null || echo "  api-gateway уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-frontend 2>/dev/null || echo "  frontend уже подключен"
	@echo "$(CYAN)🔄 Перезапуск контейнеров для применения сети...$(NC)"
	@docker restart egrul-api-gateway egrul-frontend
	@echo "$(GREEN)✅ Сервисы запущены и подключены к кластеру$(NC)"
	@echo "$(CYAN)📊 Проверка статуса...$(NC)"
	@sleep 5
	@$(DOCKER_COMPOSE) ps api-gateway frontend
	@echo ""
	@echo "$(CYAN)📝 Логи API Gateway (последние 5 строк):$(NC)"
	@docker logs --tail 5 egrul-api-gateway

cluster-backup: ## Создание backup кластера в MinIO
	@echo "$(CYAN)💾 Создание backup...$(NC)"
	@chmod +x infrastructure/scripts/backup/backup-all.sh
	@./infrastructure/scripts/backup/backup-all.sh

cluster-restore: ## Восстановление из backup (BACKUP_NAME=...)
	@echo "$(CYAN)♻️  Восстановление из backup...$(NC)"
	@chmod +x infrastructure/scripts/backup/restore-all.sh
	@./infrastructure/scripts/backup/restore-all.sh $(BACKUP_NAME)

cluster-logs: ## Просмотр логов кластера
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml logs -f

cluster-ps: ## Статус нод кластера
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml ps

