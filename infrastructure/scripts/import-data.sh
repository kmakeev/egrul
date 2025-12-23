#!/bin/bash
# ==============================================================================
# Скрипт импорта данных из Parquet файлов в ClickHouse
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
CLICKHOUSE_USER="${CLICKHOUSE_USER:-admin}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-admin}"
CLICKHOUSE_DATABASE="${CLICKHOUSE_DATABASE:-egrul}"

# Путь к debug-логу для отладки (NDJSON)
DEBUG_LOG_PATH="/Users/konstantin/cursor/egrul/.cursor/debug.log"

DATA_DIR="${DATA_DIR:-./output}"

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
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "$query"
}

# Функция для загрузки Parquet файла
import_parquet() {
    local file="$1"
    local table="$2"
    
    if [ ! -f "$file" ]; then
        log_warning "Файл не найден: $file"
        return 1
    fi
    
    local file_size=$(du -h "$file" | cut -f1)
    log_info "Загрузка $file ($file_size) в таблицу $table..."
    
    # Загрузка через HTTP интерфейс
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/?query=INSERT%20INTO%20${CLICKHOUSE_DATABASE}.${table}%20FORMAT%20Parquet" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "@$file"
    
    if [ $? -eq 0 ]; then
        log_success "Файл $file загружен успешно"
        return 0
    else
        log_error "Ошибка загрузки файла $file"
        return 1
    fi
}

# Функция для импорта учредителей из временной таблицы компаний
import_founders_from_companies_import() {
    log_info "Импорт учредителей из данных ЕГРЮЛ..."
    
    # Очищаем таблицу учредителей (для MVP - полная перезагрузка)
    clickhouse_query "TRUNCATE TABLE ${CLICKHOUSE_DATABASE}.founders"
    
    # Сначала создаем временную таблицу для обработки JSON
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.founders_temp (
        company_ogrn String,
        company_inn String,
        company_name String,
        founders_json String
    ) ENGINE = Memory
    "
    
    # Заполняем временную таблицу только записями с учредителями
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.founders_temp
    SELECT 
        ogrn,
        inn,
        full_name,
        founders
    FROM ${CLICKHOUSE_DATABASE}.companies_import
    WHERE founders != '' AND founders != '[]'
    "
    
    # Теперь импортируем учредителей из временной таблицы
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.founders (
        company_ogrn, company_inn, company_name,
        founder_type, founder_ogrn, founder_inn, founder_name,
        founder_last_name, founder_first_name, founder_middle_name,
        founder_country, share_nominal_value, share_percent,
        version_date
    )
    SELECT 
        company_ogrn,
        company_inn,
        company_name,
        -- Определяем тип учредителя по ключам в JSON
        if(JSONHas(founder_json, 'Person'), 'person',
           if(JSONHas(founder_json, 'RussianLegalEntity'), 'russian_company',
              if(JSONHas(founder_json, 'ForeignLegalEntity'), 'foreign_company',
                 if(JSONHas(founder_json, 'PublicEntity'), 'public_entity',
                    if(JSONHas(founder_json, 'MutualFund'), 'fund', 'unknown'))))) as founder_type,
        -- ОГРН учредителя (только для российских юр. лиц)
        JSONExtractString(founder_json, 'RussianLegalEntity', 'ogrn') as founder_ogrn,
        -- ИНН учредителя
        if(JSONExtractString(founder_json, 'Person', 'person', 'inn') != '',
           JSONExtractString(founder_json, 'Person', 'person', 'inn'),
           JSONExtractString(founder_json, 'RussianLegalEntity', 'inn')) as founder_inn,
        -- Имя учредителя
        if(JSONHas(founder_json, 'Person'),
           concat(JSONExtractString(founder_json, 'Person', 'person', 'last_name'), ' ',
                  JSONExtractString(founder_json, 'Person', 'person', 'first_name'),
                  if(JSONExtractString(founder_json, 'Person', 'person', 'middle_name') != '',
                     concat(' ', JSONExtractString(founder_json, 'Person', 'person', 'middle_name')), '')),
           if(JSONExtractString(founder_json, 'RussianLegalEntity', 'name') != '',
              JSONExtractString(founder_json, 'RussianLegalEntity', 'name'),
              if(JSONExtractString(founder_json, 'ForeignLegalEntity', 'name') != '',
                 JSONExtractString(founder_json, 'ForeignLegalEntity', 'name'),
                 if(JSONExtractString(founder_json, 'PublicEntity', 'name') != '',
                    JSONExtractString(founder_json, 'PublicEntity', 'name'),
                    JSONExtractString(founder_json, 'MutualFund', 'name'))))) as founder_name,
        -- Фамилия (только для физических лиц)
        JSONExtractString(founder_json, 'Person', 'person', 'last_name') as founder_last_name,
        -- Имя (только для физических лиц)
        JSONExtractString(founder_json, 'Person', 'person', 'first_name') as founder_first_name,
        -- Отчество (только для физических лиц)
        JSONExtractString(founder_json, 'Person', 'person', 'middle_name') as founder_middle_name,
        -- Страна (для иностранных юр. лиц)
        JSONExtractString(founder_json, 'ForeignLegalEntity', 'country') as founder_country,
        -- Номинальная стоимость доли
        if(JSONExtractFloat(founder_json, 'Person', 'share', 'nominal_value') != 0,
           JSONExtractFloat(founder_json, 'Person', 'share', 'nominal_value'),
           if(JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'nominal_value') != 0,
              JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'nominal_value'),
              JSONExtractFloat(founder_json, 'ForeignLegalEntity', 'share', 'nominal_value'))) as share_nominal_value,
        -- Процент доли
        if(JSONExtractFloat(founder_json, 'Person', 'share', 'percent') != 0,
           JSONExtractFloat(founder_json, 'Person', 'share', 'percent'),
           if(JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'percent') != 0,
              JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'percent'),
              JSONExtractFloat(founder_json, 'ForeignLegalEntity', 'share', 'percent'))) as share_percent,
        today() as version_date
    FROM ${CLICKHOUSE_DATABASE}.founders_temp
    ARRAY JOIN JSONExtractArrayRaw(founders_json) as founder_json
    "
    
    # Удаляем временную таблицу
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.founders_temp"
    
    local founders_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.founders")
    log_success "Импорт учредителей завершен. Всего записей: $founders_count"
}

