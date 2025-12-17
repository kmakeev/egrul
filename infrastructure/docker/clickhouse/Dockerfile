# ClickHouse Docker Image для ЕГРЮЛ/ЕГРИП
FROM clickhouse/clickhouse-server:24.11-alpine

# Метаданные
LABEL maintainer="egrul-project"
LABEL description="ClickHouse server optimized for EGRUL/EGRIP data"

# Копируем конфигурационные файлы
COPY config.xml /etc/clickhouse-server/config.d/custom-config.xml
COPY users.xml /etc/clickhouse-server/users.d/custom-users.xml

# Создаем директории для данных
RUN mkdir -p /var/lib/clickhouse/data \
    && mkdir -p /var/lib/clickhouse/metadata \
    && mkdir -p /var/lib/clickhouse/tmp \
    && mkdir -p /var/log/clickhouse-server \
    && chown -R clickhouse:clickhouse /var/lib/clickhouse \
    && chown -R clickhouse:clickhouse /var/log/clickhouse-server

# Порты: HTTP, Native, межкластерный
EXPOSE 8123 9000 9009

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=5 \
    CMD clickhouse-client --query "SELECT 1" || exit 1

# Точка входа по умолчанию
USER clickhouse

