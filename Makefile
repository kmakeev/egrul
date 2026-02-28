.PHONY: help setup dev build test test-coverage clean \
        up down \
        docker-up docker-up-full docker-up-tools docker-up-monitoring docker-up-dev docker-up-prod \
        docker-down docker-logs docker-build docker-clean \
        docker-clean-networks docker-full-clean \
        parser-build parser-run parser-test parser-check \
        services-build services-run-api services-run-search \
        services-test services-test-local services-test-coverage services-generate \
        frontend-dev frontend-build frontend-test frontend-install \
        db-migrate db-reset db-shell db-seed \
        init-db seed-data install-deps update-deps lint format \
        ch-shell ch-stats ch-truncate ch-reset \
        import import-basic import-docker import-egrul import-egrip okved-extra \
        pipeline pipeline-basic \
        cluster-up cluster-up-full cluster-down cluster-restart cluster-verify cluster-test \
        cluster-reset cluster-truncate cluster-import cluster-import-okved cluster-fill-mv \
        cluster-detect-changes cluster-optimize cluster-optimize-force cluster-optimize-stats \
        cluster-frontend cluster-backup cluster-restore cluster-logs cluster-ps \
        notifications-up notifications-down notifications-logs notifications-test dev-notifications \
        es-create-indices es-delete-indices es-reindex \
        es-sync-initial es-sync-incremental es-sync-daemon es-sync-stop \
        es-stats es-search-test es-health \
        kafka-topics kafka-create-topic kafka-console \
        minio-console minio-upload \
        adminer redisinsight screenshots \
        monitoring-up monitoring-down monitoring-status monitoring-logs \
        prometheus-reload prometheus-check prometheus-rules-check prometheus-open grafana-open \
        loki-logs promtail-logs loki-query loki-labels

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

up: ## Запуск всей системы (кластер + сервисы)
	@echo "$(CYAN)🚀 Запуск всей системы ЕГРЮЛ/ЕГРИП...$(NC)"
	@echo "$(YELLOW)1/5 Запуск ClickHouse кластера...$(NC)"
	@make cluster-up
	@echo ""
	@echo "$(YELLOW)2/5 Ожидание готовности кластера (60 сек)...$(NC)"
	@sleep 60
	@echo ""
	@echo "$(YELLOW)3/5 Запуск базовых сервисов (Postgres, Kafka, Redis, Elasticsearch)...$(NC)"
	@$(DOCKER_COMPOSE) --profile full up -d postgres redis elasticsearch kafka zookeeper mailhog minio adminer redisinsight
	@sleep 10
	@echo ""
	@echo "$(YELLOW)4/5 Запуск прикладных сервисов...$(NC)"
	@CLICKHOUSE_HOST=clickhouse-01 CLICKHOUSE_USER=egrul_app CLICKHOUSE_PASSWORD=test \
		$(DOCKER_COMPOSE) --profile full up -d api-gateway search-service frontend change-detection-service notification-service sync-service
	@sleep 5
	@echo ""
	@echo "$(YELLOW)5/5 Подключение сервисов к кластерной сети...$(NC)"
	@docker network connect egrul_egrul-cluster-network egrul-api-gateway 2>/dev/null || echo "  ✓ api-gateway уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-frontend 2>/dev/null || echo "  ✓ frontend уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-change-detection 2>/dev/null || echo "  ✓ change-detection-service уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-sync-service 2>/dev/null || echo "  ✓ sync-service уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-search-service 2>/dev/null || echo "  ✓ search-service уже подключен"
	@docker restart egrul-api-gateway egrul-frontend egrul-change-detection egrul-sync-service egrul-search-service > /dev/null 2>&1
	@sleep 3
	@echo ""
	@echo "$(GREEN)✅ Система запущена!$(NC)"
	@echo ""
	@echo "$(CYAN)📊 Доступные сервисы:$(NC)"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - GraphQL Playground: http://localhost:8080/playground"
	@echo "  - MailHog UI: http://localhost:8025"
	@echo "  - MinIO Console: http://localhost:9011"
	@echo "  - Adminer (PostgreSQL): http://localhost:8090"
	@echo "  - RedisInsight: http://localhost:8091"
	@echo ""
	@echo "$(CYAN)📝 Проверка статуса:$(NC)"
	@make cluster-ps
	@$(DOCKER_COMPOSE) ps api-gateway frontend search-service