# Функция для импорта истории изменений из временной таблицы компаний
import_history_from_companies_import() {
    log_info "Импорт истории изменений из данных ЕГРЮЛ..."
    
    # Импортируем историю из временной таблицы
    # Используем подзапрос для фильтрации перед ARRAY JOIN
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.company_history (
        entity_type, entity_id, inn,
        grn, grn_date,
        reason_code, reason_description,
        authority_code, authority_name,
        certificate_series, certificate_number, certificate_date
    )
    SELECT 
        'company' as entity_type,
        ogrn as entity_id,
        inn,
        JSONExtractString(history_json, 'grn') as grn,
        toDateOrNull(JSONExtractString(history_json, 'date')) as grn_date,
        JSONExtractString(history_json, 'reason_code') as reason_code,
        JSONExtractString(history_json, 'reason_description') as reason_description,
        JSONExtractString(history_json, 'authority_code') as authority_code,
        JSONExtractString(history_json, 'authority_name') as authority_name,
        JSONExtractString(history_json, 'certificate_series') as certificate_series,
        JSONExtractString(history_json, 'certificate_number') as certificate_number,
        toDateOrNull(JSONExtractString(history_json, 'certificate_date')) as certificate_date
    FROM (
        SELECT ogrn, inn, history
        FROM ${CLICKHOUSE_DATABASE}.companies_import
        WHERE length(history) > 2
    )
    ARRAY JOIN JSONExtractArrayRaw(history) as history_json
    SETTINGS max_partitions_per_insert_block = 1000
    "
    
    local history_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.company_history WHERE entity_type = 'company'")
    log_success "Импорт истории ЕГРЮЛ завершен. Всего записей: $history_count"
}

