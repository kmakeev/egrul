FROM clickhouse/clickhouse-server:24.11-alpine

# Копируем индивидуальную конфигурацию ноды
COPY config.xml /etc/clickhouse-server/config.xml

# Конфигурационные файлы будут подключены через docker-compose volumes:
# - shared/remote_servers.xml
# - shared/backup_disk.xml
# - shared/users.xml