down: ## Остановка всей системы (все профили)
	@echo "$(YELLOW)🛑 Остановка всей системы...$(NC)"
	@$(DOCKER_COMPOSE) --profile full --profile monitoring --profile tools down
	@$(DOCKER_COMPOSE) -f docker-compose.cluster.yml --profile cluster down
	@echo "$(GREEN)✅ Система полностью остановлена$(NC)"

dev: up ## Запуск в режиме разработки
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

docker-up: up ## Запуск Docker контейнеров (алиас для make up)

docker-down: down ## Остановка Docker контейнеров (алиас для make down)

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

# ==================== ClickHouse Cluster ====================
# Примечание: Single-node режим ClickHouse отключен, используется только кластер

ch-shell: ## Открыть ClickHouse консоль (подключение к node-01)
	@docker exec -it egrul-clickhouse-01 clickhouse-client --user egrul_app --password test

ch-stats: ## Показать статистику ClickHouse кластера
	@echo "$(CYAN)📊 Статистика ClickHouse кластера...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_app --password test --query "\
		SELECT 'Companies' as table, count() as rows FROM egrul.companies UNION ALL \
		SELECT 'Entrepreneurs', count() FROM egrul.entrepreneurs UNION ALL \
		SELECT 'Founders', count() FROM egrul.founders UNION ALL \
		SELECT 'Licenses', count() FROM egrul.licenses UNION ALL \
		SELECT 'Branches', count() FROM egrul.branches \
		FORMAT PrettyCompact"

ch-truncate: cluster-truncate ## Очистить все таблицы ClickHouse кластера (алиас)

ch-reset: cluster-reset ## Пересоздание БД ClickHouse кластера (алиас для cluster-reset; миграции 011-018 применяются автоматически)
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

docker-up-full: up ## Запуск всех сервисов (алиас для make up)

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
	@open http://localhost:9011 || xdg-open http://localhost:9011 || echo "Откройте http://localhost:9011 в браузере"

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

# ==================== Notification System ====================

notifications-up: ## Запуск полной системы уведомлений (full profile)
	@echo "$(CYAN)🔔 Запуск системы уведомлений...$(NC)"
	@$(DOCKER_COMPOSE) --profile full up -d
	@echo "$(YELLOW)⏳ Ожидание готовности Kafka...$(NC)"
	@sleep 5
	@echo "$(CYAN)📝 Создание Kafka топиков...$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --create --topic company-changes --partitions 3 --replication-factor 1 --if-not-exists --bootstrap-server localhost:9092 2>/dev/null || echo "  ✓ company-changes уже существует"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --create --topic entrepreneur-changes --partitions 3 --replication-factor 1 --if-not-exists --bootstrap-server localhost:9092 2>/dev/null || echo "  ✓ entrepreneur-changes уже существует"
	@echo "$(CYAN)🗄️  Применение PostgreSQL миграций...$(NC)"
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d egrul -c "\dt subscriptions.*" -t | grep -q "entity_subscriptions" && echo "  ✓ Миграции уже применены" || \
		($(DOCKER_COMPOSE) exec -T postgres psql -U postgres -d egrul < infrastructure/migrations/postgresql/001_subscriptions.sql && echo "  ✓ Миграция 001_subscriptions применена")
	@echo "$(GREEN)✅ Система уведомлений готова!$(NC)"
	@echo ""
	@echo "$(CYAN)Доступные сервисы:$(NC)"
	@echo "  - Change Detection Service: http://localhost:8082/health"
	@echo "  - Notification Service: http://localhost:8083/health"
	@echo "  - MailHog (SMTP Web UI): http://localhost:8025"
	@echo ""
	@echo "$(CYAN)Kafka топики:$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --list --bootstrap-server localhost:9092 | grep -E "(company|entrepreneur)-changes" || true

