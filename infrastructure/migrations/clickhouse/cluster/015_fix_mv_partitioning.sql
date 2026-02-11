-- Миграция 015: Исправление партиционирования в Materialized Views
-- Проблема: партиционирование по месяцам (toYYYYMM) создает слишком много партиций при импорте исторических данных
-- Решение: изменить партиционирование на годовое (toYear) для уменьшения количества партиций

-- ============================================================
-- УДАЛЕНИЕ СТАРЫХ MATERIALIZED VIEWS ДЛЯ РЕГИСТРАЦИЙ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster SYNC;

-- ============================================================
-- УДАЛЕНИЕ СТАРЫХ MATERIALIZED VIEWS ДЛЯ ПРЕКРАЩЕНИЙ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster SYNC;

-- ============================================================
-- СОЗДАНИЕ НОВЫХ MATERIALIZED VIEWS С ГОДОВЫМ ПАРТИЦИОНИРОВАНИЕМ
-- ============================================================

-- ============================================================
-- 1. Статистика регистраций по месяцам (партиции по годам)
-- ============================================================

-- Локальная таблица для агрегированных данных
-- ИЗМЕНЕНИЕ: PARTITION BY toYear(registration_month) вместо toYYYYMM
CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String), -- 'company' или 'entrepreneur'
    registration_month      Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYear(registration_month) -- ИСПРАВЛЕНО: партиции по годам вместо месяцев
ORDER BY (entity_type, registration_month)
SETTINGS index_granularity = 8192;

-- Distributed таблица для чтения
CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster AS egrul.stats_registrations_by_month_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_registrations_by_month_local', rand());

-- Materialized View для компаний
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_registrations_by_month_companies_local ON CLUSTER egrul_cluster
TO egrul.stats_registrations_by_month_local
AS
SELECT
    'company' as entity_type,
    toStartOfMonth(registration_date) as registration_month,
    countState() as count,
    now64(3) as updated_at
FROM egrul.companies_local
WHERE registration_date IS NOT NULL
GROUP BY registration_month;

-- Materialized View для ИП
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_registrations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster
TO egrul.stats_registrations_by_month_local
AS
SELECT
    'entrepreneur' as entity_type,
    toStartOfMonth(registration_date) as registration_month,
    countState() as count,
    now64(3) as updated_at
FROM egrul.entrepreneurs_local
WHERE registration_date IS NOT NULL
GROUP BY registration_month;

-- ============================================================
-- 2. Статистика прекращений по месяцам (партиции по годам)
-- ============================================================

-- Локальная таблица для агрегированных данных
-- ИЗМЕНЕНИЕ: PARTITION BY toYear(termination_month) вместо toYYYYMM
CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYear(termination_month) -- ИСПРАВЛЕНО: партиции по годам вместо месяцев
ORDER BY (entity_type, termination_month)
SETTINGS index_granularity = 8192;

-- Distributed таблица для чтения
CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster AS egrul.stats_terminations_by_month_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_terminations_by_month_local', rand());

-- Materialized View для компаний
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS
SELECT
    'company' as entity_type,
    toStartOfMonth(termination_date) as termination_month,
    countState() as count,
    now64(3) as updated_at
FROM egrul.companies_local
WHERE termination_date IS NOT NULL
GROUP BY termination_month;

-- Materialized View для ИП
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_terminations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS
SELECT
    'entrepreneur' as entity_type,
    toStartOfMonth(termination_date) as termination_month,
    countState() as count,
    now64(3) as updated_at
FROM egrul.entrepreneurs_local
WHERE termination_date IS NOT NULL
GROUP BY termination_month;
