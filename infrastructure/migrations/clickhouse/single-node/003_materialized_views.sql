-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Materialized Views Migration
-- Version: 003_materialized_views
-- Description: Materialized Views для аналитики и агрегированных данных
-- ============================================================================

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по регионам (компании)
-- ============================================================================

-- Целевая таблица для хранения агрегированных данных
CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_region
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
ENGINE = AggregatingMergeTree()
PARTITION BY region_code
ORDER BY (region_code, status);

-- Materialized View для автоматического обновления
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_region
TO egrul.stats_companies_by_region
AS SELECT
    region_code,
    any(region) AS region,
    status,
    countState() AS count,
    sumState(capital_amount) AS total_capital,
    avgState(capital_amount) AS avg_capital,
    minState(registration_date) AS min_registration_date,
    maxState(registration_date) AS max_registration_date
FROM egrul.companies
WHERE region_code IS NOT NULL AND region_code != ''
GROUP BY region_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по регионам (предприниматели)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_region
(
    region_code             LowCardinality(String),
    region                  Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    min_registration_date   AggregateFunction(min, Nullable(Date)),
    max_registration_date   AggregateFunction(max, Nullable(Date)),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY region_code
ORDER BY (region_code, status);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_region
TO egrul.stats_entrepreneurs_by_region
AS SELECT
    region_code,
    any(region) AS region,
    status,
    countState() AS count,
    minState(registration_date) AS min_registration_date,
    maxState(registration_date) AS max_registration_date
FROM egrul.entrepreneurs
WHERE region_code IS NOT NULL AND region_code != ''
GROUP BY region_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по ОКВЭД (компании)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_okved
(
    okved_main_code         String,
    okved_main_name         Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    total_capital           AggregateFunction(sum, Nullable(Decimal(18, 2))),
    avg_capital             AggregateFunction(avg, Nullable(Decimal(18, 2))),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
ORDER BY (okved_main_code, status);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_okved
TO egrul.stats_companies_by_okved
AS SELECT
    okved_main_code,
    any(okved_main_name) AS okved_main_name,
    status,
    countState() AS count,
    sumState(capital_amount) AS total_capital,
    avgState(capital_amount) AS avg_capital
FROM egrul.companies
WHERE okved_main_code IS NOT NULL AND okved_main_code != ''
GROUP BY okved_main_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по ОКВЭД (предприниматели)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_entrepreneurs_by_okved
(
    okved_main_code         String,
    okved_main_name         Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
ORDER BY (okved_main_code, status);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_okved
TO egrul.stats_entrepreneurs_by_okved
AS SELECT
    okved_main_code,
    any(okved_main_name) AS okved_main_name,
    status,
    countState() AS count
FROM egrul.entrepreneurs
WHERE okved_main_code IS NOT NULL AND okved_main_code != ''
GROUP BY okved_main_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика по ОПФ (организационно-правовая форма)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_companies_by_opf
(
    opf_code                String,
    opf_name                Nullable(String),
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    total_capital           AggregateFunction(sum, Nullable(Decimal(18, 2))),
    avg_capital             AggregateFunction(avg, Nullable(Decimal(18, 2))),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
ORDER BY (opf_code, status);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_opf
TO egrul.stats_companies_by_opf
AS SELECT
    opf_code,
    any(opf_name) AS opf_name,
    status,
    countState() AS count,
    sumState(capital_amount) AS total_capital,
    avgState(capital_amount) AS avg_capital
FROM egrul.companies
WHERE opf_code IS NOT NULL AND opf_code != ''
GROUP BY opf_code, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика регистраций по месяцам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_registrations_by_month
(
    entity_type             LowCardinality(String),
    registration_month      Date,
    status                  LowCardinality(String),
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYear(registration_month)
ORDER BY (entity_type, registration_month, status);

-- MV для компаний
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_company_registrations
TO egrul.stats_registrations_by_month
AS SELECT
    'company' AS entity_type,
    toStartOfMonth(registration_date) AS registration_month,
    status,
    countState() AS count
FROM egrul.companies
WHERE registration_date IS NOT NULL
GROUP BY registration_month, status;

-- MV для предпринимателей
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneur_registrations
TO egrul.stats_registrations_by_month
AS SELECT
    'entrepreneur' AS entity_type,
    toStartOfMonth(registration_date) AS registration_month,
    status,
    countState() AS count
FROM egrul.entrepreneurs
WHERE registration_date IS NOT NULL
GROUP BY registration_month, status;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика ликвидаций по месяцам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_terminations_by_month
(
    entity_type             LowCardinality(String),
    termination_month       Date,
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYear(termination_month)
ORDER BY (entity_type, termination_month);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_company_terminations
TO egrul.stats_terminations_by_month
AS SELECT
    'company' AS entity_type,
    toStartOfMonth(termination_date) AS termination_month,
    countState() AS count
FROM egrul.companies
WHERE termination_date IS NOT NULL
GROUP BY termination_month;

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneur_terminations
TO egrul.stats_terminations_by_month
AS SELECT
    'entrepreneur' AS entity_type,
    toStartOfMonth(termination_date) AS termination_month,
    countState() AS count
FROM egrul.entrepreneurs
WHERE termination_date IS NOT NULL
GROUP BY termination_month;

-- ============================================================================
-- АГРЕГАЦИЯ: Топ владельцев по количеству компаний
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_top_owners
(
    owner_type              LowCardinality(String),
    owner_id                String DEFAULT '',
    owner_inn               String DEFAULT '',
    owner_name              String,
    owned_companies_count   AggregateFunction(count),
    total_share_percent     AggregateFunction(sum, Nullable(Decimal(10, 4))),
    avg_share_percent       AggregateFunction(avg, Nullable(Decimal(10, 4))),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
ORDER BY (owner_type, owner_inn, owner_name);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_top_owners
TO egrul.stats_top_owners
AS SELECT
    owner_type,
    owner_id,
    owner_inn,
    owner_name,
    countState() AS owned_companies_count,
    sumState(share_percent) AS total_share_percent,
    avgState(share_percent) AS avg_share_percent
FROM egrul.ownership_graph
WHERE is_active = 1
GROUP BY owner_type, owner_id, owner_inn, owner_name;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика владения по типам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_ownership_by_type
(
    owner_type              LowCardinality(String),
    is_active               UInt8,
    count                   AggregateFunction(count),
    avg_share_percent       AggregateFunction(avg, Nullable(Decimal(10, 4))),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
ORDER BY (owner_type, is_active);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_ownership_by_type
TO egrul.stats_ownership_by_type
AS SELECT
    owner_type,
    is_active,
    countState() AS count,
    avgState(share_percent) AS avg_share_percent
FROM egrul.ownership_graph
GROUP BY owner_type, is_active;

-- ============================================================================
-- АГРЕГАЦИЯ: Статистика изменений (ГРН) по месяцам
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_changes_by_month
(
    entity_type             LowCardinality(String),
    change_month            Date,
    reason_code             String DEFAULT '',
    count                   AggregateFunction(count),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYear(change_month)
ORDER BY (entity_type, change_month, reason_code);

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_changes_by_month
TO egrul.stats_changes_by_month
AS SELECT
    entity_type,
    toStartOfMonth(grn_date) AS change_month,
    reason_code,
    countState() AS count
FROM egrul.company_history
GROUP BY entity_type, change_month, reason_code;

-- ============================================================================
-- АГРЕГАЦИЯ: Общая сводка по базе
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.stats_summary
(
    stat_name               LowCardinality(String),
    stat_value              UInt64,
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY stat_name;

-- ============================================================================
-- VIEW: Удобные представления для API запросов
-- ============================================================================

-- Активные компании с полной информацией
CREATE OR REPLACE VIEW egrul.v_active_companies AS
SELECT
    ogrn,
    inn,
    kpp,
    full_name,
    short_name,
    opf_code,
    opf_name,
    registration_date,
    region_code,
    region,
    city,
    full_address,
    capital_amount,
    capital_currency,
    concat(head_last_name, ' ', head_first_name, ' ', coalesce(head_middle_name, '')) AS head_fio,
    head_position,
    okved_main_code,
    okved_main_name,
    founders_count,
    email,
    version_date
FROM egrul.companies
WHERE status = 'Действующее';

-- Активные предприниматели
CREATE OR REPLACE VIEW egrul.v_active_entrepreneurs AS
SELECT
    ogrnip,
    inn,
    last_name,
    first_name,
    middle_name,
    concat(last_name, ' ', first_name, ' ', coalesce(middle_name, '')) AS full_name,
    registration_date,
    region_code,
    region,
    city,
    okved_main_code,
    okved_main_name,
    email,
    version_date
FROM egrul.entrepreneurs
WHERE status = 'Действующее';

-- Статистика по регионам (финальные значения)
CREATE OR REPLACE VIEW egrul.v_stats_companies_by_region AS
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

-- Статистика по ОКВЭД (финальные значения)
CREATE OR REPLACE VIEW egrul.v_stats_companies_by_okved AS
SELECT
    okved_main_code,
    okved_main_name,
    status,
    countMerge(count) AS companies_count,
    sumMerge(total_capital) AS total_capital,
    avgMerge(avg_capital) AS avg_capital
FROM egrul.stats_companies_by_okved
GROUP BY okved_main_code, okved_main_name, status;

-- Топ владельцев
CREATE OR REPLACE VIEW egrul.v_top_owners AS
SELECT
    owner_type,
    owner_id,
    owner_inn,
    owner_name,
    countMerge(owned_companies_count) AS companies_count,
    sumMerge(total_share_percent) AS total_share,
    avgMerge(avg_share_percent) AS avg_share
FROM egrul.stats_top_owners
GROUP BY owner_type, owner_id, owner_inn, owner_name
ORDER BY companies_count DESC;

-- Динамика регистраций
CREATE OR REPLACE VIEW egrul.v_registration_dynamics AS
SELECT
    entity_type,
    registration_month,
    status,
    countMerge(count) AS registrations_count
FROM egrul.stats_registrations_by_month
GROUP BY entity_type, registration_month, status
ORDER BY registration_month DESC;

-- Динамика ликвидаций
CREATE OR REPLACE VIEW egrul.v_termination_dynamics AS
SELECT
    entity_type,
    termination_month,
    countMerge(count) AS terminations_count
FROM egrul.stats_terminations_by_month
GROUP BY entity_type, termination_month
ORDER BY termination_month DESC;

-- ============================================================================
-- VIEW: Поиск связанных компаний (через общих учредителей)
-- ============================================================================
-- Примечание: Для поиска связанных компаний используйте запрос:
-- SELECT o1.target_ogrn, o1.target_name, o2.target_ogrn, o2.target_name
-- FROM egrul.ownership_graph o1, egrul.ownership_graph o2
-- WHERE o1.owner_inn = o2.owner_inn AND o1.target_ogrn != o2.target_ogrn
--   AND o1.is_active = 1 AND o2.is_active = 1;

CREATE OR REPLACE VIEW egrul.v_owner_companies AS
SELECT
    owner_type,
    owner_inn,
    owner_name,
    groupArray(target_ogrn) AS owned_ogrns,
    groupArray(target_name) AS owned_names,
    count() AS companies_count
FROM egrul.ownership_graph
WHERE is_active = 1
GROUP BY owner_type, owner_inn, owner_name
HAVING count() > 1;

-- ============================================================================
-- VIEW: Компании с банкротством
-- ============================================================================

CREATE OR REPLACE VIEW egrul.v_bankrupt_companies AS
SELECT
    ogrn,
    inn,
    full_name,
    status,
    bankruptcy_stage,
    registration_date,
    region,
    city,
    okved_main_code,
    okved_main_name,
    capital_amount
FROM egrul.companies
WHERE is_bankrupt = 1 OR status = 'Банкрот'
ORDER BY registration_date DESC;

-- ============================================================================
-- VIEW: Недавно зарегистрированные компании
-- ============================================================================

CREATE OR REPLACE VIEW egrul.v_recent_companies AS
SELECT
    ogrn,
    inn,
    full_name,
    short_name,
    opf_name,
    registration_date,
    region,
    city,
    okved_main_code,
    okved_main_name,
    capital_amount,
    concat(head_last_name, ' ', head_first_name) AS head_name,
    created_at
FROM egrul.companies
WHERE registration_date >= today() - INTERVAL 30 DAY
ORDER BY registration_date DESC;

-- ============================================================================
-- VIEW: Недавно ликвидированные компании
-- ============================================================================

CREATE OR REPLACE VIEW egrul.v_recent_liquidations AS
SELECT
    ogrn,
    inn,
    full_name,
    status,
    termination_method,
    registration_date,
    termination_date,
    region,
    city,
    okved_main_code,
    capital_amount
FROM egrul.companies
WHERE termination_date >= today() - INTERVAL 30 DAY
ORDER BY termination_date DESC;

-- ============================================================================
-- Обновление сводной статистики (запускать периодически)
-- ============================================================================

-- Можно использовать scheduled query или внешний cron для периодического обновления
INSERT INTO egrul.stats_summary (stat_name, stat_value)
SELECT 'total_companies', count() FROM egrul.companies;

INSERT INTO egrul.stats_summary (stat_name, stat_value)
SELECT 'active_companies', count() FROM egrul.companies WHERE status = 'Действующее';

INSERT INTO egrul.stats_summary (stat_name, stat_value)
SELECT 'total_entrepreneurs', count() FROM egrul.entrepreneurs;

INSERT INTO egrul.stats_summary (stat_name, stat_value)
SELECT 'active_entrepreneurs', count() FROM egrul.entrepreneurs WHERE status = 'Действующее';

INSERT INTO egrul.stats_summary (stat_name, stat_value)
SELECT 'total_ownership_links', count() FROM egrul.ownership_graph WHERE is_active = 1;

