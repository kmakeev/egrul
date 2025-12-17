-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Schema Migration
-- Version: 004_add_additional_activities
-- Description: Добавление JSON-поля additional_activities в основные таблицы
--              companies и entrepreneurs для последующей выгрузки ОКВЭД
-- ============================================================================

-- Таблица: egrul.companies
ALTER TABLE egrul.companies
    ADD COLUMN IF NOT EXISTS additional_activities Nullable(String)
    AFTER okved_additional_names;

-- Таблица: egrul.entrepreneurs
ALTER TABLE egrul.entrepreneurs
    ADD COLUMN IF NOT EXISTS additional_activities Nullable(String)
    AFTER okved_additional_names;
