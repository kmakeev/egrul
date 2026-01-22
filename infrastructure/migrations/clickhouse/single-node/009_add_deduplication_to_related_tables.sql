-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Schema Migration
-- Version: 009_add_deduplication_to_related_tables
-- Description: Добавление дедупликации для таблиц лицензий, филиалов и ОКВЭД
--              по аналогии с company_history
-- ============================================================================

-- ============================================================================
-- Шаг 1: Добавление поля updated_at в таблицу licenses
-- ============================================================================

ALTER TABLE egrul.licenses
    ADD COLUMN IF NOT EXISTS updated_at DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи';

-- ============================================================================
-- Шаг 2: Пересоздание таблицы licenses с ReplacingMergeTree
-- ============================================================================

-- Создаем новую таблицу с ReplacingMergeTree
-- ORDER BY включает activity для уникальности (license_number может быть пустым)
CREATE TABLE IF NOT EXISTS egrul.licenses_new
(
    id                      UUID DEFAULT generateUUIDv4(),

    -- === Владелец лицензии ===
    entity_type             LowCardinality(String) COMMENT 'Тип: company/entrepreneur',
    entity_ogrn             String COMMENT 'ОГРН или ОГРНИП',
    entity_inn              Nullable(String) COMMENT 'ИНН',
    entity_name             Nullable(String) COMMENT 'Наименование',

    -- === Сведения о лицензии ===
    license_number          String COMMENT 'Номер лицензии',
    license_series          Nullable(String) COMMENT 'Серия лицензии',
    activity                String DEFAULT '' COMMENT 'Вид деятельности',
    start_date              Nullable(Date) COMMENT 'Дата начала действия',
    end_date                Nullable(Date) COMMENT 'Дата окончания действия',
    authority               Nullable(String) COMMENT 'Лицензирующий орган',
    status                  LowCardinality(String) DEFAULT '' COMMENT 'Статус лицензии',

    -- === Метаданные ===
    version_date            Date DEFAULT today(),
    created_at              DateTime64(3) DEFAULT now64(3),
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(version_date)
ORDER BY (entity_ogrn, entity_type, license_number, activity)
SETTINGS index_granularity = 8192;

-- Копируем данные из старой таблицы
INSERT INTO egrul.licenses_new
SELECT
    id,
    entity_type,
    entity_ogrn,
    entity_inn,
    entity_name,
    license_number,
    license_series,
    coalesce(activity, '') as activity,
    start_date,
    end_date,
    authority,
    status,
    version_date,
    created_at,
    coalesce(updated_at, created_at) as updated_at
FROM egrul.licenses;

-- Удаляем старую таблицу и переименовываем новую
DROP TABLE IF EXISTS egrul.licenses_old;
RENAME TABLE egrul.licenses TO egrul.licenses_old;
RENAME TABLE egrul.licenses_new TO egrul.licenses;
DROP TABLE IF EXISTS egrul.licenses_old;

-- ============================================================================
-- Шаг 3: Добавление поля updated_at в таблицу branches
-- ============================================================================

ALTER TABLE egrul.branches
    ADD COLUMN IF NOT EXISTS updated_at DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи';

-- ============================================================================
-- Шаг 4: Пересоздание таблицы branches с ReplacingMergeTree
-- ============================================================================

-- Создаем новую таблицу с ReplacingMergeTree
-- ORDER BY включает full_address для уникальности (branch_kpp часто пустой)
CREATE TABLE IF NOT EXISTS egrul.branches_new
(
    id                      UUID DEFAULT generateUUIDv4(),

    -- === Головная компания ===
    company_ogrn            String COMMENT 'ОГРН головной компании',
    company_inn             Nullable(String) COMMENT 'ИНН головной компании',
    company_name            Nullable(String) COMMENT 'Наименование головной компании',

    -- === Сведения о филиале ===
    branch_type             LowCardinality(String) COMMENT 'Тип: branch/representative',
    branch_name             Nullable(String) COMMENT 'Наименование',
    branch_kpp              String DEFAULT '' COMMENT 'КПП филиала',

    -- === Адрес филиала ===
    postal_code             Nullable(String),
    region_code             Nullable(String),
    region                  Nullable(String),
    city                    Nullable(String),
    full_address            String DEFAULT '' COMMENT 'Полный адрес',

    -- === ГРН ===
    grn                     Nullable(String),
    grn_date                Nullable(Date),

    -- === Метаданные ===
    version_date            Date DEFAULT today(),
    created_at              DateTime64(3) DEFAULT now64(3),
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(version_date)
ORDER BY (company_ogrn, branch_type, branch_kpp, full_address)
SETTINGS index_granularity = 8192;

-- Копируем данные из старой таблицы
INSERT INTO egrul.branches_new
SELECT
    id,
    company_ogrn,
    company_inn,
    company_name,
    branch_type,
    branch_name,
    branch_kpp,
    postal_code,
    region_code,
    region,
    city,
    coalesce(full_address, '') as full_address,
    grn,
    grn_date,
    version_date,
    created_at,
    coalesce(updated_at, created_at) as updated_at
FROM egrul.branches;

-- Удаляем старую таблицу и переименовываем новую
DROP TABLE IF EXISTS egrul.branches_old;
RENAME TABLE egrul.branches TO egrul.branches_old;
RENAME TABLE egrul.branches_new TO egrul.branches;
DROP TABLE IF EXISTS egrul.branches_old;

-- ============================================================================
-- Шаг 5: Добавление поля updated_at в таблицу companies_okved_additional
-- ============================================================================

ALTER TABLE egrul.companies_okved_additional
    ADD COLUMN IF NOT EXISTS updated_at DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи';

-- ============================================================================
-- Шаг 6: Пересоздание таблицы companies_okved_additional с ReplacingMergeTree
-- ============================================================================

-- Создаем новую таблицу с ReplacingMergeTree
CREATE TABLE IF NOT EXISTS egrul.companies_okved_additional_new
(
    ogrn        String COMMENT 'ОГРН компании',
    inn         Nullable(String) COMMENT 'ИНН компании',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД',
    updated_at  DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (ogrn, okved_code);

-- Копируем данные из старой таблицы
INSERT INTO egrul.companies_okved_additional_new
SELECT
    ogrn,
    inn,
    okved_code,
    okved_name,
    coalesce(updated_at, now64(3)) as updated_at
FROM egrul.companies_okved_additional;

-- Удаляем старую таблицу и переименовываем новую
DROP TABLE IF EXISTS egrul.companies_okved_additional_old;
RENAME TABLE egrul.companies_okved_additional TO egrul.companies_okved_additional_old;
RENAME TABLE egrul.companies_okved_additional_new TO egrul.companies_okved_additional;
DROP TABLE IF EXISTS egrul.companies_okved_additional_old;

-- ============================================================================
-- Шаг 7: Добавление поля updated_at в таблицу entrepreneurs_okved_additional
-- ============================================================================

ALTER TABLE egrul.entrepreneurs_okved_additional
    ADD COLUMN IF NOT EXISTS updated_at DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи';

-- ============================================================================
-- Шаг 8: Пересоздание таблицы entrepreneurs_okved_additional с ReplacingMergeTree
-- ============================================================================

-- Создаем новую таблицу с ReplacingMergeTree
CREATE TABLE IF NOT EXISTS egrul.entrepreneurs_okved_additional_new
(
    ogrnip      String COMMENT 'ОГРНИП предпринимателя',
    inn         String COMMENT 'ИНН предпринимателя',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД',
    updated_at  DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (ogrnip, okved_code);

-- Копируем данные из старой таблицы
INSERT INTO egrul.entrepreneurs_okved_additional_new
SELECT
    ogrnip,
    inn,
    okved_code,
    okved_name,
    coalesce(updated_at, now64(3)) as updated_at
FROM egrul.entrepreneurs_okved_additional;

-- Удаляем старую таблицу и переименовываем новую
DROP TABLE IF EXISTS egrul.entrepreneurs_okved_additional_old;
RENAME TABLE egrul.entrepreneurs_okved_additional TO egrul.entrepreneurs_okved_additional_old;
RENAME TABLE egrul.entrepreneurs_okved_additional_new TO egrul.entrepreneurs_okved_additional;
DROP TABLE IF EXISTS egrul.entrepreneurs_okved_additional_old;

-- ============================================================================
-- Шаг 9: Оптимизация таблиц для применения дедупликации
-- ============================================================================

OPTIMIZE TABLE egrul.licenses FINAL;
OPTIMIZE TABLE egrul.branches FINAL;
OPTIMIZE TABLE egrul.companies_okved_additional FINAL;
OPTIMIZE TABLE egrul.entrepreneurs_okved_additional FINAL;
