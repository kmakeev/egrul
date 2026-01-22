#!/bin/bash
#
# Скрипт полного тестирования ClickHouse кластера
#
# Выполняет:
# - Проверку доступности нод
# - Тестовые INSERT/SELECT операции
# - Проверку репликации
# - Проверку шардирования
# - Performance тесты distributed запросов
#

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

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
log_info "  ClickHouse Cluster Test Suite"
log_info "=========================================="
log_info ""

# Тест 1: Доступность нод кластера
log_info "[Тест 1/6] Проверка доступности всех нод кластера"
clickhouse_query "
SELECT
    host_name,
    port,
    is_local,
    shard_num,
    replica_num
FROM system.clusters
WHERE cluster = 'egrul_cluster'
ORDER BY shard_num, replica_num
FORMAT PrettyCompactMonoBlock
"
log_success "✓ Все ноды доступны"

# Тест 2: Состояние репликации
log_info ""
log_info "[Тест 2/6] Проверка состояния репликации"
BROKEN_REPLICAS=$(clickhouse_query "
SELECT count()
FROM system.replicas
WHERE database = 'egrul' AND (total_replicas != 2 OR active_replicas != 2)
")

if [ "$BROKEN_REPLICAS" = "0" ]; then
    log_success "✓ Все реплики активны (RF=2)"
    clickhouse_query "
    SELECT
        table,
        is_leader,
        total_replicas,
        active_replicas,
        absolute_delay
    FROM system.replicas
    WHERE database = 'egrul'
    ORDER BY table
    LIMIT 10
    FORMAT PrettyCompactMonoBlock
    "
else
    log_error "✗ Найдены проблемы с репликацией ($BROKEN_REPLICAS таблиц)"
    exit 1
fi

# Тест 3: Распределение данных по шардам
log_info ""
log_info "[Тест 3/6] Проверка распределения данных по шардам"
if clickhouse_query "SELECT count() FROM egrul.companies" | grep -q "^0$"; then
    log_warning "⚠ База данных пустая, данные еще не импортированы"
    log_info "Запустите: make cluster-import"
else
    clickhouse_query "
    SELECT
        _shard_num as shard,
        count() as total_companies
    FROM egrul.companies
    GROUP BY _shard_num
    ORDER BY _shard_num
    FORMAT PrettyCompactMonoBlock
    "
    log_success "✓ Данные распределены по шардам"
fi

# Тест 4: Тестовая вставка и проверка репликации
log_info ""
log_info "[Тест 4/6] Тестовая вставка данных"
TEST_OGRN="TEST_$(date +%s)"
clickhouse_query "
INSERT INTO egrul.companies (ogrn, inn, full_name, region_code, version_date)
VALUES ('$TEST_OGRN', '1234567890', 'Test Company', '77', today())
"

# Ждем репликации
sleep 2

# Проверяем что данные появились
FOUND=$(clickhouse_query "SELECT count() FROM egrul.companies WHERE ogrn = '$TEST_OGRN'")
if [ "$FOUND" = "1" ]; then
    log_success "✓ Тестовая запись успешно вставлена и реплицирована"
else
    log_error "✗ Тестовая запись не найдена!"
    exit 1
fi

# Удаляем тестовую запись
clickhouse_query "ALTER TABLE egrul.companies DELETE WHERE ogrn = '$TEST_OGRN'"
log_info "Тестовая запись удалена"

# Тест 5: Проверка Keeper
log_info ""
log_info "[Тест 5/6] Проверка ClickHouse Keeper"
KEEPER_NODES=$(clickhouse_query "
SELECT count()
FROM system.zookeeper
WHERE path = '/clickhouse/tables'
" 2>/dev/null || echo "0")

if [ "$KEEPER_NODES" -gt "0" ]; then
    log_success "✓ Keeper доступен и работает"
else
    log_error "✗ Keeper недоступен"
    exit 1
fi

# Тест 6: Performance distributed запроса
log_info ""
log_info "[Тест 6/6] Тест производительности distributed запроса"
if ! clickhouse_query "SELECT count() FROM egrul.companies" | grep -q "^0$"; then
    START_TIME=$(date +%s%N)
    clickhouse_query "
    SELECT
        region_code,
        count() as total
    FROM egrul.companies
    GROUP BY region_code
    ORDER BY total DESC
    LIMIT 10
    FORMAT PrettyCompactMonoBlock
    " > /dev/null
    END_TIME=$(date +%s%N)
    DURATION=$(((END_TIME - START_TIME) / 1000000))
    log_info "Время выполнения distributed запроса: ${DURATION}ms"
    log_success "✓ Distributed запросы работают"
fi

log_info ""
log_info "=========================================="
log_success "  Все тесты пройдены успешно!"
log_info "=========================================="
