-- ============================================================
-- Миграция 014: Исправление MV для ликвидаций компаний
-- Проблема: termination_date всегда NULL для компаний
-- Решение: Использовать extract_date для ликвидированных по status_code
-- ============================================================

-- 1. Удаляем старые MV для компаний
DROP VIEW IF EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster;

-- 2. Создаем новую MV с правильной логикой
CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_terminations_by_month_companies_local ON CLUSTER egrul_cluster
TO egrul.stats_terminations_by_month_local
AS
SELECT
    'company' as entity_type,
    toStartOfMonth(
        COALESCE(
            termination_date,  -- Если есть termination_date - используем его
            -- Иначе для ликвидированных по status_code используем extract_date
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
WHERE
    termination_date IS NOT NULL
    OR status_code IN ('101', '105', '106', '107', '113', '114', '115', '116', '117', '701', '702', '801', '802')
GROUP BY termination_month;
