-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Indexes Migration
-- Version: 002_indexes
-- Description: Создание индексов и проекций для оптимизации запросов
-- ============================================================================

-- ============================================================================
-- ИНДЕКСЫ ДЛЯ ТАБЛИЦЫ companies
-- ============================================================================

-- Bloom filter индекс для быстрого поиска по ИНН
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Bloom filter для поиска по КПП
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_kpp kpp TYPE bloom_filter(0.01) GRANULARITY 4;

-- Полнотекстовый поиск по наименованию (ngrambf - n-gram bloom filter)
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_full_name_ngram full_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

-- Bloom filter для поиска по короткому наименованию (Nullable - используем bloom_filter)
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_short_name short_name TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для фильтрации по статусу
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_status status TYPE set(100) GRANULARITY 4;

-- Индекс для фильтрации по региону
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_region_code region_code TYPE set(100) GRANULARITY 4;

-- Индекс для фильтрации по коду ОКВЭД
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_okved_main okved_main_code TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для поиска по массиву дополнительных ОКВЭД
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_okved_additional okved_additional TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для фильтрации по дате регистрации
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_registration_date registration_date TYPE minmax GRANULARITY 4;

-- Индекс для фильтрации по уставному капиталу
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_capital capital_amount TYPE minmax GRANULARITY 4;

-- Индекс для поиска по email
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_email email TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для поиска по ОПФ
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_opf_code opf_code TYPE set(500) GRANULARITY 4;

-- Индекс для поиска по городу (Nullable - используем bloom_filter)
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_city city TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для ИНН руководителя
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_head_inn head_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по дате версии для инкрементальных обновлений
ALTER TABLE egrul.companies
    ADD INDEX IF NOT EXISTS idx_companies_version_date version_date TYPE minmax GRANULARITY 4;

-- ============================================================================
-- ИНДЕКСЫ ДЛЯ ТАБЛИЦЫ entrepreneurs
-- ============================================================================

-- Bloom filter для поиска по ИНН
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- N-gram для поиска по фамилии
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_last_name last_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

-- N-gram для поиска по имени
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_first_name first_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

-- Индекс по статусу
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_status status TYPE set(100) GRANULARITY 4;

-- Индекс по региону
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_region_code region_code TYPE set(100) GRANULARITY 4;

-- Индекс по основному ОКВЭД
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_okved_main okved_main_code TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по дополнительным ОКВЭД
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_okved_additional okved_additional TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по дате регистрации
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_registration_date registration_date TYPE minmax GRANULARITY 4;

-- Индекс по email
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_email email TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по дате версии
ALTER TABLE egrul.entrepreneurs
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_version_date version_date TYPE minmax GRANULARITY 4;

-- ============================================================================
-- ИНДЕКСЫ ДЛЯ ТАБЛИЦЫ company_history
-- ============================================================================

-- Индекс для поиска по типу сущности
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_entity_type entity_type TYPE set(10) GRANULARITY 4;

-- Индекс для поиска по ИНН
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для поиска по коду причины
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_reason_code reason_code TYPE set(500) GRANULARITY 4;

-- Индекс по дате ГРН
ALTER TABLE egrul.company_history
    ADD INDEX IF NOT EXISTS idx_history_grn_date grn_date TYPE minmax GRANULARITY 4;

-- ============================================================================
-- ИНДЕКСЫ ДЛЯ ТАБЛИЦЫ ownership_graph
-- ============================================================================

-- Индекс для поиска владельца по типу
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_owner_type owner_type TYPE set(10) GRANULARITY 4;

-- Индекс для поиска по ОГРН владельца
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_owner_id owner_id TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для поиска по ИНН владельца
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_owner_inn owner_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- N-gram для поиска по имени владельца
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_owner_name owner_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

-- Индекс для поиска по ИНН целевой компании
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_target_inn target_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс для фильтрации активных связей
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_is_active is_active TYPE set(2) GRANULARITY 4;

-- Индекс для фильтрации по проценту владения
ALTER TABLE egrul.ownership_graph
    ADD INDEX IF NOT EXISTS idx_ownership_share_percent share_percent TYPE minmax GRANULARITY 4;

-- ============================================================================
-- ИНДЕКСЫ ДЛЯ ТАБЛИЦЫ founders
-- ============================================================================

-- Индекс по типу учредителя
ALTER TABLE egrul.founders
    ADD INDEX IF NOT EXISTS idx_founders_type founder_type TYPE set(10) GRANULARITY 4;

-- Индекс по ОГРН учредителя
ALTER TABLE egrul.founders
    ADD INDEX IF NOT EXISTS idx_founders_ogrn founder_ogrn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по ИНН учредителя
ALTER TABLE egrul.founders
    ADD INDEX IF NOT EXISTS idx_founders_inn founder_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Индекс по ИНН компании
ALTER TABLE egrul.founders
    ADD INDEX IF NOT EXISTS idx_founders_company_inn company_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- ============================================================================
-- ПРОЕКЦИИ (закомментированы для MVP, требуют настройки deduplicate_merge_projection_mode)
-- ============================================================================
-- Проекции не поддерживаются в ReplacingMergeTree без дополнительной настройки.
-- Для production можно раскомментировать после добавления:
-- ALTER TABLE egrul.companies MODIFY SETTING deduplicate_merge_projection_mode = 'rebuild';

-- ============================================================================
-- МАТЕРИАЛИЗАЦИЯ ИНДЕКСОВ
-- ============================================================================
-- Материализуем индексы для существующих данных

ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_inn;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_kpp;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_full_name_ngram;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_short_name;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_status;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_region_code;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_okved_main;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_okved_additional;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_registration_date;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_capital;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_email;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_opf_code;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_city;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_head_inn;
ALTER TABLE egrul.companies MATERIALIZE INDEX idx_companies_version_date;

ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_inn;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_last_name;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_first_name;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_status;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_region_code;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_okved_main;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_okved_additional;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_registration_date;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_email;
ALTER TABLE egrul.entrepreneurs MATERIALIZE INDEX idx_entrepreneurs_version_date;

ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_entity_type;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_inn;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_reason_code;
ALTER TABLE egrul.company_history MATERIALIZE INDEX idx_history_grn_date;

ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_owner_type;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_owner_id;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_owner_inn;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_owner_name;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_target_inn;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_is_active;
ALTER TABLE egrul.ownership_graph MATERIALIZE INDEX idx_ownership_share_percent;

ALTER TABLE egrul.founders MATERIALIZE INDEX idx_founders_type;
ALTER TABLE egrul.founders MATERIALIZE INDEX idx_founders_ogrn;
ALTER TABLE egrul.founders MATERIALIZE INDEX idx_founders_inn;
ALTER TABLE egrul.founders MATERIALIZE INDEX idx_founders_company_inn;