# Функция для импорта истории изменений из временной таблицы предпринимателей
import_history_from_entrepreneurs_import() {
    log_info "Импорт истории изменений из данных ЕГРИП..."
    
    # Импортируем историю из временной таблицы
    # Используем подзапрос для фильтрации перед ARRAY JOIN
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.company_history (
        entity_type, entity_id, inn,
        grn, grn_date,
        reason_code, reason_description,
        authority_code, authority_name,
        certificate_series, certificate_number, certificate_date
    )
    SELECT 
        'entrepreneur' as entity_type,
        ogrnip as entity_id,
        inn,
        JSONExtractString(history_json, 'grn') as grn,
        toDateOrNull(JSONExtractString(history_json, 'date')) as grn_date,
        JSONExtractString(history_json, 'reason_code') as reason_code,
        JSONExtractString(history_json, 'reason_description') as reason_description,
        JSONExtractString(history_json, 'authority_code') as authority_code,
        JSONExtractString(history_json, 'authority_name') as authority_name,
        JSONExtractString(history_json, 'certificate_series') as certificate_series,
        JSONExtractString(history_json, 'certificate_number') as certificate_number,
        toDateOrNull(JSONExtractString(history_json, 'certificate_date')) as certificate_date
    FROM (
        SELECT ogrnip, inn, history
        FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import
        WHERE length(history) > 2
    )
    ARRAY JOIN JSONExtractArrayRaw(history) as history_json
    SETTINGS max_partitions_per_insert_block = 1000
    "
    
    local history_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.company_history WHERE entity_type = 'entrepreneur'")
    log_success "Импорт истории ЕГРИП завершен. Всего записей: $history_count"
}

