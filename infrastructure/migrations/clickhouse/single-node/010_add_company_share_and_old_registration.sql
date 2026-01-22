-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Schema Migration
-- Version: 010_add_company_share_and_old_registration
-- Description: Добавление полей для доли компании в УК и регистрации до 01.07.2002
-- ============================================================================

-- Добавление полей для доли компании в уставном капитале
ALTER TABLE egrul.companies
    ADD COLUMN IF NOT EXISTS company_share_percent Nullable(Decimal(10, 4)) COMMENT 'Процент доли компании в УК',
    ADD COLUMN IF NOT EXISTS company_share_nominal Nullable(Decimal(18, 2)) COMMENT 'Номинальная стоимость доли компании в УК';

-- Добавление полей для регистрации до 01.07.2002
ALTER TABLE egrul.companies
    ADD COLUMN IF NOT EXISTS old_reg_number Nullable(String) COMMENT 'Регистрационный номер до 01.07.2002',
    ADD COLUMN IF NOT EXISTS old_reg_date Nullable(Date) COMMENT 'Дата регистрации до 01.07.2002',
    ADD COLUMN IF NOT EXISTS old_reg_authority Nullable(String) COMMENT 'Орган регистрации до 01.07.2002';
