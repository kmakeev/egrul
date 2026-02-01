-- Миграция 012: Таблицы для отслеживания изменений в компаниях и ИП (КЛАСТЕР)
-- Описание: Создает таблицы для хранения истории изменений (change detection)
--           и поддержки системы уведомлений в кластерной архитектуре

-- ============================================================================
-- Таблица: company_changes_local (локальные данные на каждой ноде)
-- Описание: История изменений в данных компаний для системы уведомлений
-- ============================================================================

CREATE TABLE IF NOT EXISTS company_changes_local ON CLUSTER egrul_cluster (
    -- Идентификаторы
    ogrn String,                                    -- ОГРН компании
    inn String,                                     -- ИНН компании (для быстрого поиска)
    company_name String,                            -- Название компании на момент изменения

    -- Информация об изменении
    change_type String,                             -- Тип изменения: status, director, founder_added, founder_removed,
                                                    --                founder_share_change, address, capital, activity_added,
                                                    --                activity_removed, main_activity_change
    change_id String DEFAULT generateUUIDv4(),      -- Уникальный ID события (для idempotency в Kafka)

    -- Детали изменения
    field_name String,                              -- Название поля (если применимо)
    old_value String,                               -- Старое значение (JSON string)
    new_value String,                               -- Новое значение (JSON string)

    -- Контекст изменения
    change_description String,                      -- Человеко-читаемое описание изменения
    is_significant UInt8 DEFAULT 1,                 -- Флаг значимости (1 = значимое, 0 = незначимое)
                                                    -- Значимые: статус, руководитель, крупные изменения капитала
                                                    -- Незначимые: мелкие правки в адресе, добавление дополнительных ОКВЭД

    -- Дополнительные данные (JSON)
    metadata String DEFAULT '{}',                   -- Дополнительная информация (JSON)
                                                    -- Примеры: {"region_code": "77", "previous_status": "ACTIVE"}

    -- Временные метки
    detected_at DateTime DEFAULT now(),             -- Время детектирования изменения
    import_session_id String,                       -- ID сессии импорта (если применимо)

    -- Технические поля
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/company_changes_local', '{replica}', updated_at)
PARTITION BY toYYYYMM(detected_at)                  -- Партиционирование по месяцам
ORDER BY (ogrn, change_type, detected_at, change_id)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: company_changes (Distributed поверх company_changes_local)
-- Описание: Распределенная таблица для доступа ко всем изменениям компаний
-- ============================================================================

CREATE TABLE IF NOT EXISTS company_changes ON CLUSTER egrul_cluster AS company_changes_local
ENGINE = Distributed(egrul_cluster, default, company_changes_local, rand());

-- ============================================================================
-- Таблица: entrepreneur_changes_local (локальные данные на каждой ноде)
-- Описание: История изменений в данных индивидуальных предпринимателей
-- ============================================================================

CREATE TABLE IF NOT EXISTS entrepreneur_changes_local ON CLUSTER egrul_cluster (
    -- Идентификаторы
    ogrnip String,                                  -- ОГРНИП предпринимателя
    inn String,                                     -- ИНН ИП
    full_name String,                               -- ФИО предпринимателя

    -- Информация об изменении
    change_type String,                             -- Тип изменения: status, activity_added, activity_removed,
                                                    --                main_activity_change, address
    change_id String DEFAULT generateUUIDv4(),      -- Уникальный ID события

    -- Детали изменения
    field_name String,
    old_value String,
    new_value String,

    -- Контекст изменения
    change_description String,
    is_significant UInt8 DEFAULT 1,

    -- Дополнительные данные
    metadata String DEFAULT '{}',

    -- Временные метки
    detected_at DateTime DEFAULT now(),
    import_session_id String,

    -- Технические поля
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/entrepreneur_changes_local', '{replica}', updated_at)
PARTITION BY toYYYYMM(detected_at)
ORDER BY (ogrnip, change_type, detected_at, change_id)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: entrepreneur_changes (Distributed поверх entrepreneur_changes_local)
-- Описание: Распределенная таблица для доступа ко всем изменениям ИП
-- ============================================================================

CREATE TABLE IF NOT EXISTS entrepreneur_changes ON CLUSTER egrul_cluster AS entrepreneur_changes_local
ENGINE = Distributed(egrul_cluster, default, entrepreneur_changes_local, rand());

-- ============================================================================
-- Индексы для быстрого поиска
-- ============================================================================

-- Индексы для company_changes_local
ALTER TABLE company_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_company_inn (inn) TYPE bloom_filter GRANULARITY 3;

ALTER TABLE company_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_company_change_type (change_type) TYPE set(20) GRANULARITY 4;

ALTER TABLE company_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_company_is_significant (is_significant) TYPE set(2) GRANULARITY 4;

ALTER TABLE company_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_company_detected_date (toDate(detected_at)) TYPE minmax GRANULARITY 1;

-- Индексы для entrepreneur_changes_local
ALTER TABLE entrepreneur_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneur_inn (inn) TYPE bloom_filter GRANULARITY 3;

ALTER TABLE entrepreneur_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneur_change_type (change_type) TYPE set(20) GRANULARITY 4;

ALTER TABLE entrepreneur_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneur_is_significant (is_significant) TYPE set(2) GRANULARITY 4;

ALTER TABLE entrepreneur_changes_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneur_detected_date (toDate(detected_at)) TYPE minmax GRANULARITY 1;

-- ============================================================================
-- Материализованные представления для аналитики изменений
-- ============================================================================

-- МВ: Статистика изменений компаний по типам (за последние 30 дней)
-- Создаем локальную таблицу для каждой ноды
CREATE TABLE IF NOT EXISTS company_changes_stats_mv_local ON CLUSTER egrul_cluster
(
    change_date Date,
    change_type String,
    is_significant UInt8,
    changes_count UInt64,
    affected_companies_count AggregateFunction(uniq, String)
)
ENGINE = ReplicatedSummingMergeTree('/clickhouse/tables/{shard}/company_changes_stats_mv_local', '{replica}')
PARTITION BY toYYYYMM(change_date)
ORDER BY (change_date, change_type, is_significant);

-- Материализованное представление для автоматического заполнения
CREATE MATERIALIZED VIEW IF NOT EXISTS company_changes_stats_mv_insert ON CLUSTER egrul_cluster
TO company_changes_stats_mv_local
AS
SELECT
    toDate(detected_at) AS change_date,
    change_type,
    is_significant,
    count() AS changes_count,
    uniqState(ogrn) AS affected_companies_count
FROM company_changes_local
WHERE detected_at >= now() - INTERVAL 30 DAY
GROUP BY change_date, change_type, is_significant;

-- Distributed таблица поверх локальной
CREATE TABLE IF NOT EXISTS company_changes_stats_mv ON CLUSTER egrul_cluster AS company_changes_stats_mv_local
ENGINE = Distributed(egrul_cluster, default, company_changes_stats_mv_local, rand());

-- МВ: Статистика изменений ИП по типам
CREATE TABLE IF NOT EXISTS entrepreneur_changes_stats_mv_local ON CLUSTER egrul_cluster
(
    change_date Date,
    change_type String,
    is_significant UInt8,
    changes_count UInt64,
    affected_entrepreneurs_count AggregateFunction(uniq, String)
)
ENGINE = ReplicatedSummingMergeTree('/clickhouse/tables/{shard}/entrepreneur_changes_stats_mv_local', '{replica}')
PARTITION BY toYYYYMM(change_date)
ORDER BY (change_date, change_type, is_significant);

CREATE MATERIALIZED VIEW IF NOT EXISTS entrepreneur_changes_stats_mv_insert ON CLUSTER egrul_cluster
TO entrepreneur_changes_stats_mv_local
AS
SELECT
    toDate(detected_at) AS change_date,
    change_type,
    is_significant,
    count() AS changes_count,
    uniqState(ogrnip) AS affected_entrepreneurs_count
FROM entrepreneur_changes_local
WHERE detected_at >= now() - INTERVAL 30 DAY
GROUP BY change_date, change_type, is_significant;

CREATE TABLE IF NOT EXISTS entrepreneur_changes_stats_mv ON CLUSTER egrul_cluster AS entrepreneur_changes_stats_mv_local
ENGINE = Distributed(egrul_cluster, default, entrepreneur_changes_stats_mv_local, rand());

-- ============================================================================
-- Вспомогательные представления
-- ============================================================================

-- Представление: Последние изменения компаний (за последние 7 дней)
CREATE OR REPLACE VIEW recent_company_changes ON CLUSTER egrul_cluster AS
SELECT
    ogrn,
    company_name,
    change_type,
    change_description,
    old_value,
    new_value,
    is_significant,
    detected_at
FROM company_changes
WHERE detected_at >= now() - INTERVAL 7 DAY
ORDER BY detected_at DESC
LIMIT 1000;

-- Представление: Последние изменения ИП
CREATE OR REPLACE VIEW recent_entrepreneur_changes ON CLUSTER egrul_cluster AS
SELECT
    ogrnip,
    full_name,
    change_type,
    change_description,
    old_value,
    new_value,
    is_significant,
    detected_at
FROM entrepreneur_changes
WHERE detected_at >= now() - INTERVAL 7 DAY
ORDER BY detected_at DESC
LIMIT 1000;

-- Представление: Все изменения за сегодня (для дашборда)
CREATE OR REPLACE VIEW today_all_changes ON CLUSTER egrul_cluster AS
SELECT
    'company' AS entity_type,
    ogrn AS entity_id,
    company_name AS entity_name,
    change_type,
    change_description,
    is_significant,
    detected_at
FROM company_changes
WHERE toDate(detected_at) = today()

UNION ALL

SELECT
    'entrepreneur' AS entity_type,
    ogrnip AS entity_id,
    full_name AS entity_name,
    change_type,
    change_description,
    is_significant,
    detected_at
FROM entrepreneur_changes
WHERE toDate(detected_at) = today()

ORDER BY detected_at DESC;

-- ============================================================================
-- Словари типов изменений (для валидации и отображения)
-- ============================================================================

-- Создание словаря типов изменений компаний (на каждой ноде)
CREATE DICTIONARY IF NOT EXISTS company_change_types_dict ON CLUSTER egrul_cluster
(
    change_type String,
    display_name_ru String,
    display_name_en String,
    category String,              -- status, management, ownership, legal, activities
    severity UInt8                -- 1 = low, 2 = medium, 3 = high, 4 = critical
)
PRIMARY KEY change_type
SOURCE(CLICKHOUSE(
    QUERY $format$
    SELECT
        change_type,
        display_name_ru,
        display_name_en,
        category,
        severity
    FROM (
        SELECT 'status' AS change_type, 'Изменение статуса' AS display_name_ru, 'Status change' AS display_name_en, 'status' AS category, 4 AS severity
        UNION ALL SELECT 'director', 'Смена руководителя', 'Director change', 'management', 3
        UNION ALL SELECT 'founder_added', 'Добавлен учредитель', 'Founder added', 'ownership', 3
        UNION ALL SELECT 'founder_removed', 'Удален учредитель', 'Founder removed', 'ownership', 3
        UNION ALL SELECT 'founder_share_change', 'Изменение доли учредителя', 'Founder share change', 'ownership', 2
        UNION ALL SELECT 'address', 'Изменение адреса', 'Address change', 'legal', 2
        UNION ALL SELECT 'capital', 'Изменение уставного капитала', 'Capital change', 'legal', 2
        UNION ALL SELECT 'activity_added', 'Добавлен вид деятельности', 'Activity added', 'activities', 1
        UNION ALL SELECT 'activity_removed', 'Удален вид деятельности', 'Activity removed', 'activities', 1
        UNION ALL SELECT 'main_activity_change', 'Изменение основного ОКВЭД', 'Main activity change', 'activities', 2
    )
    $format$
))
LAYOUT(HASHED())
LIFETIME(MIN 0 MAX 0);

-- ============================================================================
-- Информация о миграции
-- ============================================================================

-- Версия: 012
-- Дата: 2026-01-31
-- Автор: Claude Code
-- Описание: Создание таблиц для отслеживания изменений в компаниях и ИП.
--           Поддержка системы уведомлений через Kafka.
--           Материализованные представления для аналитики изменений.
--           КЛАСТЕРНАЯ версия с Replicated*MergeTree и Distributed таблицами.
