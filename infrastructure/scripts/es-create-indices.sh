#!/bin/bash
set -e

ELASTICSEARCH_URL=${ELASTICSEARCH_URL:-http://localhost:9200}

echo "Creating Elasticsearch indices..."

# Проверка доступности Elasticsearch
echo "Checking Elasticsearch connection..."
if ! curl -s "$ELASTICSEARCH_URL/_cluster/health" > /dev/null; then
    echo "Error: Cannot connect to Elasticsearch at $ELASTICSEARCH_URL"
    exit 1
fi

echo "Elasticsearch is available"

# Создание индекса companies
echo "Creating index: egrul_companies..."
if curl -s -X PUT "$ELASTICSEARCH_URL/egrul_companies" \
  -H 'Content-Type: application/json' \
  -d @infrastructure/elasticsearch/mappings/companies.json | grep -q '"acknowledged":true'; then
    echo "✓ Index egrul_companies created successfully"
else
    echo "⚠ Index egrul_companies might already exist or creation failed"
fi

# Создание индекса entrepreneurs
echo "Creating index: egrul_entrepreneurs..."
if curl -s -X PUT "$ELASTICSEARCH_URL/egrul_entrepreneurs" \
  -H 'Content-Type: application/json' \
  -d @infrastructure/elasticsearch/mappings/entrepreneurs.json | grep -q '"acknowledged":true'; then
    echo "✓ Index egrul_entrepreneurs created successfully"
else
    echo "⚠ Index egrul_entrepreneurs might already exist or creation failed"
fi

echo ""
echo "Indices created. Checking status..."
curl -s "$ELASTICSEARCH_URL/_cat/indices/egrul_*?v"

echo ""
echo "✓ All done!"
