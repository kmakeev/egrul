-- Миграция 012: Таблицы для отслеживания изменений в компаниях и ИП
-- Описание: Создает таблицы для хранения истории изменений (change detection)
--           и поддержки системы уведомлений

-- ============================================================================
-- Таблица: company_changes
-- Описание: История изменений в данных компаний для системы уведомлений
-- ============================================================================

CREATE TABLE IF NOT EXISTS company_changes (
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
ENGINE = ReplacingMergeTree(updated_at)             -- Для дедупликации при повторной обработке
PARTITION BY toYYYYMM(detected_at)                  -- Партиционирование по месяцам
ORDER BY (ogrn, change_type, detected_at, change_id)
SETTINGS index_granularity = 8192;

-- Комментарии к таблице
ALTER TABLE company_changes COMMENT COLUMN ogrn 'ОГРН компании';
ALTER TABLE company_changes COMMENT COLUMN change_type 'Тип изменения (status, director, founder_*, address, capital, activity_*)';
ALTER TABLE company_changes COMMENT COLUMN change_id 'Уникальный UUID события для idempotency в Kafka consumer';
ALTER TABLE company_changes COMMENT COLUMN old_value 'Старое значение в JSON формате';
ALTER TABLE company_changes COMMENT COLUMN new_value 'Новое значение в JSON формате';
ALTER TABLE company_changes COMMENT COLUMN is_significant 'Флаг значимости изменения (1 = важное, 0 = второстепенное)';
ALTER TABLE company_changes COMMENT COLUMN detected_at 'Время детектирования изменения (NOT время в реестре!)';

-- ============================================================================
-- Таблица: entrepreneur_changes
-- Описание: История изменений в данных индивидуальных предпринимателей
-- ============================================================================

CREATE TABLE IF NOT EXISTS entrepreneur_changes (
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
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(detected_at)
ORDER BY (ogrnip, change_type, detected_at, change_id)
SETTINGS index_granularity = 8192;

-- Комментарии к таблице
ALTER TABLE entrepreneur_changes COMMENT COLUMN ogrnip 'ОГРНИП индивидуального предпринимателя';
ALTER TABLE entrepreneur_changes COMMENT COLUMN change_type 'Тип изменения (status, activity_*, address)';
ALTER TABLE entrepreneur_changes COMMENT COLUMN change_id 'Уникальный UUID события для idempotency';

-- ============================================================================
-- Индексы для быстрого поиска
-- ============================================================================

-- Индексы для company_changes
ALTER TABLE company_changes ADD INDEX idx_company_inn (inn) TYPE bloom_filter GRANULARITY 3;
ALTER TABLE company_changes ADD INDEX idx_company_change_type (change_type) TYPE set(20) GRANULARITY 4;
ALTER TABLE company_changes ADD INDEX idx_company_is_significant (is_significant) TYPE set(2) GRANULARITY 4;
ALTER TABLE company_changes ADD INDEX idx_company_detected_date (toDate(detected_at)) TYPE minmax GRANULARITY 1;

-- Индексы для entrepreneur_changes
ALTER TABLE entrepreneur_changes ADD INDEX idx_entrepreneur_inn (inn) TYPE bloom_filter GRANULARITY 3;
ALTER TABLE entrepreneur_changes ADD INDEX idx_entrepreneur_change_type (change_type) TYPE set(20) GRANULARITY 4;
ALTER TABLE entrepreneur_changes ADD INDEX idx_entrepreneur_is_significant (is_significant) TYPE set(2) GRANULARITY 4;
ALTER TABLE entrepreneur_changes ADD INDEX idx_entrepreneur_detected_date (toDate(detected_at)) TYPE minmax GRANULARITY 1;

-- ============================================================================
-- Материализованные представления для аналитики изменений
-- ============================================================================

-- МВ: Статистика изменений компаний по типам (за последние 30 дней)
CREATE MATERIALIZED VIEW IF NOT EXISTS company_changes_stats_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(change_date)
ORDER BY (change_date, change_type, is_significant)
AS
SELECT
    toDate(detected_at) AS change_date,
    change_type,
    is_significant,
    count() AS changes_count,
    uniq(ogrn) AS affected_companies_count
FROM company_changes
WHERE detected_at >= now() - INTERVAL 30 DAY
GROUP BY change_date, change_type, is_significant;

-- МВ: Статистика изменений ИП по типам
CREATE MATERIALIZED VIEW IF NOT EXISTS entrepreneur_changes_stats_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(change_date)
ORDER BY (change_date, change_type, is_significant)
AS
SELECT
    toDate(detected_at) AS change_date,
    change_type,
    is_significant,
    count() AS changes_count,
    uniq(ogrnip) AS affected_entrepreneurs_count
FROM entrepreneur_changes
WHERE detected_at >= now() - INTERVAL 30 DAY
GROUP BY change_date, change_type, is_significant;

-- МВ: ТОП компаний с наибольшим количеством изменений (за месяц)
CREATE MATERIALIZED VIEW IF NOT EXISTS company_most_changes_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(detected_month)
ORDER BY (detected_month, ogrn)
AS
SELECT
    toStartOfMonth(detected_at) AS detected_month,
    ogrn,
    company_name,
    count() AS total_changes,
    sum(is_significant) AS significant_changes_count
FROM company_changes
WHERE detected_at >= now() - INTERVAL 1 MONTH
GROUP BY detected_month, ogrn, company_name;

-- ============================================================================
-- Вспомогательные функции и представления
-- ============================================================================

-- Представление: Последние изменения компаний (за последние 7 дней)
CREATE OR REPLACE VIEW recent_company_changes AS
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
CREATE OR REPLACE VIEW recent_entrepreneur_changes AS
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
CREATE OR REPLACE VIEW today_all_changes AS
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
-- Запросы для проверки работы таблиц
-- ============================================================================

-- Проверка: Подсчет изменений по типам (компании)
-- SELECT change_type, count() AS cnt, sum(is_significant) AS significant_cnt
-- FROM company_changes
-- WHERE detected_at >= now() - INTERVAL 1 DAY
-- GROUP BY change_type
-- ORDER BY cnt DESC;

-- Проверка: Топ компаний с изменениями
-- SELECT ogrn, company_name, count() AS changes_count
-- FROM company_changes
-- WHERE detected_at >= now() - INTERVAL 7 DAY
-- GROUP BY ogrn, company_name
-- ORDER BY changes_count DESC
-- LIMIT 20;

-- Проверка: Изменения конкретной компании
-- SELECT
--     change_type,
--     change_description,
--     old_value,
--     new_value,
--     detected_at
-- FROM company_changes
-- WHERE ogrn = '1234567890123'
-- ORDER BY detected_at DESC
-- LIMIT 10;

-- ============================================================================
-- TTL политики (опционально, для очистки старых данных)
-- ============================================================================

-- Удалять незначимые изменения старше 90 дней
-- ALTER TABLE company_changes
--     MODIFY TTL detected_at + INTERVAL 90 DAY DELETE WHERE is_significant = 0;

-- Удалять значимые изменения старше 2 лет
-- ALTER TABLE company_changes
--     MODIFY TTL detected_at + INTERVAL 730 DAY DELETE WHERE is_significant = 1;

-- Аналогично для entrepreneur_changes
-- ALTER TABLE entrepreneur_changes
--     MODIFY TTL detected_at + INTERVAL 90 DAY DELETE WHERE is_significant = 0;
-- ALTER TABLE entrepreneur_changes
--     MODIFY TTL detected_at + INTERVAL 730 DAY DELETE WHERE is_significant = 1;

-- ============================================================================
-- Словари типов изменений (для валидации и отображения)
-- ============================================================================

-- Создание словаря типов изменений компаний
CREATE DICTIONARY IF NOT EXISTS company_change_types_dict
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

-- Использование словаря в запросах:
-- SELECT
--     change_type,
--     dictGet('company_change_types_dict', 'display_name_ru', change_type) AS change_type_name,
--     dictGet('company_change_types_dict', 'severity', change_type) AS severity,
--     count() AS cnt
-- FROM company_changes
-- WHERE detected_at >= now() - INTERVAL 1 DAY
-- GROUP BY change_type
-- ORDER BY severity DESC, cnt DESC;

-- ============================================================================
-- Информация о миграции
-- ============================================================================

-- Версия: 012
-- Дата: 2026-01-22
-- Автор: Claude Code
-- Описание: Создание таблиц для отслеживания изменений в компаниях и ИП.
--           Поддержка системы уведомлений через Kafka.
--           Материализованные представления для аналитики изменений.
