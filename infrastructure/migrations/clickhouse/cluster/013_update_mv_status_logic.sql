-- Миграция 013: Обновление Materialized Views с правильной логикой определения статусов
-- Соответствует логике определения статусов в frontend badge компонентах
-- Использует AggregatingMergeTree (без репликации) для упрощения работы с MV

-- ============================================================
-- УДАЛЕНИЕ СТАРЫХ MATERIALIZED VIEWS
-- ============================================================

-- Останавливаем и удаляем MV для компаний
DROP TABLE IF EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_companies_by_region ON CLUSTER egrul_cluster SYNC;

-- Останавливаем и удаляем MV для ИП
DROP TABLE IF EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_entrepreneurs_by_region ON CLUSTER egrul_cluster SYNC;

-- Останавливаем и удаляем MV для регистраций
DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster SYNC;

-- Останавливаем и удаляем MV для прекращений
DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster SYNC;

-- ============================================================
-- СОЗДАНИЕ НОВЫХ MATERIALIZED VIEWS БЕЗ РЕПЛИКАЦИИ
-- ============================================================

-- ============================================================
-- 1. Статистика компаний по регионам с правильной логикой статусов
-- ============================================================

-- Локальная таблица для агрегированных данных (БЕЗ репликации)
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  String,
    status                  LowCardinality(String), -- 'active', 'liquidated', 'bankrupt'
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY tuple()
ORDER BY (region_code, status)
SETTINGS index_granularity = 8192;

-- Distributed таблица для чтения
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region ON CLUSTER egrul_cluster AS egrul.stats_companies_by_region_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_companies_by_region_local', rand());

-- Materialized View для автоматического обновления
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_companies_by_region_local
AS
SELECT
    region_code,
    any(region) as region,
    -- Логика определения статуса согласно company-status-badge.tsx:
    -- active: нет termination_date И код НЕ в списке недействующих
    -- liquidated: есть termination_date ИЛИ код в списке недействующих
    -- bankrupt: код банкротства (113-117)
    multiIf(
        -- Банкротство (приоритетнее чем просто ликвидация)
        status_code IN ('113', '114', '115', '116', '117'), 'bankrupt',
        -- Ликвидирована (согласно company-status-badge.tsx getVariantByCode строки 92-120)
        termination_date IS NOT NULL OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'), 'liquidated',
        -- Активная (по умолчанию) - включая реорганизацию (коды 121-139) и прочие
        'active'
    ) as status,
    countState() as count,
    now64(3) as updated_at
FROM egrul.companies_local
GROUP BY region_code, status;

-- ============================================================
-- 2. Статистика ИП по регионам с правильной логикой статусов
-- ============================================================

-- Локальная таблица для агрегированных данных (БЕЗ репликации)
CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  String,
    status                  LowCardinality(String), -- 'active', 'liquidated'
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY tuple()
ORDER BY (region_code, status)
SETTINGS index_granularity = 8192;

-- Distributed таблица для чтения
CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region ON CLUSTER egrul_cluster AS egrul.stats_entrepreneurs_by_region_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_entrepreneurs_by_region_local', rand());

-- Materialized View для автоматического обновления
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_entrepreneurs_by_region_local
AS
SELECT
    region_code,
    any(region) as region,
    -- Логика определения статуса согласно entrepreneur-status-badge.tsx:
    -- active: НЕТ termination_date И НЕТ status_code (NULL)
    -- liquidated: ЕСТЬ termination_date ИЛИ ЕСТЬ любой status_code
    if(
        termination_date IS NULL AND status_code IS NULL,
        'active',
        'liquidated'
    ) as status,
    countState() as count,
    now64(3) as updated_at
FROM egrul.entrepreneurs_local
GROUP BY region_code, status;

-- ============================================================
-- 3. Статистика регистраций по месяцам
-- ============================================================

-- Локальная таблица для агрегированных данных (БЕЗ репликации)
CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String), -- 'company' или 'entrepreneur'
    registration_month      Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(registration_month)
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
-- 4. Статистика прекращений по месяцам
-- ============================================================

-- Локальная таблица для агрегированных данных (БЕЗ репликации)
CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(termination_month)
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
