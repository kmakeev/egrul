#!/bin/bash
# ==============================================================================
# Скрипт очистки старых версий данных (OPTIMIZE FINAL)
# ==============================================================================
# Удаляет дубликаты и старые версии из ReplacingMergeTree таблиц
# ВАЖНО: Запускать только ПОСЛЕ детектирования изменений!
# ==============================================================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация по умолчанию
CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-egrul_import}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-123}"
CLICKHOUSE_DATABASE="${CLICKHOUSE_DATABASE:-egrul}"

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Функция для выполнения запросов к ClickHouse
clickhouse_query() {
    local query="$1"
    curl -sS "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$query"
}

# Функция для получения статистики до очистки
get_stats_before() {
    log_info "Сбор статистики до очистки..."

    echo ""
    echo "=== Статистика ДО очистки ==="
    echo ""

    echo "Компании:"
    clickhouse_query "
        SELECT
            'Total records' as metric,
            formatReadableQuantity(COUNT(*)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        UNION ALL
        SELECT
            'Unique OGRNs' as metric,
            formatReadableQuantity(uniqExact(ogrn)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        UNION ALL
        SELECT
            'Duplicates (same OGRN)' as metric,
            formatReadableQuantity(COUNT(*) - uniqExact(ogrn)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        UNION ALL
        SELECT
            'Records with multiple versions' as metric,
            formatReadableQuantity(COUNT(*)) as value
        FROM (
            SELECT ogrn
            FROM ${CLICKHOUSE_DATABASE}.companies
            GROUP BY ogrn
            HAVING COUNT(*) > 1
        )
        FORMAT Pretty
    "

    echo ""
    echo "ИП:"
    clickhouse_query "
        SELECT
            'Total records' as metric,
            formatReadableQuantity(COUNT(*)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        UNION ALL
        SELECT
            'Unique OGRNIPs' as metric,
            formatReadableQuantity(uniqExact(ogrnip)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        UNION ALL
        SELECT
            'Duplicates (same OGRNIP)' as metric,
            formatReadableQuantity(COUNT(*) - uniqExact(ogrnip)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        FORMAT Pretty
    "

    echo ""
}

# Функция для получения статистики после очистки
get_stats_after() {
    log_info "Сбор статистики после очистки..."

    echo ""
    echo "=== Статистика ПОСЛЕ очистки ==="
    echo ""

    echo "Компании:"
    clickhouse_query "
        SELECT
            'Total records' as metric,
            formatReadableQuantity(COUNT(*)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        UNION ALL
        SELECT
            'Unique OGRNs' as metric,
            formatReadableQuantity(uniqExact(ogrn)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        UNION ALL
        SELECT
            'Duplicates remaining' as metric,
            formatReadableQuantity(COUNT(*) - uniqExact(ogrn)) as value
        FROM ${CLICKHOUSE_DATABASE}.companies
        FORMAT Pretty
    "

    echo ""
    echo "ИП:"
    clickhouse_query "
        SELECT
            'Total records' as metric,
            formatReadableQuantity(COUNT(*)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        UNION ALL
        SELECT
            'Unique OGRNIPs' as metric,
            formatReadableQuantity(uniqExact(ogrnip)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        UNION ALL
        SELECT
            'Duplicates remaining' as metric,
            formatReadableQuantity(COUNT(*) - uniqExact(ogrnip)) as value
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
        FORMAT Pretty
    "

    echo ""
}

# Функция для выполнения OPTIMIZE
optimize_tables() {
    log_info "Запуск OPTIMIZE FINAL для всех таблиц..."
    echo ""

    # Предупреждение
    log_warning "OPTIMIZE FINAL - это тяжелая операция, которая может занять длительное время!"
    log_warning "Будут удалены все старые версии данных (кроме последних по extract_date)"

    if [ "${FORCE:-false}" != "true" ]; then
        echo ""
        read -p "Продолжить? (yes/no): " confirm
        if [ "$confirm" != "yes" ]; then
            log_info "Операция отменена"
            exit 0
        fi
    fi

    echo ""
    log_info "Очистка компаний..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.companies_local ON CLUSTER egrul_cluster FINAL"
    log_success "Компании оптимизированы"

    echo ""
    log_info "Очистка ИП..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs_local ON CLUSTER egrul_cluster FINAL"
    log_success "ИП оптимизированы"

    echo ""
    log_info "Очистка учредителей..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.founders_local ON CLUSTER egrul_cluster FINAL"
    log_success "Учредители оптимизированы"

    echo ""
    log_info "Очистка истории изменений..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.company_history_local ON CLUSTER egrul_cluster FINAL"
    log_success "История изменений оптимизирована"

    echo ""
    log_info "Очистка лицензий..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.licenses_local ON CLUSTER egrul_cluster FINAL"
    log_success "Лицензии оптимизированы"

    echo ""
    log_info "Очистка филиалов..."
    clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.branches_local ON CLUSTER egrul_cluster FINAL"
    log_success "Филиалы оптимизированы"
}

# Главная функция
main() {
    echo ""
    echo "======================================"
    echo "  Очистка старых версий данных"
    echo "======================================"
    echo ""

    # Проверка подключения
    log_info "Проверка подключения к ClickHouse..."
    local result=$(clickhouse_query "SELECT 1")
    if [ "$result" != "1" ]; then
        log_error "Не удалось подключиться к ClickHouse"
        exit 1
    fi
    log_success "Подключение успешно"

    echo ""

    # Статистика до
    get_stats_before

    # OPTIMIZE
    optimize_tables

    # Даем время на завершение фоновых операций
    log_info "Ожидание завершения фоновых операций слияния (30 секунд)..."
    sleep 30

    # Статистика после
    get_stats_after

    echo ""
    log_success "Очистка завершена!"
    log_info "Старые версии удалены, осталась только последняя версия каждой записи"
}

# Обработка аргументов
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать справку"
        echo "  --force        Не запрашивать подтверждение"
        echo "  --stats        Показать только статистику"
        echo ""
        echo "Переменные окружения:"
        echo "  CLICKHOUSE_HOST      Хост ClickHouse (default: localhost)"
        echo "  CLICKHOUSE_PORT      HTTP порт (default: 8123)"
        echo "  CLICKHOUSE_USER      Пользователь (default: egrul_import)"
        echo "  CLICKHOUSE_PASSWORD  Пароль (default: 123)"
        echo "  CLICKHOUSE_DATABASE  База данных (default: egrul)"
        echo ""
        echo "ВАЖНО: Запускайте этот скрипт только ПОСЛЕ детектирования изменений!"
        echo "       Иначе можно потерять историю до того, как она будет обработана."
        exit 0
        ;;
    --stats)
        get_stats_before
        ;;
    --force)
        FORCE=true
        main
        ;;
    *)
        main
        ;;
esac