# Функция для создания промежуточной таблицы и трансформации данных
import_egrul_with_transform() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        log_warning "Файл ЕГРЮЛ не найден: $file"
        return 1
    fi
    
    local file_size=$(du -h "$file" | cut -f1)
    log_info "Импорт ЕГРЮЛ из $file ($file_size)..."
    
    # Создаем временную таблицу для импорта
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.companies_import"
    
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.companies_import (
        ogrn String,
        ogrn_date Nullable(String),
        inn String,
        kpp Nullable(String),
        full_name String,
        short_name Nullable(String),
        status Nullable(String),
        status_code Nullable(String),
        registration_date Nullable(String),
        termination_date Nullable(String),
        -- Все поля адреса
        postal_code Nullable(String),
        region_code Nullable(String),
        region Nullable(String),
        district Nullable(String),
        city Nullable(String),
        locality Nullable(String),
        street Nullable(String),
        house Nullable(String),
        building Nullable(String),
        flat Nullable(String),
        full_address Nullable(String),
        fias_id Nullable(String),
        kladr_code Nullable(String),
        capital_amount Nullable(Float64),
        capital_currency Nullable(String),
        head_name Nullable(String),
        head_inn Nullable(String),
        head_middle_name Nullable(String),
        head_position Nullable(String),
        main_activity_code Nullable(String),
        main_activity_name Nullable(String),
        additional_activities Nullable(String),
        email Nullable(String),
        founders_count Nullable(Int32),
        founders String DEFAULT '',
        history String DEFAULT '',
        extract_date Nullable(String)
    ) ENGINE = Memory
    "
    
    # Загружаем Parquet во временную таблицу
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/?query=INSERT%20INTO%20${CLICKHOUSE_DATABASE}.companies_import%20FORMAT%20Parquet" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "@$file"
    
    # Получаем количество загруженных записей
    local imported_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.companies_import")
    log_info "Загружено во временную таблицу: $imported_count записей"

    # region agent log: счетчик импортированных записей ЕГРЮЛ (H1)
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H1","location":"import-data.sh:companies_import","message":"companies_import_count","data":{"imported_count":'"$imported_count"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРН в временной таблице (H3)
    local uniq_ogrn_import=$(clickhouse_query "SELECT uniqExact(ogrn) FROM ${CLICKHOUSE_DATABASE}.companies_import")
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H3","location":"import-data.sh:companies_import","message":"companies_import_uniq_ogrn","data":{"uniq_ogrn":'"$uniq_ogrn_import"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log
    
    # Для MVP не удаляем старые записи здесь, а просто вставляем новые.
    # Дедупликация по extract_date будет реализована позже через отдельный процесс.
    # Сейчас важно убедиться, что весь объём данных корректно загружается в таблицу companies.
    
    # Трансформируем и вставляем в основную таблицу
    # Используем логику обновления: заменяем запись только если новая extract_date >= существующей
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.companies (
        ogrn, ogrn_date, inn, kpp, full_name, short_name, status, status_code,
        registration_date, termination_date, 
        postal_code, region_code, region, district, city, locality, street, house, building, flat, full_address, fias_id, kladr_code,
        capital_amount, capital_currency,
        head_last_name, head_first_name, head_middle_name, head_inn, head_position,
        okved_main_code, okved_main_name,
        okved_additional, okved_additional_names, additional_activities,
        email, founders_count, extract_date,
        version_date
    )
    SELECT
        ogrn,
        toDateOrNull(ogrn_date),
        inn,
        kpp,
        full_name,
        short_name,
        coalesce(status, 'unknown'),
        status_code,
        toDateOrNull(registration_date),
        toDateOrNull(termination_date),
        -- Все поля адреса
        postal_code,
        region_code,
        region,
        district,
        city,
        locality,
        street,
        house,
        building,
        flat,
        full_address,
        fias_id,
        kladr_code,
        capital_amount,
        coalesce(capital_currency, 'RUB'),
        -- Разбиваем head_name на части (исправленная версия)
        arrayElement(splitByChar(' ', coalesce(head_name, '')), 1) AS head_last_name,
        if(length(splitByChar(' ', coalesce(head_name, ''))) > 1, 
           arrayElement(splitByChar(' ', coalesce(head_name, '')), 2), NULL) AS head_first_name,
        head_middle_name,
        head_inn,
        head_position,
        main_activity_code,
        main_activity_name,
        [] AS okved_additional,
        [] AS okved_additional_names,
        additional_activities,
        email,
        founders_count,
        -- Используем extract_date, если она не NULL, иначе минимальную дату для правильной работы ReplacingMergeTree
        coalesce(toDateOrNull(extract_date), toDate('1970-01-01')) AS extract_date,
        today()
    FROM ${CLICKHOUSE_DATABASE}.companies_import
    "
    
    # Получаем количество записей в основной таблице
    local total_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.companies")
    log_success "Импорт ЕГРЮЛ завершен. Всего записей в таблице: $total_count"

    # region agent log: итоговое количество записей ЕГРЮЛ (H1)
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H1","location":"import-data.sh:companies_total","message":"companies_total_count","data":{"total_count":'"$total_count"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРН в основной таблице (H3)
    local uniq_ogrn_total=$(clickhouse_query "SELECT uniqExact(ogrn) FROM ${CLICKHOUSE_DATABASE}.companies")
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H3","location":"import-data.sh:companies_total","message":"companies_total_uniq_ogrn","data":{"uniq_ogrn":'"$uniq_ogrn_total"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log
    
    # Принудительное слияние для удаления дублей (опционально, может быть медленно для больших таблиц)
    if [ "${OPTIMIZE_AFTER_IMPORT:-false}" = "true" ]; then
        log_info "Выполнение принудительного слияния для удаления дублей..."
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.companies FINAL"
    fi
    
    # Импорт учредителей
    import_founders_from_companies_import
    
    # Импорт истории изменений
    import_history_from_companies_import
    
    # Удаляем временную таблицу
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.companies_import"
}

