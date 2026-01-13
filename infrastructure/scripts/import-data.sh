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
CLICKHOUSE_USER="${CLICKHOUSE_USER:-egrul_import}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-123}"
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

# Функция для импорта учредителей из временной таблицы компаний (устаревшая, оставлена для совместимости)
import_founders_from_companies_import() {
    log_warning "Функция import_founders_from_companies_import устарела - учредители теперь импортируются в рамках обработки отдельных файлов"
}

# Функция для импорта истории изменений из временной таблицы компаний
import_history_from_companies_import() {
    log_warning "Функция import_history_from_companies_import устарела - история теперь импортируется в рамках обработки отдельных файлов"
}

# Функция для импорта истории изменений из временной таблицы предпринимателей
import_history_from_entrepreneurs_import() {
    log_warning "Функция import_history_from_entrepreneurs_import устарела - история теперь импортируется в рамках обработки отдельных файлов"
}

# Функция для импорта одного файла ЕГРЮЛ с полной обработкой
import_single_egrul_file() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        log_warning "Файл не найден: $file"
        return 1
    fi
    
    local file_size=$(du -h "$file" | cut -f1)
    local file_basename=$(basename "$file")
    log_info "Обработка файла $file_basename ($file_size)..."
    
    # Создаем временную таблицу для одного файла
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.companies_import_single"
    
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.companies_import_single (
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
        licenses String DEFAULT '',
        branches String DEFAULT '',
        extract_date Nullable(String)
    ) ENGINE = Memory
    "
    
    # Загружаем файл во временную таблицу
    log_info "  → Загрузка данных..."
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/?query=INSERT%20INTO%20${CLICKHOUSE_DATABASE}.companies_import_single%20FORMAT%20Parquet" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "@$file"
    
    if [ $? -ne 0 ]; then
        log_error "Ошибка загрузки файла $file"
        clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.companies_import_single"
        return 1
    fi
    
    local imported_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.companies_import_single")
    log_info "  → Загружено записей: $imported_count"
    
    # Импорт основных данных компаний
    log_info "  → Импорт основных данных..."
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
        -- Разбиваем head_name на части
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
        coalesce(toDateOrNull(extract_date), toDate('1970-01-01')) AS extract_date,
        today()
    FROM ${CLICKHOUSE_DATABASE}.companies_import_single
    "
    
    # Импорт учредителей из этого файла
    log_info "  → Импорт учредителей..."
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.founders_temp_single (
        company_ogrn String,
        company_inn String,
        company_name String,
        founders_json String
    ) ENGINE = Memory
    "
    
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.founders_temp_single
    SELECT 
        ogrn,
        inn,
        full_name,
        founders
    FROM ${CLICKHOUSE_DATABASE}.companies_import_single
    WHERE founders != '' AND founders != '[]'
    "
    
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
        if(JSONHas(founder_json, 'Person'), 'person',
           if(JSONHas(founder_json, 'RussianLegalEntity'), 'russian_company',
              if(JSONHas(founder_json, 'ForeignLegalEntity'), 'foreign_company',
                 if(JSONHas(founder_json, 'PublicEntity'), 'public_entity',
                    if(JSONHas(founder_json, 'MutualFund'), 'fund', 'unknown'))))) as founder_type,
        JSONExtractString(founder_json, 'RussianLegalEntity', 'ogrn') as founder_ogrn,
        if(JSONExtractString(founder_json, 'Person', 'person', 'inn') != '',
           JSONExtractString(founder_json, 'Person', 'person', 'inn'),
           JSONExtractString(founder_json, 'RussianLegalEntity', 'inn')) as founder_inn,
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
        JSONExtractString(founder_json, 'Person', 'person', 'last_name') as founder_last_name,
        JSONExtractString(founder_json, 'Person', 'person', 'first_name') as founder_first_name,
        JSONExtractString(founder_json, 'Person', 'person', 'middle_name') as founder_middle_name,
        JSONExtractString(founder_json, 'ForeignLegalEntity', 'country') as founder_country,
        if(JSONExtractFloat(founder_json, 'Person', 'share', 'nominal_value') != 0,
           JSONExtractFloat(founder_json, 'Person', 'share', 'nominal_value'),
           if(JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'nominal_value') != 0,
              JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'nominal_value'),
              JSONExtractFloat(founder_json, 'ForeignLegalEntity', 'share', 'nominal_value'))) as share_nominal_value,
        if(JSONExtractFloat(founder_json, 'Person', 'share', 'percent') != 0,
           JSONExtractFloat(founder_json, 'Person', 'share', 'percent'),
           if(JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'percent') != 0,
              JSONExtractFloat(founder_json, 'RussianLegalEntity', 'share', 'percent'),
              JSONExtractFloat(founder_json, 'ForeignLegalEntity', 'share', 'percent'))) as share_percent,
        today() as version_date
    FROM ${CLICKHOUSE_DATABASE}.founders_temp_single
    ARRAY JOIN JSONExtractArrayRaw(founders_json) as founder_json
    "
    
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.founders_temp_single"

    # Импорт лицензий из этого файла
    log_info "  → Импорт лицензий..."
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.licenses (
        entity_type, entity_ogrn, entity_inn, entity_name,
        license_number, license_series, activity,
        start_date, end_date, authority, status,
        version_date
    )
    SELECT
        'company' as entity_type,
        ogrn as entity_ogrn,
        inn as entity_inn,
        full_name as entity_name,
        JSONExtractString(license_json, 'number') as license_number,
        JSONExtractString(license_json, 'series') as license_series,
        JSONExtractString(license_json, 'activity') as activity,
        toDateOrNull(JSONExtractString(license_json, 'start_date')) as start_date,
        toDateOrNull(JSONExtractString(license_json, 'end_date')) as end_date,
        JSONExtractString(license_json, 'authority') as authority,
        coalesce(JSONExtractString(license_json, 'status'), 'active') as status,
        today() as version_date
    FROM ${CLICKHOUSE_DATABASE}.companies_import_single
    ARRAY JOIN JSONExtractArrayRaw(licenses) as license_json
    WHERE licenses != '' AND licenses != '[]'
    "

    # Импорт филиалов из этого файла
    log_info "  → Импорт филиалов..."
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.branches (
        company_ogrn, company_inn, company_name,
        branch_type, branch_name, branch_kpp,
        postal_code, region_code, region, city, full_address,
        grn, grn_date,
        version_date
    )
    SELECT
        ogrn as company_ogrn,
        inn as company_inn,
        full_name as company_name,
        if(JSONExtractString(branch_json, 'branch_type') = 'Representative', 'representative', 'branch') as branch_type,
        JSONExtractString(branch_json, 'name') as branch_name,
        JSONExtractString(branch_json, 'kpp') as branch_kpp,
        JSONExtractString(branch_json, 'address', 'postal_code') as postal_code,
        JSONExtractString(branch_json, 'address', 'region_code') as region_code,
        JSONExtractString(branch_json, 'address', 'region') as region,
        JSONExtractString(branch_json, 'address', 'city') as city,
        JSONExtractString(branch_json, 'address', 'full_address') as full_address,
        JSONExtractString(branch_json, 'grn') as grn,
        toDateOrNull(JSONExtractString(branch_json, 'grn_date')) as grn_date,
        today() as version_date
    FROM ${CLICKHOUSE_DATABASE}.companies_import_single
    ARRAY JOIN JSONExtractArrayRaw(branches) as branch_json
    WHERE branches != '' AND branches != '[]'
    "

    # Импорт истории изменений из этого файла (ЕГРЮЛ)
    log_info "  → Импорт истории изменений..."
    local buckets="${HISTORY_BUCKETS:-100}"
    local memory_limit="${HISTORY_MAX_MEMORY:-2000000000}" # 2 ГБ по умолчанию
    
    log_info "  → Настройки памяти: $buckets батчей, лимит $memory_limit байт ($(($memory_limit / 1024 / 1024)) МБ)"
    
    for ((bucket=0; bucket<buckets; bucket++)); do
        clickhouse_query "
        INSERT INTO ${CLICKHOUSE_DATABASE}.company_history (
            entity_type, entity_id, inn,
            grn, grn_date,
            reason_code, reason_description,
            authority_code, authority_name,
            certificate_series, certificate_number, certificate_date,
            source_files, extract_date, file_hash,
            created_at, updated_at
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
            toDateOrNull(JSONExtractString(history_json, 'certificate_date')) as certificate_date,
            ['parquet_import'] as source_files,
            coalesce(toDateOrNull(extract_date), today()) as extract_date,
            'parquet_' || toString(cityHash64(ogrn || JSONExtractString(history_json, 'grn'))) as file_hash,
            now64(3) as created_at,
            now64(3) as updated_at
        FROM (
            SELECT ogrn, inn, history, extract_date
            FROM ${CLICKHOUSE_DATABASE}.companies_import_single
            WHERE length(history) > 2
              AND cityHash64(ogrn) % $buckets = $bucket
        )
        ARRAY JOIN JSONExtractArrayRaw(history) as history_json
        SETTINGS max_memory_usage=$memory_limit, max_partitions_per_insert_block = 1000
        " 2>/dev/null || true
    done
    
    # Очищаем временную таблицу
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.companies_import_single"
    
    log_success "  → Файл $file_basename обработан успешно"
}

