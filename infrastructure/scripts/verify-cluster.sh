#!/bin/bash
#
# Скрипт быстрой проверки состояния ClickHouse кластера
#
# Проверяет:
# - Доступность всех нод
# - Видимость кластера
# - Состояние репликации
# - Готовность к импорту данных
#

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-egrul_app}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-test}"

clickhouse_query() {
    curl -sS "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$1"
}

log_info "=========================================="
log_info "  ClickHouse Cluster Verification"
log_info "=========================================="
log_info ""

# Проверка 1: Подключение к ClickHouse
log_info "[1/4] Проверка подключения к ClickHouse..."
if clickhouse_query "SELECT version()" > /dev/null 2>&1; then
    log_info "✓ ClickHouse доступен"
else
    log_error "✗ Не удалось подключиться к ClickHouse"
    exit 1
fi

# Проверка 2: Видимость кластера
log_info ""
log_info "[2/4] Проверка конфигурации кластера..."
CLUSTER_NODES=$(clickhouse_query "
SELECT count()
FROM system.clusters
WHERE cluster = 'egrul_cluster'
" 2>/dev/null)

if [ "$CLUSTER_NODES" = "6" ]; then
    log_info "✓ Кластер egrul_cluster виден (6 нод)"
    clickhouse_query "
    SELECT
        host_name,
        shard_num as shard,
        replica_num as replica
    FROM system.clusters
    WHERE cluster = 'egrul_cluster'
    ORDER BY shard_num, replica_num
    FORMAT PrettyCompactMonoBlock
    "
else
    log_error "✗ Кластер настроен неправильно (ожидается 6 нод, найдено: $CLUSTER_NODES)"
    exit 1
fi

# Проверка 3: Состояние репликации
log_info ""
log_info "[3/4] Проверка состояния репликации..."
REPLICA_COUNT=$(clickhouse_query "
SELECT count()
FROM system.replicas
WHERE database = 'egrul'
" 2>/dev/null || echo "0")

if [ "$REPLICA_COUNT" -gt "0" ]; then
    log_info "✓ Найдено $REPLICA_COUNT реплицированных таблиц"
    clickhouse_query "
    SELECT
        table,
        is_leader,
        total_replicas,
        active_replicas
    FROM system.replicas
    WHERE database = 'egrul'
    ORDER BY table
    FORMAT PrettyCompactMonoBlock
    "
else
    log_warning "⚠ Реплицированные таблицы не найдены"
    log_warning "Возможно миграции еще не применены"
fi

# Проверка 4: Keeper подключение
log_info ""
log_info "[4/4] Проверка подключения к ClickHouse Keeper..."
if clickhouse_query "SELECT * FROM system.zookeeper WHERE path = '/'" > /dev/null 2>&1; then
    log_info "✓ Keeper доступен"
else
    log_error "✗ Не удалось подключиться к Keeper"
    log_error "Проверьте что Keeper ноды запущены"
    exit 1
fi

log_info ""
log_info "=========================================="
log_info "  Кластер готов к работе"
log_info "=========================================="
log_info ""
log_info "Следующие шаги:"
log_info "  1. Импорт данных: make cluster-import"
log_info "  2. Тестирование: make cluster-test"
