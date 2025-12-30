-- ============================================================================
-- МИГРАЦИЯ: Оптимизированная схема истории с автоматической дедупликацией
-- Version: 006_optimized_history_schema
-- Description: Создание таблицы истории с автоматической дедупликацией при импорте
-- ============================================================================

-- Удаляем старую таблицу если существует
DROP TABLE IF EXISTS egrul.company_history;

-- Создаем оптимизированную таблицу истории с автоматической дедупликацией
CREATE TABLE egrul.company_history
(
    -- === Идентификаторы ===
    id                      UUID DEFAULT generateUUIDv4() COMMENT 'Уникальный идентификатор записи',
    entity_type             LowCardinality(String) COMMENT 'Тип сущности: company/entrepreneur',
    entity_id               String COMMENT 'ОГРН или ОГРНИП',
    inn                     Nullable(String) COMMENT 'ИНН',
    
    -- === Сведения о записи ГРН ===
    grn                     String COMMENT 'Государственный регистрационный номер записи',
    grn_date                Date COMMENT 'Дата записи',
    
    -- === Причина изменения ===
    reason_code             Nullable(String) COMMENT 'Код причины внесения записи',
    reason_description      Nullable(String) COMMENT 'Описание причины',
    
    -- === Регистрирующий орган ===
    authority_code          Nullable(String) COMMENT 'Код регистрирующего органа',
    authority_name          Nullable(String) COMMENT 'Наименование регистрирующего органа',
    
    -- === Свидетельство ===
    certificate_series      Nullable(String) COMMENT 'Серия свидетельства',
    certificate_number      Nullable(String) COMMENT 'Номер свидетельства',
    certificate_date        Nullable(Date) COMMENT 'Дата выдачи свидетельства',
    
    -- === Снимок основных данных на момент изменения ===
    snapshot_full_name      Nullable(String) COMMENT 'Наименование на момент изменения',
    snapshot_status         Nullable(String) COMMENT 'Статус на момент изменения',
    snapshot_address        Nullable(String) COMMENT 'Адрес на момент изменения',
    snapshot_json           Nullable(String) COMMENT 'Полный снимок данных в JSON',
    
    -- === Метаданные для дедупликации ===
    source_files            Array(String) DEFAULT [] COMMENT 'Список файлов-источников',
    extract_date            Date COMMENT 'Дата выписки из XML',
    file_hash               String DEFAULT '' COMMENT 'Хеш файла для отслеживания источника',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи',
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата последнего обновления'
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(grn_date)
ORDER BY (entity_type, entity_id, grn)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- СОЗДАНИЕ ИНДЕКСОВ
-- ============================================================================

-- Индекс для поиска по типу сущности
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_entity_type entity_type TYPE set(10) GRANULARITY 4;

-- Индекс для поиска по ИНН
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для поиска по коду причины
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_reason_code reason_code TYPE set(500) GRANULARITY 4;

-- Индекс по дате ГРН
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_grn_date grn_date TYPE minmax GRANULARITY 4;

-- Индекс по дате выписки
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_extract_date extract_date TYPE minmax GRANULARITY 4;

-- Индекс по хешу файла
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_file_hash file_hash TYPE bloom_filter(0.01) GRANULARITY 4;

-- ============================================================================
-- СОЗДАНИЕ VIEW ДЛЯ ДЕДУПЛИЦИРОВАННЫХ ДАННЫХ
-- ============================================================================

CREATE OR REPLACE VIEW egrul.company_history_view AS
SELECT 
    id,
    entity_type,
    entity_id,
    inn,
    grn,
    grn_date,
    reason_code,
    reason_description,
    authority_code,
    authority_name,
    certificate_series,
    certificate_number,
    certificate_date,
    snapshot_full_name,
    snapshot_status,
    snapshot_address,
    snapshot_json,
    source_files,
    extract_date,
    file_hash,
    created_at,
    updated_at
FROM egrul.company_history FINAL
ORDER BY entity_id, grn_date DESC, grn DESC;

-- ============================================================================
-- МАТЕРИАЛИЗОВАННОЕ ПРЕДСТАВЛЕНИЕ ДЛЯ СТАТИСТИКИ
-- ============================================================================

-- Создаем таблицу для хранения статистики дедупликации
CREATE TABLE IF NOT EXISTS egrul.deduplication_stats
(
    date                    Date DEFAULT today(),
    entity_type             LowCardinality(String),
    total_records           UInt64,
    unique_records          UInt64,
    duplicates_removed      UInt64,
    deduplication_ratio     Float32,
    created_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = MergeTree()
ORDER BY (date, entity_type)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- ФУНКЦИИ ДЛЯ РАБОТЫ С ДЕДУПЛИКАЦИЕЙ (КОММЕНТАРИИ)
-- ============================================================================

-- Примечание: ClickHouse не поддерживает пользовательские функции в миграциях
-- Дедупликация будет происходить автоматически через ReplacingMergeTree
-- При необходимости ручной вставки используйте стандартные INSERT запросы

-- ============================================================================
-- МАТЕРИАЛИЗАЦИЯ ИНДЕКСОВ
-- ============================================================================

ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_entity_type;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_inn;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_reason_code;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_grn_date;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_extract_date;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_file_hash;

-- ============================================================================
-- ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ
-- ============================================================================

-- Пример вставки записи с автоматической дедупликацией:
/*
INSERT INTO egrul.company_history (
    entity_type, entity_id, grn, grn_date, reason_code, reason_description,
    source_files, extract_date, file_hash, created_at, updated_at
) VALUES (
    'company', '1067425004635', '1067425004635', '2006-09-11', '11201', 'Создание юридического лица',
    ['file1.xml'], '2024-08-01', 'hash1', now64(3), now64(3)
);
*/

-- Получение дедуплицированных данных:
/*
SELECT grn, grn_date, reason_code, source_files
FROM egrul.company_history_view
WHERE entity_id = '1067425004635'
ORDER BY grn_date DESC, grn DESC
LIMIT 10;
*/

-- Статистика дедупликации:
/*
SELECT 
    entity_type,
    count(*) as total_in_table,
    count(DISTINCT grn) as unique_grns,
    count(*) - count(DISTINCT grn) as potential_duplicates
FROM egrul.company_history
GROUP BY entity_type;
*/

-- ============================================================================
-- КОММЕНТАРИИ ПО ИСПОЛЬЗОВАНИЮ
-- ============================================================================

-- ВАЖНЫЕ ОСОБЕННОСТИ:
-- 1. Таблица использует ReplacingMergeTree с ключом сортировки (entity_type, entity_id, grn)
-- 2. Дедупликация происходит автоматически по ГРН в рамках одной сущности
-- 3. source_files содержит массив всех файлов-источников для отслеживания
-- 4. Для получения актуальных данных используйте VIEW company_history_view или добавляйте FINAL
-- 5. Периодически запускайте OPTIMIZE TABLE для принудительной дедупликации
-- 6. file_hash помогает отслеживать уже обработанные файлы

-- РЕКОМЕНДАЦИИ ДЛЯ ПАРСЕРА:
-- 1. Перед импортом вычисляйте хеш файла
-- 2. Проверяйте, не был ли файл уже обработан
-- 3. При вставке указывайте корректные extract_date из XML
-- 4. Используйте batch-вставки для повышения производительности