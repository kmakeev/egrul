#!/bin/bash
set -e

ELASTICSEARCH_URL=${ELASTICSEARCH_URL:-http://localhost:9200}

echo "Starting full reindexing process..."
echo ""

# Удаление старых индексов
echo "Step 1: Deleting old indices..."
if curl -s -X DELETE "$ELASTICSEARCH_URL/egrul_*" | grep -q "acknowledged"; then
    echo "✓ Old indices deleted"
else
    echo "⚠ No indices to delete or deletion failed"
fi

echo ""

# Создание новых индексов
echo "Step 2: Creating new indices..."
./infrastructure/scripts/es-create-indices.sh

echo ""

# Запуск initial sync
echo "Step 3: Starting initial data sync..."
echo "This may take several minutes depending on data size..."
echo ""

if docker compose ps sync-service | grep -q "Up"; then
    echo "Using running sync-service container..."
    docker compose exec sync-service ./sync-service --mode=initial
else
    echo "Starting one-time sync-service container..."
    docker compose run --rm sync-service ./sync-service --mode=initial
fi

echo ""
echo "✓ Reindexing completed successfully!"
echo ""
echo "Final statistics:"
curl -s "$ELASTICSEARCH_URL/_cat/indices/egrul_*?v&h=index,docs.count,store.size"
