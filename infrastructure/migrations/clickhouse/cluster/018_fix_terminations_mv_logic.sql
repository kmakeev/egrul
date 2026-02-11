-- Миграция 018: Исправление логики MV для ликвидаций
-- Проблема: MV использует только termination_date, но некоторые компании имеют status_code без termination_date
-- Решение: использовать ту же логику что и при ручном заполнении (COALESCE + status_code)

-- ============================================================
-- СТАТИСТИКА ПРЕКРАЩЕНИЙ ПО МЕСЯЦАМ (КОМПАНИИ)
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster SYNC;

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS
SELECT
    'company' as entity_type,
    toStartOfMonth(
        COALESCE(
            termination_date,
            multiIf(
                status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802'),
                extract_date,
                NULL
            )
        )
    ) as termination_month,
    countState() as count,
    now64(3) as updated_at
FROM egrul.companies_local
WHERE termination_date IS NOT NULL
   OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802')
GROUP BY termination_month;

-- ============================================================
-- СТАТИСТИКА ПРЕКРАЩЕНИЙ ПО МЕСЯЦАМ (ИП)
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_terminations_by_month_entrepreneurs_local ON CLUSTER egrul_cluster SYNC;

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