# Функция для импорта ЕГРИП
import_egrip_with_transform() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        log_warning "Файл ЕГРИП не найден: $file"
        return 1
    fi
    
    local file_size=$(du -h "$file" | cut -f1)
    log_info "Импорт ЕГРИП из $file ($file_size)..."
    
    # Создаем временную таблицу для импорта
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.entrepreneurs_import"
    
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs_import (
        ogrnip String,
        ogrnip_date Nullable(String),
        inn String,
        last_name String,
        first_name String,
        middle_name Nullable(String),
        full_name String,
        gender Nullable(String),
        citizenship Nullable(String),
        status Nullable(String),
        status_code Nullable(String),
        registration_date Nullable(String),
        termination_date Nullable(String),
        -- Все поля адреса
        postal_code Nullable(String),
        region_code Nullable(String),
        region Nullable(String),
        district Nullable(String),
        city Nullable(String),
        locality Nullable(String),
        street Nullable(String),
        house Nullable(String),
        building Nullable(String),
        flat Nullable(String),
        full_address Nullable(String),
        fias_id Nullable(String),
        kladr_code Nullable(String),
        main_activity_code Nullable(String),
        main_activity_name Nullable(String),
        additional_activities Nullable(String),
        email Nullable(String),
        history String DEFAULT '',
        extract_date Nullable(String)
    ) ENGINE = Memory
    "
    
    # Загружаем Parquet во временную таблицу
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/?query=INSERT%20INTO%20${CLICKHOUSE_DATABASE}.entrepreneurs_import%20FORMAT%20Parquet" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "@$file"
    
    # Получаем количество загруженных записей
    local imported_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import")
    log_info "Загружено во временную таблицу: $imported_count записей"

    # region agent log: счетчик импортированных записей ЕГРИП (H2)
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H2","location":"import-data.sh:egrip_import","message":"entrepreneurs_import_count","data":{"imported_count":'"$imported_count"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРНИП в временной таблице (H4)
    local uniq_ogrnip_import=$(clickhouse_query "SELECT uniqExact(ogrnip) FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import")
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H4","location":"import-data.sh:egrip_import","message":"entrepreneurs_import_uniq_ogrnip","data":{"uniq_ogrnip":'"$uniq_ogrnip_import"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: схема временной таблицы ЕГРИП (H2)
    clickhouse_query "DESCRIBE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs_import FORMAT JSONEachRow" >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: схема основной таблицы ЕГРИП (H2)
    clickhouse_query "DESCRIBE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs FORMAT JSONEachRow" >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # Для MVP не удаляем старые записи здесь, а просто вставляем новые.
    # Дедупликация по extract_date будет реализована позже через отдельный процесс.
    # Сейчас важно убедиться, что весь объём данных корректно загружается в таблицу entrepreneurs.
    
    # Трансформируем и вставляем в основную таблицу
    # Заполняем как базовые, так и основные аналитические поля (регион, ОКВЭД, email).
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.entrepreneurs (
        ogrnip, ogrnip_date, inn,
        last_name, first_name, middle_name, gender,
        citizenship_type,
        status, status_code, registration_date, termination_date,
        postal_code, region_code, region, district, city, locality, full_address, fias_id, kladr_code,
        okved_main_code, okved_main_name,
        okved_additional, okved_additional_names, additional_activities,
        email, extract_date,
        version_date
    )
    SELECT
        ogrnip,
        toDateOrNull(ogrnip_date),
        inn,
        last_name,
        first_name,
        middle_name,
        coalesce(gender, '') AS gender,
        coalesce(citizenship, '') AS citizenship_type,
        coalesce(status, 'unknown') AS status,
        status_code,
        toDateOrNull(registration_date),
        toDateOrNull(termination_date),
        -- Все поля адреса
        postal_code,
        region_code,
        region,
        district,
        city,
        locality,
        full_address,
        fias_id,
        kladr_code,
        main_activity_code,
        main_activity_name,
        [] AS okved_additional,
        [] AS okved_additional_names,
        additional_activities,
        email,
        coalesce(toDateOrNull(extract_date), toDate('1970-01-01')) AS extract_date,
        today()
    FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import
    "
    
    # Получаем количество записей в основной таблице
    local total_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.entrepreneurs")
    log_success "Импорт ЕГРИП завершен. Всего записей в таблице: $total_count"

    # region agent log: итоговое количество записей ЕГРИП (H2)
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H2","location":"import-data.sh:egrip_total","message":"entrepreneurs_total_count","data":{"total_count":'"$total_count"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРНИП в основной таблице (H4)
    local uniq_ogrnip_total=$(clickhouse_query "SELECT uniqExact(ogrnip) FROM ${CLICKHOUSE_DATABASE}.entrepreneurs")
    echo '{"sessionId":"debug-session","runId":"pre-fix","hypothesisId":"H4","location":"import-data.sh:egrip_total","message":"entrepreneurs_total_uniq_ogrnip","data":{"uniq_ogrnip":'"$uniq_ogrnip_total"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log
    
    # Принудительное слияние для удаления дублей (опционально, может быть медленно для больших таблиц)
    if [ "${OPTIMIZE_AFTER_IMPORT:-false}" = "true" ]; then
        log_info "Выполнение принудительного слияния для удаления дублей..."
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs FINAL"
    fi
    
    # Импорт истории изменений
    import_history_from_entrepreneurs_import
    
    # Удаляем временную таблицу
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.entrepreneurs_import"
}

