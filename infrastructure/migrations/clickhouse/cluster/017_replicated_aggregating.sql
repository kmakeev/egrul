-- Миграция 017: Переход на ReplicatedAggregatingMergeTree для агрегатных таблиц
-- Проблема: AggregatingMergeTree без репликации приводит к нестабильным данным в Distributed запросах
-- Решение: использовать ReplicatedAggregatingMergeTree для репликации агрегатов между нодами

-- ВАЖНО: Эта миграция пересоздаёт MV и требует повторного заполнения данных через make cluster-fill-mv

-- ============================================================
-- 1. СТАТИСТИКА КОМПАНИЙ ПО РЕГИОНАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_companies_by_region ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster SYNC;

-- Локальная таблица с репликацией
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  String,
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/stats_companies_by_region', '{replica}')
PARTITION BY tuple()
ORDER BY (region_code, status)
SETTINGS index_granularity = 8192;

-- Distributed таблица
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region ON CLUSTER egrul_cluster AS egrul.stats_companies_by_region_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_companies_by_region_local', rand());

-- Materialized View
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_companies_by_region_local
AS
SELECT
    region_code,
    coalesce(any(region), '') as region,
    multiIf(
        status_code IN ('113', '114', '115', '116', '117'), 'bankrupt',
        termination_date IS NOT NULL OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'), 'liquidated',
        'active'
    ) as status,
    countState() as count,
    now64(3) as updated_at
FROM egrul.companies_local
GROUP BY region_code, status;

-- ============================================================
-- 2. СТАТИСТИКА ИП ПО РЕГИОНАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_entrepreneurs_by_region ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster SYNC;

-- Локальная таблица с репликацией
CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  String,
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/stats_entrepreneurs_by_region', '{replica}')
PARTITION BY tuple()
ORDER BY (region_code, status)
SETTINGS index_granularity = 8192;

-- Distributed таблица
CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region ON CLUSTER egrul_cluster AS egrul.stats_entrepreneurs_by_region_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_entrepreneurs_by_region_local', rand());

-- Materialized View
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_entrepreneurs_by_region_local
AS
SELECT
    region_code,
    coalesce(any(region), '') as region,
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
-- 3. СТАТИСТИКА РЕГИСТРАЦИЙ ПО МЕСЯЦАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_registrations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster SYNC;

-- Локальная таблица с репликацией
CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    registration_month      Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/stats_registrations_by_month', '{replica}')
PARTITION BY toYear(registration_month)
ORDER BY (entity_type, registration_month)
SETTINGS index_granularity = 8192;

-- Distributed таблица
CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster AS egrul.stats_registrations_by_month_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_registrations_by_month_local', rand());

-- Materialized Views
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
-- 4. СТАТИСТИКА ПРЕКРАЩЕНИЙ ПО МЕСЯЦАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster SYNC;
DROP TABLE IF EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster SYNC;

-- Локальная таблица с репликацией
CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/stats_terminations_by_month', '{replica}')
PARTITION BY toYear(termination_month)
ORDER BY (entity_type, termination_month)
SETTINGS index_granularity = 8192;

-- Distributed таблица
CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster AS egrul.stats_terminations_by_month_local
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_terminations_by_month_local', rand());

-- Materialized Views
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
