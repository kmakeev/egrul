#!/bin/bash
# ============================================================================
# ClickHouse Migration Script для ЕГРЮЛ/ЕГРИП
# ============================================================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Настройки по умолчанию
CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-9000}"
CLICKHOUSE_HTTP_PORT="${CLICKHOUSE_HTTP_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-admin}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-admin}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$(dirname "$0")/../migrations/clickhouse}"

# Функция для вывода сообщений
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

# Функция проверки доступности ClickHouse
check_clickhouse() {
    log_info "Проверка подключения к ClickHouse..."
    
    if command -v clickhouse-client &> /dev/null; then
        # Используем native клиент
        if clickhouse-client --host "$CLICKHOUSE_HOST" --port "$CLICKHOUSE_PORT" \
            --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
            --query "SELECT 1" &> /dev/null; then
            log_success "ClickHouse доступен (native protocol)"
            return 0
        fi
    fi
    
    # Fallback на HTTP
    if curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/ping" &> /dev/null; then
        log_success "ClickHouse доступен (HTTP protocol)"
        return 0
    fi
    
    log_error "Не удалось подключиться к ClickHouse"
    return 1
}

# Функция выполнения SQL файла
execute_migration() {
    local file="$1"
    local filename=$(basename "$file")
    
    log_info "Применение миграции: $filename"
    
    if command -v clickhouse-client &> /dev/null; then
        # Используем native клиент
        if clickhouse-client --host "$CLICKHOUSE_HOST" --port "$CLICKHOUSE_PORT" \
            --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
            --multiquery < "$file"; then
            log_success "Миграция $filename успешно применена"
            return 0
        else
            log_error "Ошибка при применении миграции $filename"
            return 1
        fi
    else
        # Используем curl для HTTP API
        local sql_content=$(cat "$file")
        local response=$(curl -s -X POST \
            "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/?user=${CLICKHOUSE_USER}&password=${CLICKHOUSE_PASSWORD}" \
            --data-binary "$sql_content")
        
        if [ -z "$response" ]; then
            log_success "Миграция $filename успешно применена"
            return 0
        else
            log_error "Ошибка при применении миграции $filename: $response"
            return 1
        fi
    fi
}

# Основная функция применения миграций
run_migrations() {
    log_info "Запуск миграций из директории: $MIGRATIONS_DIR"
    
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_error "Директория миграций не найдена: $MIGRATIONS_DIR"
        exit 1
    fi
    
    # Получаем список SQL файлов, отсортированных по имени
    local migration_files=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort)
    
    if [ -z "$migration_files" ]; then
        log_warning "Файлы миграций не найдены"
        exit 0
    fi
    
    local total=0
    local success=0
    local failed=0
    
    for file in $migration_files; do
        total=$((total + 1))
        if execute_migration "$file"; then
            success=$((success + 1))
        else
            failed=$((failed + 1))
            if [ "$STOP_ON_ERROR" = "true" ]; then
                log_error "Остановка из-за ошибки"
                exit 1
            fi
        fi
        echo ""
    done
    
    echo "============================================"
    log_info "Результаты миграции:"
    log_info "  Всего файлов: $total"
    log_success "  Успешно: $success"
    if [ $failed -gt 0 ]; then
        log_error "  С ошибками: $failed"
    fi
    echo "============================================"
    
    if [ $failed -gt 0 ]; then
        exit 1
    fi
}

# Функция отображения статуса базы данных
show_status() {
    log_info "Статус базы данных egrul:"
    
    local query="
    SELECT 
        'companies' as table_name,
        count(*) as rows
    FROM egrul.companies
    UNION ALL
    SELECT 
        'entrepreneurs',
        count(*)
    FROM egrul.entrepreneurs
    UNION ALL
    SELECT 
        'company_history',
        count(*)
    FROM egrul.company_history
    UNION ALL
    SELECT 
        'ownership_graph',
        count(*)
    FROM egrul.ownership_graph
    FORMAT Pretty
    "
    
    if command -v clickhouse-client &> /dev/null; then
        clickhouse-client --host "$CLICKHOUSE_HOST" --port "$CLICKHOUSE_PORT" \
            --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" \
            --query "$query" 2>/dev/null || log_warning "Таблицы еще не созданы"
    else
        curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_HTTP_PORT}/?user=${CLICKHOUSE_USER}&password=${CLICKHOUSE_PASSWORD}" \
            --data-binary "$query" 2>/dev/null || log_warning "Таблицы еще не созданы"
    fi
}

# Функция отображения справки
show_help() {
    echo "ClickHouse Migration Script для ЕГРЮЛ/ЕГРИП"
    echo ""
    echo "Использование: $0 [ОПЦИЯ]"
    echo ""
    echo "Опции:"
    echo "  migrate      Применить все миграции"
    echo "  status       Показать статус базы данных"
    echo "  check        Проверить подключение к ClickHouse"
    echo "  help         Показать эту справку"
    echo ""
    echo "Переменные окружения:"
    echo "  CLICKHOUSE_HOST      Хост ClickHouse (по умолчанию: localhost)"
    echo "  CLICKHOUSE_PORT      Порт native protocol (по умолчанию: 9000)"
    echo "  CLICKHOUSE_HTTP_PORT Порт HTTP protocol (по умолчанию: 8123)"
    echo "  CLICKHOUSE_USER      Пользователь (по умолчанию: admin)"
    echo "  CLICKHOUSE_PASSWORD  Пароль (по умолчанию: admin)"
    echo "  MIGRATIONS_DIR       Директория с миграциями"
    echo "  STOP_ON_ERROR        Остановка при ошибке (true/false)"
    echo ""
    echo "Примеры:"
    echo "  $0 migrate"
    echo "  CLICKHOUSE_HOST=clickhouse $0 migrate"
    echo "  $0 status"
}

# Обработка аргументов командной строки
case "${1:-migrate}" in
    migrate)
        check_clickhouse
        run_migrations
        ;;
    status)
        check_clickhouse
        show_status
        ;;
    check)
        check_clickhouse
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Неизвестная команда: $1"
        show_help
        exit 1
        ;;
esac