# Функция для создания промежуточной таблицы и трансформации данных
import_egrul_with_transform() {
    local file_pattern="$1"
    
    # Находим все файлы ЕГРЮЛ (включая разбитые на части)
    local files=($(find "$DATA_DIR" -name "*egrul*.parquet" 2>/dev/null | sort))
    
    if [ ${#files[@]} -eq 0 ]; then
        log_warning "Файлы ЕГРЮЛ не найдены в $DATA_DIR"
        return 1
    fi
    
    log_info "Найдено ${#files[@]} файлов ЕГРЮЛ для последовательной обработки"
    
    # Обрабатываем каждый файл по отдельности
    local processed_files=0
    for file in "${files[@]}"; do
        import_single_egrul_file "$file"
        if [ $? -eq 0 ]; then
            ((processed_files++))
        fi
    done
    
    # Получаем итоговую статистику
    local total_companies=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.companies")
    local total_founders=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.founders")
    local total_history=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.company_history WHERE entity_type = 'company'")
    
    log_success "Импорт ЕГРЮЛ завершен:"
    log_success "  → Обработано файлов: $processed_files/${#files[@]}"
    log_success "  → Компаний: $total_companies"
    log_success "  → Учредителей: $total_founders" 
    log_success "  → История изменений: $total_history"

    # region agent log: итоговое количество записей ЕГРЮЛ (H1)
    echo '{"sessionId":"debug-session","runId":"sequential-fix","hypothesisId":"H1","location":"import-data.sh:companies_total","message":"companies_total_count","data":{"total_count":'"$total_companies"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРН в основной таблице (H3)
    local uniq_ogrn_total=$(clickhouse_query "SELECT uniqExact(ogrn) FROM ${CLICKHOUSE_DATABASE}.companies")
    echo '{"sessionId":"debug-session","runId":"sequential-fix","hypothesisId":"H3","location":"import-data.sh:companies_total","message":"companies_total_uniq_ogrn","data":{"uniq_ogrn":'"$uniq_ogrn_total"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log
    
    # Принудительное слияние для удаления дублей (опционально)
    if [ "${OPTIMIZE_AFTER_IMPORT:-false}" = "true" ]; then
        log_info "Выполнение принудительного слияния для удаления дублей..."
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.companies FINAL"
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.founders FINAL"
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.company_history FINAL"
    fi
}

# Функция для импорта одного файла ЕГРИП с полной обработкой
import_single_egrip_file() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        log_warning "Файл не найден: $file"
        return 1
    fi
    
    local file_size=$(du -h "$file" | cut -f1)
    local file_basename=$(basename "$file")
    log_info "Обработка файла $file_basename ($file_size)..."
    
    # Создаем временную таблицу для одного файла
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single"
    
    clickhouse_query "
    CREATE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single (
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
    
    # Загружаем файл во временную таблицу
    log_info "  → Загрузка данных..."
    curl -s "http://${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/?query=INSERT%20INTO%20${CLICKHOUSE_DATABASE}.entrepreneurs_import_single%20FORMAT%20Parquet" \
        --user "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
        --data-binary "@$file"
    
    if [ $? -ne 0 ]; then
        log_error "Ошибка загрузки файла $file"
        clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single"
        return 1
    fi
    
    local imported_count=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single")
    log_info "  → Загружено записей: $imported_count"
    
    # Импорт основных данных предпринимателей
    log_info "  → Импорт основных данных..."
    clickhouse_query "
    INSERT INTO ${CLICKHOUSE_DATABASE}.entrepreneurs (
        ogrnip, ogrnip_date, inn,
        last_name, first_name, middle_name, gender,
        citizenship_type,
        status, status_code, registration_date, termination_date,
        postal_code, region_code, region, district, city, locality, street, house, building, flat, full_address, fias_id, kladr_code,
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
        street,
        house,
        building,
        flat,
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
    FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single
    "
    
    # Импорт истории изменений из этого файла (ЕГРИП)
    log_info "  → Импорт истории изменений..."
    local buckets="${HISTORY_BUCKETS:-100}"
    local memory_limit="${HISTORY_MAX_MEMORY:-2000000000}" # 2 ГБ по умолчанию
    
    log_info "  → Настройки памяти: $buckets батчей, лимит $memory_limit байт ($(($memory_limit / 1024 / 1024)) МБ)"
    
    for ((bucket=0; bucket<buckets; bucket++)); do
        clickhouse_query "
        INSERT INTO ${CLICKHOUSE_DATABASE}.company_history (
            entity_type, entity_id, inn,
            grn, grn_date,
            reason_code, reason_description,
            authority_code, authority_name,
            certificate_series, certificate_number, certificate_date,
            source_files, extract_date, file_hash,
            created_at, updated_at
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
            toDateOrNull(JSONExtractString(history_json, 'certificate_date')) as certificate_date,
            ['parquet_import'] as source_files,
            coalesce(toDateOrNull(extract_date), today()) as extract_date,
            'parquet_' || toString(cityHash64(ogrnip || JSONExtractString(history_json, 'grn'))) as file_hash,
            now64(3) as created_at,
            now64(3) as updated_at
        FROM (
            SELECT ogrnip, inn, history, extract_date
            FROM ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single
            WHERE length(history) > 2
              AND cityHash64(ogrnip) % $buckets = $bucket
        )
        ARRAY JOIN JSONExtractArrayRaw(history) as history_json
        SETTINGS max_memory_usage=$memory_limit, max_partitions_per_insert_block = 1000
        " 2>/dev/null || true
    done
    
    # Очищаем временную таблицу
    clickhouse_query "DROP TABLE IF EXISTS ${CLICKHOUSE_DATABASE}.entrepreneurs_import_single"
    
    log_success "  → Файл $file_basename обработан успешно"
}

# Функция для импорта ЕГРИП
import_egrip_with_transform() {
    local file_pattern="$1"
    
    # Находим все файлы ЕГРИП (включая разбитые на части)
    local files=($(find "$DATA_DIR" -name "*egrip*.parquet" 2>/dev/null | sort))
    
    if [ ${#files[@]} -eq 0 ]; then
        log_warning "Файлы ЕГРИП не найдены в $DATA_DIR"
        return 1
    fi
    
    log_info "Найдено ${#files[@]} файлов ЕГРИП для последовательной обработки"
    
    # Обрабатываем каждый файл по отдельности
    local processed_files=0
    for file in "${files[@]}"; do
        import_single_egrip_file "$file"
        if [ $? -eq 0 ]; then
            ((processed_files++))
        fi
    done
    
    # Получаем итоговую статистику
    local total_entrepreneurs=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.entrepreneurs")
    local total_history=$(clickhouse_query "SELECT count() FROM ${CLICKHOUSE_DATABASE}.company_history WHERE entity_type = 'entrepreneur'")
    
    log_success "Импорт ЕГРИП завершен:"
    log_success "  → Обработано файлов: $processed_files/${#files[@]}"
    log_success "  → Предпринимателей: $total_entrepreneurs"
    log_success "  → История изменений: $total_history"

    # region agent log: итоговое количество записей ЕГРИП (H2)
    echo '{"sessionId":"debug-session","runId":"sequential-fix","hypothesisId":"H2","location":"import-data.sh:egrip_total","message":"entrepreneurs_total_count","data":{"total_count":'"$total_entrepreneurs"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log

    # region agent log: количество уникальных ОГРНИП в основной таблице (H4)
    local uniq_ogrnip_total=$(clickhouse_query "SELECT uniqExact(ogrnip) FROM ${CLICKHOUSE_DATABASE}.entrepreneurs")
    echo '{"sessionId":"debug-session","runId":"sequential-fix","hypothesisId":"H4","location":"import-data.sh:egrip_total","message":"entrepreneurs_total_uniq_ogrnip","data":{"uniq_ogrnip":'"$uniq_ogrnip_total"'},"timestamp":'$(date +%s%3N)'}' >> "$DEBUG_LOG_PATH" || true
    # endregion agent log
    
    # Принудительное слияние для удаления дублей (опционально)
    if [ "${OPTIMIZE_AFTER_IMPORT:-false}" = "true" ]; then
        log_info "Выполнение принудительного слияния для удаления дублей..."
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.entrepreneurs FINAL"
        clickhouse_query "OPTIMIZE TABLE ${CLICKHOUSE_DATABASE}.company_history FINAL"
    fi
}

# Проверка подключения к ClickHouse
check_connection() {
    log_info "Проверка подключения к ClickHouse..."
    
    local result=$(clickhouse_query "SELECT 1")
    if [ "$result" == "1" ]; then
        log_success "Подключение к ClickHouse успешно"
        
        # Проверяем текущие лимиты памяти
        log_info "Проверка лимитов памяти ClickHouse..."
        local max_memory_usage=$(clickhouse_query "SELECT value FROM system.settings WHERE name = 'max_memory_usage'" 2>/dev/null || echo "0")
        
        log_info "Текущий лимит памяти для запросов: $max_memory_usage байт"
        if [ -n "$max_memory_usage" ] && [ "$max_memory_usage" -gt 0 ]; then
            log_info "  → $(($max_memory_usage / 1024 / 1024)) МБ"
        fi
        
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
    echo "Лицензии:"
    clickhouse_query "SELECT
        count() as total,
        countIf(status = 'active') as active,
        countIf(status = 'expired') as expired
    FROM ${CLICKHOUSE_DATABASE}.licenses FORMAT Pretty"
    echo ""
    echo "Филиалы и представительства:"
    clickhouse_query "SELECT
        branch_type,
        count() as total
    FROM ${CLICKHOUSE_DATABASE}.branches
    GROUP BY branch_type
    FORMAT Pretty"
    echo ""
    echo "История изменений (с дедупликацией):"
    clickhouse_query "SELECT 
        entity_type,
        count() as total_records,
        count(DISTINCT grn) as unique_grns,
        count() - count(DISTINCT grn) as potential_duplicates,
        round((count() - count(DISTINCT grn)) / count() * 100, 2) as dedup_ratio_percent
    FROM ${CLICKHOUSE_DATABASE}.company_history
    GROUP BY entity_type
    FORMAT Pretty"
    echo ""
    echo "Дедуплицированная история (через VIEW):"
    clickhouse_query "SELECT 
        entity_type,
        count() as unique_records
    FROM ${CLICKHOUSE_DATABASE}.company_history_view
    GROUP BY entity_type
    FORMAT Pretty"
    echo ""
    echo "Статистика по источникам данных:"
    clickhouse_query "SELECT 
        arrayJoin(source_files) as source_file,
        count() as records_count,
        count(DISTINCT entity_id) as unique_entities
    FROM ${CLICKHOUSE_DATABASE}.company_history_view
    GROUP BY source_file
    ORDER BY records_count DESC
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
    
    # Устанавливаем более высокие лимиты памяти для сессии
    local session_memory_limit="${HISTORY_MAX_MEMORY:-2000000000}"
    log_info "Установка лимитов памяти для сессии: $(($session_memory_limit / 1024 / 1024)) МБ"
    
    clickhouse_query "SET max_memory_usage = $session_memory_limit" || log_warning "Не удалось установить max_memory_usage"
    
    # Очищаем таблицу истории перед импортом (для MVP - полная перезагрузка)
    log_info "Очистка таблицы истории изменений..."
    clickhouse_query "TRUNCATE TABLE ${CLICKHOUSE_DATABASE}.company_history"
    
    # Импорт ЕГРЮЛ
    local egrul_files=($(find "$DATA_DIR" -name "*egrul*.parquet" 2>/dev/null))
    if [ ${#egrul_files[@]} -gt 0 ]; then
        import_egrul_with_transform
    else
        log_warning "Файлы ЕГРЮЛ не найдены в $DATA_DIR"
    fi
    
    echo ""
    
    # Импорт ЕГРИП
    local egrip_files=($(find "$DATA_DIR" -name "*egrip*.parquet" 2>/dev/null))
    if [ ${#egrip_files[@]} -gt 0 ]; then
        import_egrip_with_transform
    else
        log_warning "Файлы ЕГРИП не найдены в $DATA_DIR"
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
        echo "  CLICKHOUSE_USER      Пользователь (default: egrul_import)"
        echo "  CLICKHOUSE_PASSWORD  Пароль (default: 123)"
        echo "  CLICKHOUSE_DATABASE  База данных (default: egrul)"
        echo "  DATA_DIR             Директория с Parquet файлами (default: ./output)"
        echo "  OPTIMIZE_AFTER_IMPORT Выполнить OPTIMIZE после импорта (default: false)"
        echo "  HISTORY_BUCKETS      Количество батчей для истории (default: 100)"
        echo "  HISTORY_MAX_MEMORY   Лимит памяти для батчей истории в байтах (default: 2000000000)"
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

