#!/bin/bash
# ==============================================================================
# Batch-выгрузка дополнительных ОКВЭД из JSON поля additional_activities
# в отдельные таблицы egrul.companies_okved_additional и
# egrul.entrepreneurs_okved_additional.
# ==============================================================================

set -e

CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-localhost}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-admin}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-admin}"
CLICKHOUSE_DATABASE="${CLICKHOUSE_DATABASE:-egrul}"

BUCKETS="${OKVED_BUCKETS:-100}"
MEMORY_LIMIT="${OKVED_MAX_MEMORY:-2000000000}" # 2 ГБ по умолчанию

clickhouse_query() {
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$1"
}

echo "📊 Выгрузка дополнительных ОКВЭД (компании и ИП) батчами..."
echo "   Бакетов: ${BUCKETS}, лимит памяти: ${MEMORY_LIMIT}"

for kind in companies entrepreneurs; do
  echo ""
  echo "=== Обработка ${kind} ==="

  for ((bucket=0; bucket<BUCKETS; bucket++)); do
    echo "  → Бакет ${bucket}/${BUCKETS}"

    if [ "$kind" = "companies" ]; then
      clickhouse_query "
      INSERT INTO ${CLICKHOUSE_DATABASE}.companies_okved_additional
      SELECT
          ogrn,
          inn,
          JSONExtractString(x, 'code') AS okved_code,
          JSONExtractString(x, 'name') AS okved_name
      FROM
      (
          SELECT
              ogrn,
              inn,
              arrayJoin(JSONExtractArrayRaw(coalesce(additional_activities, '[]'))) AS x
          FROM ${CLICKHOUSE_DATABASE}.companies
          WHERE additional_activities IS NOT NULL
            AND additional_activities != ''
            AND cityHash64(ogrn) % ${BUCKETS} = ${bucket}
      )
      SETTINGS max_memory_usage=${MEMORY_LIMIT};
      "
    else
      clickhouse_query "
      INSERT INTO ${CLICKHOUSE_DATABASE}.entrepreneurs_okved_additional
      SELECT
          ogrnip,
          inn,
          JSONExtractString(x, 'code') AS okved_code,
          JSONExtractString(x, 'name') AS okved_name
      FROM
      (
          SELECT
              ogrnip,
              inn,
              arrayJoin(JSONExtractArrayRaw(coalesce(additional_activities, '[]'))) AS x
          FROM ${CLICKHOUSE_DATABASE}.entrepreneurs
          WHERE additional_activities IS NOT NULL
            AND additional_activities != ''
            AND cityHash64(ogrnip) % ${BUCKETS} = ${bucket}
      )
      SETTINGS max_memory_usage=${MEMORY_LIMIT};
      "
    fi
  done
done

echo ""
echo "✅ Выгрузка дополнительных ОКВЭД завершена."


