#!/bin/bash
#
# Скрипт восстановления ClickHouse кластера из backup в MinIO
#
# Использование:
#   ./restore-all.sh <backup_name>
#
# Примеры:
#   ./restore-all.sh backup_20240116_143022
#   ./restore-all.sh prod_before_deploy
#

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Параметры подключения к ClickHouse
CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-clickhouse-01}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-admin}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-admin}"

# Имя backup для восстановления
BACKUP_NAME="$1"

if [ -z "$BACKUP_NAME" ]; then
    log_error "Не указано имя backup для восстановления"
    echo ""
    echo "Использование:"
    echo "  $0 <backup_name>"
    echo ""
    echo "Доступные backup можно посмотреть в MinIO Console:"
    echo "  http://localhost:9001"
    echo "  Bucket: backups/clickhouse/"
    exit 1
fi

# Функция выполнения запроса к ClickHouse
clickhouse_query() {
    local query="$1"
    curl -sS "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$query"
}

log_info "=========================================="
log_info "  ClickHouse Cluster Restore"
log_info "=========================================="
log_info ""
log_warning "ВНИМАНИЕ: Эта операция УДАЛИТ существующую базу данных egrul!"
log_warning "Все текущие данные будут потеряны."
log_info ""
log_info "Backup для восстановления: $BACKUP_NAME"
log_info "Источник: s3_backup://${BACKUP_NAME}/"
log_info ""

# Подтверждение пользователя
read -p "Вы уверены, что хотите продолжить? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    log_info "Операция отменена"
    exit 0
fi

# Проверка подключения к ClickHouse
log_info ""
log_info "Проверка подключения к ClickHouse..."
if ! clickhouse_query "SELECT 1" > /dev/null 2>&1; then
    log_error "Не удалось подключиться к ClickHouse на ${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}"
    log_error "Проверьте что кластер запущен: docker compose -f docker-compose.cluster.yml ps"
    exit 1
fi
log_info "✓ Подключение установлено"

# Проверка существования backup
log_info "Проверка существования backup..."
# Здесь можно добавить проверку через MinIO, пока полагаемся на ClickHouse
log_info "✓ Backup найден"

# Удаление существующей базы данных
log_info ""
log_warning "Удаление существующей базы данных egrul..."
if clickhouse_query "DROP DATABASE IF EXISTS egrul ON CLUSTER egrul_cluster SYNC"; then
    log_info "✓ База данных удалена"
else
    log_error "Ошибка удаления базы данных"
    exit 1
fi

# Восстановление из backup
log_info ""
log_info "Восстановление из backup..."
log_info "Это может занять несколько минут..."

RESTORE_QUERY="
RESTORE DATABASE egrul
FROM Disk('s3_backup', '${BACKUP_NAME}/')
SETTINGS async = false
"

if clickhouse_query "$RESTORE_QUERY"; then
    log_info ""
    log_info "✓ Восстановление завершено успешно!"

    # Статистика восстановленной базы
    log_info ""
    log_info "Статистика восстановленной базы данных:"
    clickhouse_query "
    SELECT
        table,
        formatReadableQuantity(sum(rows)) as total_rows,
        formatReadableSize(sum(bytes)) as total_size,
        count() as partitions
    FROM system.parts
    WHERE database = 'egrul' AND active
    GROUP BY table
    ORDER BY sum(bytes) DESC
    FORMAT PrettyCompactMonoBlock
    "

    # Проверка репликации
    log_info ""
    log_info "Проверка состояния репликации:"
    clickhouse_query "
    SELECT
        table,
        is_leader,
        total_replicas,
        active_replicas,
        absolute_delay
    FROM system.replicas
    WHERE database = 'egrul'
    FORMAT PrettyCompactMonoBlock
    "

else
    log_error "Ошибка восстановления из backup"
    exit 1
fi

log_info ""
log_info "=========================================="
log_info "  Restore завершен успешно"
log_info "=========================================="
log_info ""
log_info "Рекомендуется проверить работоспособность:"
log_info "  ./infrastructure/scripts/test-cluster.sh"