# Проверка подключения к ClickHouse
check_connection() {
    log_info "Проверка подключения к ClickHouse..."
    
    local result=$(clickhouse_query "SELECT 1")
    if [ "$result" == "1" ]; then
        log_success "Подключение к ClickHouse успешно"
        return 0
    else
        log_error "Не удалось подключиться к ClickHouse"
        return 1
    fi
}

# Показать статистику
show_stats() {
    log_info "Статистика базы данных:"
    echo ""
    echo "Компании (ЕГРЮЛ):"
    clickhouse_query "SELECT 
        count() as total,
        countIf(status = 'active') as active,
        countIf(status = 'liquidated') as liquidated
    FROM ${CLICKHOUSE_DATABASE}.companies FORMAT Pretty"
    echo ""
    echo "ИП (ЕГРИП):"
    clickhouse_query "SELECT 
        count() as total,
        countIf(status = 'active') as active,
        countIf(status = 'liquidated') as liquidated
    FROM ${CLICKHOUSE_DATABASE}.entrepreneurs FORMAT Pretty"
    echo ""
    echo "История изменений:"
    clickhouse_query "SELECT 
        entity_type,
        count() as total
    FROM ${CLICKHOUSE_DATABASE}.company_history
    GROUP BY entity_type
    FORMAT Pretty"
}

# Главная функция
main() {
    echo ""
    echo "======================================"
    echo "  ЕГРЮЛ/ЕГРИП Data Import Tool"
    echo "======================================"
    echo ""
    
    # Проверяем подключение
    check_connection || exit 1
    
    # Очищаем таблицу истории перед импортом (для MVP - полная перезагрузка)
    log_info "Очистка таблицы истории изменений..."
    clickhouse_query "TRUNCATE TABLE ${CLICKHOUSE_DATABASE}.company_history"
    
    # Импорт ЕГРЮЛ
    local egrul_file="${DATA_DIR}/egrul_egrul.parquet"
    if [ -f "$egrul_file" ]; then
        import_egrul_with_transform "$egrul_file"
    else
        log_warning "Файл ЕГРЮЛ не найден: $egrul_file"
    fi
    
    echo ""
    
    # Импорт ЕГРИП
    local egrip_file="${DATA_DIR}/egrip_egrip.parquet"
    if [ -f "$egrip_file" ]; then
        import_egrip_with_transform "$egrip_file"
    else
        log_warning "Файл ЕГРИП не найден: $egrip_file"
    fi
    
    echo ""
    
    # Показываем статистику
    show_stats
    
    echo ""
    log_success "Импорт данных завершен!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать справку"
        echo "  --stats        Показать только статистику"
        echo "  --egrul FILE   Импортировать только ЕГРЮЛ из указанного файла"
        echo "  --egrip FILE   Импортировать только ЕГРИП из указанного файла"
        echo ""
        echo "Переменные окружения:"
        echo "  CLICKHOUSE_HOST      Хост ClickHouse (default: localhost)"
        echo "  CLICKHOUSE_PORT      HTTP порт (default: 8123)"
        echo "  CLICKHOUSE_USER      Пользователь (default: admin)"
        echo "  CLICKHOUSE_PASSWORD  Пароль (default: admin)"
        echo "  CLICKHOUSE_DATABASE  База данных (default: egrul)"
        echo "  DATA_DIR             Директория с Parquet файлами (default: ./output)"
        echo "  OPTIMIZE_AFTER_IMPORT Выполнить OPTIMIZE после импорта (default: false)"
        exit 0
        ;;
    --stats)
        check_connection && show_stats
        ;;
    --egrul)
        check_connection && import_egrul_with_transform "$2"
        ;;
    --egrip)
        check_connection && import_egrip_with_transform "$2"
        ;;
    *)
        main
        ;;
esac