notifications-down: ## Остановка сервисов уведомлений
	@echo "$(YELLOW)🛑 Остановка сервисов уведомлений...$(NC)"
	@$(DOCKER_COMPOSE) stop change-detection-service notification-service mailhog
	@echo "$(GREEN)✅ Сервисы остановлены$(NC)"

notifications-logs: ## Просмотр логов сервисов уведомлений
	@echo "$(CYAN)📜 Логи сервисов уведомлений:$(NC)"
	@$(DOCKER_COMPOSE) logs -f change-detection-service notification-service

notifications-test: ## Тестирование системы уведомлений (отправка тестового события)
	@echo "$(CYAN)🧪 Тестирование системы уведомлений...$(NC)"
	@echo "$(YELLOW)1. Проверка статуса сервисов...$(NC)"
	@curl -sf http://localhost:8082/health && echo "  ✓ Change Detection Service: OK" || echo "  ✗ Change Detection Service: FAILED"
	@curl -sf http://localhost:8083/health && echo "  ✓ Notification Service: OK" || echo "  ✗ Notification Service: FAILED"
	@echo ""
	@echo "$(YELLOW)2. Проверка Kafka топиков...$(NC)"
	@$(DOCKER_COMPOSE) exec kafka kafka-topics --list --bootstrap-server localhost:9092 | grep -E "(company|entrepreneur)-changes" && echo "  ✓ Kafka топики созданы" || echo "  ✗ Kafka топики не найдены"
	@echo ""
	@echo "$(YELLOW)3. Проверка PostgreSQL схемы...$(NC)"
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d egrul -c "\dt subscriptions.*" -t | grep -q "entity_subscriptions" && echo "  ✓ PostgreSQL схема subscriptions готова" || echo "  ✗ PostgreSQL схема не найдена"
	@echo ""
	@echo "$(CYAN)Для отправки тестового события используйте:$(NC)"
	@echo "  curl -X POST http://localhost:8082/detect -H 'Content-Type: application/json' -d '{\"entity_type\":\"company\",\"entity_ids\":[\"1234567890123\"]}'"

dev-notifications: ## Запуск с MailHog для разработки
	@echo "$(CYAN)🚀 Запуск системы уведомлений (dev режим с MailHog)...$(NC)"
	@$(DOCKER_COMPOSE) --profile full --profile tools up -d
	@echo "$(GREEN)✅ Сервисы запущены$(NC)"
	@echo ""
	@echo "$(CYAN)MailHog Web UI: $(NC)http://localhost:8025"
	@echo "$(CYAN)Change Detection Service: $(NC)http://localhost:8082/health"
	@echo "$(CYAN)Notification Service: $(NC)http://localhost:8083/health"

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

# ==================== Screenshots ====================

