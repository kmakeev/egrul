#!/bin/bash
#
# Скрипт полного backup всех шардов ClickHouse кластера в MinIO
#
# Использование:
#   ./backup-all.sh [backup_name]
#
# Примеры:
#   ./backup-all.sh                    # backup_20240116_143022
#   ./backup-all.sh prod_before_deploy # prod_before_deploy
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

# Имя backup (по умолчанию с timestamp)
BACKUP_NAME="${1:-backup_$(date +%Y%m%d_%H%M%S)}"

# Функция выполнения запроса к ClickHouse
clickhouse_query() {
    local query="$1"
    curl -sS "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$query"
}

log_info "=========================================="
log_info "  ClickHouse Cluster Backup"
log_info "=========================================="
log_info ""
log_info "Backup name: $BACKUP_NAME"
log_info "Target: s3_backup://${BACKUP_NAME}/"
log_info ""

# Проверка доступности ClickHouse
log_info "Проверка подключения к ClickHouse..."
if ! clickhouse_query "SELECT 1" > /dev/null 2>&1; then
    log_error "Не удалось подключиться к ClickHouse на ${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}"
    log_error "Проверьте что кластер запущен: docker compose -f docker-compose.cluster.yml ps"
    exit 1
fi
log_info "✓ Подключение установлено"

# Проверка доступности S3 диска
log_info "Проверка S3 backup диска..."
if ! clickhouse_query "SELECT name FROM system.disks WHERE name = 's3_backup'" | grep -q "s3_backup"; then
    log_error "S3 backup диск не настроен"
    log_error "Проверьте конфигурацию в infrastructure/docker/clickhouse-cluster/shared/backup_disk.xml"
    exit 1
fi
log_info "✓ S3 диск доступен"

# Создание backup
log_info ""
log_info "Создание backup всей базы данных egrul..."
log_info "Это может занять несколько минут в зависимости от объема данных..."

# BACKUP команда для всей базы данных
BACKUP_QUERY="
BACKUP DATABASE egrul
TO Disk('s3_backup', '${BACKUP_NAME}/')
SETTINGS async = false
"

if clickhouse_query "$BACKUP_QUERY"; then
    log_info ""
    log_info "✓ Backup успешно создан!"
    log_info ""

    # Получение информации о backup
    log_info "Информация о backup:"
    clickhouse_query "
    SELECT
        name,
        status,
        num_files,
        formatReadableSize(uncompressed_size) as uncompressed,
        formatReadableSize(compressed_size) as compressed,
        formatDateTime(start_time, '%Y-%m-%d %H:%M:%S') as start_time,
        formatDateTime(end_time, '%Y-%m-%d %H:%M:%S') as end_time,
        dateDiff('second', start_time, end_time) as duration_sec
    FROM system.backups
    WHERE name LIKE '%${BACKUP_NAME}%'
    ORDER BY end_time DESC
    LIMIT 1
    FORMAT Vertical
    "

    log_info ""
    log_info "Backup сохранен в MinIO:"
    log_info "  Bucket: backups/clickhouse/${BACKUP_NAME}/"
    log_info "  MinIO Console: http://localhost:9001"
    log_info ""
    log_info "Для восстановления используйте:"
    log_info "  ./infrastructure/scripts/backup/restore-all.sh ${BACKUP_NAME}"

else
    log_error "Ошибка создания backup"
    exit 1
fi

log_info ""
log_info "=========================================="
log_info "  Backup завершен успешно"
log_info "=========================================="
