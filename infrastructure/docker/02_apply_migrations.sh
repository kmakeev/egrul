#!/bin/bash
# Скрипт автоматического применения PostgreSQL миграций при инициализации контейнера

set -e

echo "=== Применение PostgreSQL миграций ==="

MIGRATIONS_DIR="/docker-entrypoint-initdb.d/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "Папка миграций не найдена: $MIGRATIONS_DIR"
  exit 1
fi

# Получить список всех .sql файлов и отсортировать
MIGRATIONS=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort)

if [ -z "$MIGRATIONS" ]; then
  echo "Миграции не найдены в $MIGRATIONS_DIR"
  exit 0
fi

# Применить каждую миграцию
for migration in $MIGRATIONS; do
  filename=$(basename "$migration")
  echo "Применение миграции: $filename"

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$migration"

  if [ $? -eq 0 ]; then
    echo "✓ Миграция $filename применена успешно"
  else
    echo "✗ Ошибка при применении миграции $filename"
    exit 1
  fi
done

echo "=== Все миграции применены успешно ==="