screenshots: ## Скриншоты всех основных страниц через Docker (требует: make up)
	@echo "$(CYAN)📸 Создание скриншотов системы...$(NC)"
	@if ! curl -sf http://localhost:3000 > /dev/null 2>&1; then \
		echo "$(RED)❌ Frontend недоступен. Запустите сначала: make up$(NC)"; \
		exit 1; \
	fi
	@mkdir -p docs/screenshots
	@docker run --rm \
		-e FRONTEND_URL=http://host.docker.internal:3000 \
		-e API_URL=http://host.docker.internal:8080 \
		-e OUTPUT_DIR=/screenshots \
		-e PLAYWRIGHT_BROWSERS_PATH=/ms-playwright \
		-e PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 \
		-v "$(PWD)/scripts/screenshots.mjs:/scripts/screenshots.mjs:ro" \
		-v "$(PWD)/docs/screenshots:/screenshots" \
		mcr.microsoft.com/playwright:v1.50.0-noble \
		bash -c "cd /tmp && npm init -y --silent 2>/dev/null && PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm install playwright@1.50.0 --silent 2>/dev/null && cp /scripts/screenshots.mjs . && node screenshots.mjs"
	@echo "$(GREEN)✅ Скриншоты сохранены в docs/screenshots/$(NC)"

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
	@echo "$(CYAN)⏳ Ожидание готовности кластера (проверка health check)...$(NC)"
	@sleep 30
	@echo "$(CYAN)🔍 Ожидание всех нод (макс 120 сек)...$(NC)"
	@for i in {1..12}; do \
		if docker compose -f docker-compose.cluster.yml ps | grep -E 'clickhouse-0[1-6].*healthy' | wc -l | grep -q 6; then \
			echo "$(GREEN)✅ Все ноды кластера готовы$(NC)"; \
			break; \
		fi; \
		echo "  ⏳ Ожидание... (попытка $$i/12)"; \
		sleep 10; \
	done
	@echo "$(CYAN)📊 Создание базы данных...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "\
		CREATE DATABASE IF NOT EXISTS egrul ON CLUSTER egrul_cluster ENGINE = Atomic" 2>&1 | tail -1
	@echo "$(GREEN)✅ База данных создана на всех нодах$(NC)"
	@echo "$(CYAN)📊 Применение миграции 011 (основные таблицы)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 011 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 012 (change tracking)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/012_change_tracking.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 012 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 013 (MV с правильной логикой статусов)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/013_update_mv_status_logic.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 013 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 014 (исправление MV ликвидаций)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/014_fix_terminations_mv.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 014 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 015 (исправление партиционирования MV)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/015_fix_mv_partitioning.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 015 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 016 (исправление NULL в region)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/016_fix_mv_null_region.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 016 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 017 (ReplicatedAggregatingMergeTree)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/017_replicated_aggregating.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 017 применена$(NC)"
	@echo "$(CYAN)📊 Применение миграции 018 (исправление логики MV ликвидаций)...$(NC)"
	@cat infrastructure/migrations/clickhouse/cluster/018_fix_terminations_mv_logic.sql | \
		docker exec -i egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --multiquery 2>&1 | tail -20
	@echo "$(GREEN)✅ Миграция 018 применена, все таблицы созданы$(NC)"
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

cluster-import: ## Импорт данных в кластер (использует make import + заполняет MV)
	@echo "$(CYAN)📥 Импорт данных в кластер...$(NC)"
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 make import
	@echo "$(CYAN)📊 Заполнение Materialized Views...$(NC)"
	@make cluster-fill-mv
	@echo "$(GREEN)✅ Импорт и заполнение MV завершены$(NC)"

