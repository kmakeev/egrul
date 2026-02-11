-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Materialized Views для кластера
-- Version: 012_materialized_views_cluster
-- Description: Адаптация миграции 003 для кластерной архитектуры
-- Включены только критические MV для аналитического дашборда
-- ============================================================================

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по регионам (компании)
-- ============================================================================

-- Локальная таблица на каждом шарде
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    total_capital           AggregateFunction(sum, Nullable(Decimal(18, 2))),
    avg_capital             AggregateFunction(avg, Nullable(Decimal(18, 2))),
    min_registration_date   AggregateFunction(min, Nullable(Date)),
    max_registration_date   AggregateFunction(max, Nullable(Date)),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/egrul/stats_companies_by_region', '{replica}')
PARTITION BY region_code
ORDER BY (region_code, status);

-- Distributed таблица
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    total_capital           AggregateFunction(sum, Nullable(Decimal(18, 2))),
    avg_capital             AggregateFunction(avg, Nullable(Decimal(18, 2))),
    min_registration_date   AggregateFunction(min, Nullable(Date)),
    max_registration_date   AggregateFunction(max, Nullable(Date)),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_companies_by_region_local', rand());

-- Materialized View пишет в локальную таблицу
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_region ON CLUSTER egrul_cluster
TO egrul.stats_companies_by_region_local
AS SELECT
    region_code,
    any(region) AS region,
    status,
    countState() AS count,
    sumState(capital_amount) AS total_capital,
    avgState(capital_amount) AS avg_capital,
    minState(registration_date) AS min_registration_date,
    maxState(registration_date) AS max_registration_date
FROM egrul.companies_local
WHERE region_code IS NOT NULL AND region_code != ''
GROUP BY region_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по регионам (предприниматели)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    min_registration_date   AggregateFunction(min, Nullable(Date)),
    max_registration_date   AggregateFunction(max, Nullable(Date)),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/egrul/stats_entrepreneurs_by_region', '{replica}')
PARTITION BY region_code
ORDER BY (region_code, status);

CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region ON CLUSTER egrul_cluster
(
    region_code             LowCardinality(String),
    region                  Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    min_registration_date   AggregateFunction(min, Nullable(Date)),
    max_registration_date   AggregateFunction(max, Nullable(Date)),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_entrepreneurs_by_region_local', rand());

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_region ON CLUSTER egrul_cluster
TO egrul.stats_entrepreneurs_by_region_local
AS SELECT
    region_code,
    any(region) AS region,
    status,
    countState() AS count,
    minState(registration_date) AS min_registration_date,
    maxState(registration_date) AS max_registration_date
FROM egrul.entrepreneurs_local
WHERE region_code IS NOT NULL AND region_code != ''
GROUP BY region_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика регистраций по месяцам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    registration_month      Date,
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/egrul/stats_registrations_by_month', '{replica}')
PARTITION BY toYear(registration_month)
ORDER BY (entity_type, registration_month, status);

CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    registration_month      Date,
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_registrations_by_month_local', rand());

-- MV для компаний
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_company_registrations ON CLUSTER egrul_cluster
TO egrul.stats_registrations_by_month_local
AS SELECT
    'company' AS entity_type,
    toStartOfMonth(registration_date) AS registration_month,
    status,
    countState() AS count
FROM egrul.companies_local
WHERE registration_date IS NOT NULL
GROUP BY registration_month, status;

-- MV для предпринимателей
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneur_registrations ON CLUSTER egrul_cluster
TO egrul.stats_registrations_by_month_local
AS SELECT
    'entrepreneur' AS entity_type,
    toStartOfMonth(registration_date) AS registration_month,
    status,
    countState() AS count
FROM egrul.entrepreneurs_local
WHERE registration_date IS NOT NULL
GROUP BY registration_month, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика ликвидаций по месяцам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month_local ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/egrul/stats_terminations_by_month', '{replica}')
PARTITION BY toYear(termination_month)
ORDER BY (entity_type, termination_month);

CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month ON CLUSTER egrul_cluster
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = Distributed('egrul_cluster', 'egrul', 'stats_terminations_by_month_local', rand());

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_company_terminations ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS SELECT
    'company' AS entity_type,
    toStartOfMonth(termination_date) AS termination_month,
    countState() AS count
FROM egrul.companies_local
WHERE termination_date IS NOT NULL
GROUP BY termination_month;

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneur_terminations ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS SELECT
    'entrepreneur' AS entity_type,
    toStartOfMonth(termination_date) AS termination_month,
    countState() AS count
FROM egrul.entrepreneurs_local
WHERE termination_date IS NOT NULL
GROUP BY termination_month;

-- ============================================================================
-- VIEW: Удобные представления для API запросов (Distributed)
-- ============================================================================

-- Статистика по регионам (финальные значения)
CREATE OR REPLACE VIEW egrul.v_stats_companies_by_region ON CLUSTER egrul_cluster AS
SELECT
    region_code,
    region,
    status,
    countMerge(count) AS companies_count,
    sumMerge(total_capital) AS total_capital,
    avgMerge(avg_capital) AS avg_capital,
    minMerge(min_registration_date) AS first_registration,
    maxMerge(max_registration_date) AS last_registration
FROM egrul.stats_companies_by_region
GROUP BY region_code, region, status;

-- Динамика регистраций
CREATE OR REPLACE VIEW egrul.v_registration_dynamics ON CLUSTER egrul_cluster AS
SELECT
    entity_type,
    registration_month,
    status,
    countMerge(count) AS registrations_count
FROM egrul.stats_registrations_by_month
GROUP BY entity_type, registration_month, status
ORDER BY registration_month DESC;

-- Динамика ликвидаций
CREATE OR REPLACE VIEW egrul.v_termination_dynamics ON CLUSTER egrul_cluster AS
SELECT
    entity_type,
    termination_month,
    countMerge(count) AS terminations_count
FROM egrul.stats_terminations_by_month
GROUP BY entity_type, termination_month
ORDER BY termination_month DESC;

-- ============================================================================
-- Готово! MV автоматически начнут собирать статистику для новых данных
-- ============================================================================
