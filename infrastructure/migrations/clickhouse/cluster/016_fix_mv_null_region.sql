-- Миграция 016: Исправление NULL значений в столбце region материализованных представлений
-- Проблема: any(region) может вернуть NULL, но целевой столбец имеет тип String (не nullable)
-- Решение: использовать coalesce(any(region), '') для замены NULL на пустую строку

-- ============================================================
-- ПЕРЕСОЗДАНИЕ MV ДЛЯ КОМПАНИЙ ПО РЕГИОНАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster SYNC;

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_companies_by_region_local
AS
SELECT
    region_code,
    coalesce(any(region), '') as region, -- ИСПРАВЛЕНО: заменяем NULL на пустую строку
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
-- ПЕРЕСОЗДАНИЕ MV ДЛЯ ИП ПО РЕГИОНАМ
-- ============================================================

DROP TABLE IF EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster SYNC;

CREATE MATERIALIZED VIEW IF NOT EXISTS egrul.mv_stats_entrepreneurs_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_entrepreneurs_by_region_local
AS
SELECT
    region_code,
    coalesce(any(region), '') as region, -- ИСПРАВЛЕНО: заменяем NULL на пустую строку
    if(
        termination_date IS NULL AND status_code IS NULL,
        'active',
        'liquidated'
    ) as status,
    countState() as count,
    now64(3) as updated_at
FROM egrul.entrepreneurs_local
GROUP BY region_code, status;