cluster-fill-mv: ## Заполнение Materialized Views данными из основных таблиц
	@echo "$(CYAN)📊 Очистка старых агрегатов...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "TRUNCATE TABLE egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "TRUNCATE TABLE egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "TRUNCATE TABLE egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "TRUNCATE TABLE egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster"
	@echo "$(CYAN)📊 Заполнение stats_companies_by_region (через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "SET max_partitions_per_insert_block = 1000; INSERT INTO egrul.stats_companies_by_region SELECT region_code, coalesce(any(region), '') as region, multiIf(status_code IN ('113', '114', '115', '116', '117'), 'bankrupt', termination_date IS NOT NULL OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'), 'liquidated', 'active') as status, countState() as count, now64(3) as updated_at FROM egrul.companies GROUP BY region_code, status"
	@echo "$(GREEN)✅ stats_companies_by_region заполнена$(NC)"
	@echo "$(CYAN)📊 Заполнение stats_entrepreneurs_by_region (через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "INSERT INTO egrul.stats_entrepreneurs_by_region SELECT region_code, coalesce(any(region), '') as region, if(termination_date IS NULL AND status_code IS NULL, 'active', 'liquidated') as status, countState() as count, now64(3) as updated_at FROM egrul.entrepreneurs GROUP BY region_code, status"
	@echo "$(GREEN)✅ stats_entrepreneurs_by_region заполнена$(NC)"
	@echo "$(CYAN)📊 Заполнение stats_registrations_by_month (компании через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "SET max_partitions_per_insert_block = 1000; INSERT INTO egrul.stats_registrations_by_month SELECT 'company' as entity_type, toStartOfMonth(registration_date) as registration_month, countState() as count, now64(3) as updated_at FROM egrul.companies WHERE registration_date IS NOT NULL GROUP BY registration_month"
	@echo "$(CYAN)📊 Заполнение stats_registrations_by_month (ИП через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "SET max_partitions_per_insert_block = 1000; INSERT INTO egrul.stats_registrations_by_month SELECT 'entrepreneur' as entity_type, toStartOfMonth(registration_date) as registration_month, countState() as count, now64(3) as updated_at FROM egrul.entrepreneurs WHERE registration_date IS NOT NULL GROUP BY registration_month"
	@echo "$(GREEN)✅ stats_registrations_by_month заполнена$(NC)"
	@echo "$(CYAN)📊 Заполнение stats_terminations_by_month (компании через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "SET max_partitions_per_insert_block = 1000; INSERT INTO egrul.stats_terminations_by_month SELECT 'company' as entity_type, toStartOfMonth(COALESCE(termination_date, multiIf(status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'), extract_date, NULL))) as termination_month, countState() as count, now64(3) as updated_at FROM egrul.companies WHERE termination_date IS NOT NULL OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802') GROUP BY termination_month"
	@echo "$(CYAN)📊 Заполнение stats_terminations_by_month (ИП через Distributed)...$(NC)"
	@docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 --query "SET max_partitions_per_insert_block = 1000; INSERT INTO egrul.stats_terminations_by_month SELECT 'entrepreneur' as entity_type, toStartOfMonth(termination_date) as termination_month, countState() as count, now64(3) as updated_at FROM egrul.entrepreneurs WHERE termination_date IS NOT NULL GROUP BY termination_month"
	@echo "$(GREEN)✅ stats_terminations_by_month заполнена$(NC)"

cluster-import-okved: ## Импорт только дополнительных ОКВЭД в кластер
	@echo "$(CYAN)📊 Импорт дополнительных ОКВЭД в кластер...$(NC)"
	@chmod +x infrastructure/scripts/import-okved-extra.sh
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 CLICKHOUSE_USER=egrul_import CLICKHOUSE_PASSWORD=123 \
		./infrastructure/scripts/import-okved-extra.sh
	@echo "$(GREEN)✅ Импорт ОКВЭД завершен$(NC)"

cluster-detect-changes: ## Запуск детектирования изменений вручную
	@echo "$(CYAN)🔍 Запуск детектирования изменений...$(NC)"
	@if ! curl -s -f http://localhost:8082/health > /dev/null 2>&1; then \
		echo "$(RED)❌ Change-detection-service недоступен$(NC)"; \
		echo "$(YELLOW)Запустите: make up$(NC)"; \
		exit 1; \
	fi
	@echo "$(CYAN)Получение списка OGRN с изменениями...$(NC)"
	@OGRNS=$$(docker exec egrul-clickhouse-01 clickhouse-client --query \
		"SELECT arrayJoin(groupArray(ogrn)) FROM (SELECT ogrn FROM egrul.companies GROUP BY ogrn HAVING uniqExact(extract_date) > 1) LIMIT 10000" \
		| jq -Rs 'split("\n") | map(select(length > 0))'); \
	if [ -z "$$OGRNS" ] || [ "$$OGRNS" = "[]" ]; then \
		echo "$(YELLOW)Новых изменений не обнаружено$(NC)"; \
		exit 0; \
	fi; \
	COUNT=$$(echo "$$OGRNS" | jq 'length'); \
	echo "$(CYAN)Обнаружено компаний с изменениями: $$COUNT$(NC)"; \
	curl -X POST http://localhost:8082/detect \
		-H 'Content-Type: application/json' \
		-d "{\"entity_type\": \"company\", \"entity_ids\": $$OGRNS}" | jq .
	@echo "$(GREEN)✅ Детектирование завершено$(NC)"

cluster-optimize: ## Очистка старых версий данных (OPTIMIZE FINAL) - запускать после детектирования!
	@echo "$(CYAN)🧹 Очистка старых версий данных...$(NC)"
	@echo "$(YELLOW)⚠️  ВАЖНО: Эта операция удалит все старые версии данных!$(NC)"
	@echo "$(YELLOW)⚠️  Убедитесь, что детектирование изменений уже выполнено!$(NC)"
	@chmod +x infrastructure/scripts/cleanup-old-versions.sh
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 \
		./infrastructure/scripts/cleanup-old-versions.sh

cluster-optimize-force: ## Очистка старых версий без подтверждения (для автоматизации)
	@echo "$(CYAN)🧹 Очистка старых версий данных (без подтверждения)...$(NC)"
	@chmod +x infrastructure/scripts/cleanup-old-versions.sh
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 FORCE=true \
		./infrastructure/scripts/cleanup-old-versions.sh

cluster-optimize-stats: ## Показать статистику дублей и версий без очистки
	@echo "$(CYAN)📊 Статистика дублей и версий...$(NC)"
	@chmod +x infrastructure/scripts/cleanup-old-versions.sh
	@CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=8123 \
		./infrastructure/scripts/cleanup-old-versions.sh --stats

cluster-frontend: ## Перезапуск frontend и API Gateway с подключением к кластеру (опционально, уже включено в make up)
	@echo "$(CYAN)🌐 Перезапуск frontend и API Gateway...$(NC)"
	@echo "$(YELLOW)Примечание: Эта команда автоматически выполняется при 'make up'$(NC)"
	@echo ""
	@$(DOCKER_COMPOSE) stop api-gateway frontend 2>/dev/null || true
	@echo "$(CYAN)🚀 Пересоздание контейнеров...$(NC)"
	@CLICKHOUSE_HOST=clickhouse-01 CLICKHOUSE_USER=egrul_app CLICKHOUSE_PASSWORD=test $(DOCKER_COMPOSE) up -d --force-recreate --no-deps api-gateway frontend
	@sleep 2
	@echo "$(CYAN)🔗 Подключение к кластерной сети...$(NC)"
	@docker network connect egrul_egrul-cluster-network egrul-api-gateway 2>/dev/null || echo "  ✓ api-gateway уже подключен"
	@docker network connect egrul_egrul-cluster-network egrul-frontend 2>/dev/null || echo "  ✓ frontend уже подключен"
	@docker restart egrul-api-gateway egrul-frontend > /dev/null 2>&1
	@echo "$(GREEN)✅ Сервисы перезапущены$(NC)"
	@sleep 3
	@$(DOCKER_COMPOSE) ps api-gateway frontend

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

# ==================== Docker Network Management ====================

docker-clean-networks: ## Очистка orphan контейнеров и отключение от кластерной сети
	@echo "$(CYAN)🧹 Очистка сетевых конфликтов...$(NC)"
	@echo "$(YELLOW)Отключение контейнеров от кластерной сети...$(NC)"
	@docker network disconnect egrul_egrul-cluster-network egrul-api-gateway 2>/dev/null || true
	@docker network disconnect egrul_egrul-cluster-network egrul-frontend 2>/dev/null || true
	@echo "$(YELLOW)Удаление orphan контейнеров...$(NC)"
	@docker compose down --remove-orphans
	@echo "$(GREEN)✅ Очистка завершена$(NC)"

docker-full-clean: docker-clean-networks ## Полная очистка с удалением volumes
	@echo "$(YELLOW)🗑️  Удаление volumes...$(NC)"
	@docker compose down -v
	@echo "$(GREEN)✅ Полная очистка завершена$(NC)"

# ====================================================================================
# Мониторинг и Observability
# ====================================================================================

monitoring-up: ## Запуск Prometheus + Grafana + cAdvisor + Loki + Promtail
	@echo "$(CYAN)📊 Запуск системы мониторинга...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml --profile monitoring up -d prometheus grafana cadvisor loki promtail
	@echo "$(GREEN)✅ Мониторинг запущен!$(NC)"
	@echo ""
	@echo "Доступные сервисы:"
	@echo "  - Prometheus:  http://localhost:9090"
	@echo "  - Grafana:     http://localhost:3001 (логин: admin/admin)"
	@echo "  - cAdvisor:    http://localhost:8085"
	@echo "  - Loki:        http://localhost:3100"
	@echo ""
	@echo "Проверьте:"
	@echo "  - Prometheus targets: http://localhost:9090/targets"
	@echo "  - Grafana Explore:    http://localhost:3001/explore (выберите Loki)"

monitoring-down: ## Остановка только сервисов мониторинга
	@echo "$(CYAN)🛑 Остановка мониторинга...$(NC)"
	@$(DOCKER_COMPOSE) stop prometheus grafana cadvisor loki promtail
	@echo "$(GREEN)✅ Мониторинг остановлен$(NC)"
	@echo "$(YELLOW)Контейнеры остановлены, но не удалены. Для полного удаления используйте: make down$(NC)"

prometheus-reload: ## Перезагрузка конфигурации Prometheus
	@echo "$(CYAN)🔄 Перезагрузка Prometheus...$(NC)"
	@curl -X POST http://localhost:9090/-/reload 2>/dev/null && \
		echo "$(GREEN)✅ Prometheus перезагружен$(NC)" || \
		echo "$(RED)❌ Ошибка перезагрузки. Prometheus запущен?$(NC)"

prometheus-check: ## Проверка конфигурации Prometheus
	@echo "$(CYAN)🔍 Проверка конфигурации Prometheus...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml exec prometheus promtool check config /etc/prometheus/prometheus.yml

prometheus-rules-check: ## Проверка alert rules
	@echo "$(CYAN)🔍 Проверка alert rules...$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml exec prometheus promtool check rules /etc/prometheus/rules/alerts.yml

grafana-open: ## Открыть Grafana UI в браузере
	@open http://localhost:3001 2>/dev/null || xdg-open http://localhost:3001 2>/dev/null || \
		echo "Grafana UI: http://localhost:3001 (логин: admin/admin)"

prometheus-open: ## Открыть Prometheus UI в браузере
	@open http://localhost:9090 2>/dev/null || xdg-open http://localhost:9090 2>/dev/null || \
		echo "Prometheus UI: http://localhost:9090"

monitoring-status: ## Проверка статуса мониторинга
	@echo "$(CYAN)📊 Статус мониторинга:$(NC)"
	@echo ""
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml ps prometheus grafana cadvisor loki promtail 2>/dev/null || \
		echo "$(YELLOW)Мониторинг не запущен. Запустите: make monitoring-up$(NC)"

monitoring-logs: ## Просмотр логов мониторинга
	@echo "$(CYAN)📄 Логи мониторинга (Ctrl+C для выхода):$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml logs -f prometheus grafana loki promtail

loki-logs: ## Просмотр логов Loki
	@echo "$(CYAN)📄 Логи Loki (Ctrl+C для выхода):$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml logs -f loki

promtail-logs: ## Просмотр логов Promtail
	@echo "$(CYAN)📄 Логи Promtail (Ctrl+C для выхода):$(NC)"
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.cluster.yml logs -f promtail

loki-query: ## Запрос логов через Loki API (использование: make loki-query QUERY='{service="api-gateway"}')
	@echo "$(CYAN)🔍 Запрос логов через Loki...$(NC)"
	@curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
		--data-urlencode 'query=$(or $(QUERY),{service=~".+"})' \
		--data-urlencode 'limit=100' | jq -r '.data.result[].values[][1]' | head -20

loki-labels: ## Показать доступные labels в Loki
	@echo "$(CYAN)🏷️  Доступные labels:$(NC)"
	@curl -s http://localhost:3100/loki/api/v1/labels | jq -r '.data[]'

