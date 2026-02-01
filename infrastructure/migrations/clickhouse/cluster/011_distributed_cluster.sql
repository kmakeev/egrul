-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Schema Migration
-- Version: 011_distributed_cluster
-- Description: Преобразование single-node схемы в распределенный кластер
-- ============================================================================
--
-- ВАЖНО: Данная миграция НЕ мигрирует существующие данные!
-- Все таблицы будут пересозданы с нуля.
-- Данные необходимо загрузить заново через скрипты импорта.
--
-- Архитектура кластера:
-- - Кластер: egrul_cluster
-- - Конфигурация: 3 шарда × 2 реплики
-- - Для каждой таблицы создается:
--   1. _local версия с Replicated*MergeTree (на каждом узле)
--   2. Distributed таблица поверх _local (точка входа для запросов)
--
-- Ключи шардирования:
-- - companies: cityHash64(ogrn)
-- - entrepreneurs: cityHash64(ogrnip)
-- - company_history, ownership_graph, founders: cityHash64(entity_id/ogrn/ogrnip)
-- - licenses, branches: cityHash64(entity_ogrn/company_ogrn)
-- - *_okved_additional: cityHash64(ogrn/ogrnip)
-- - import_log: rand() (без репликации)
--
-- ============================================================================

-- ============================================================================
-- ТАБЛИЦА: companies (Юридические лица - ЕГРЮЛ)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.companies ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.companies_local ON CLUSTER egrul_cluster
(
    -- === Основные идентификаторы ===
    ogrn                    String COMMENT 'ОГРН - основной государственный регистрационный номер',
    ogrn_date               Nullable(Date) COMMENT 'Дата присвоения ОГРН',
    inn                     String COMMENT 'ИНН - идентификационный номер налогоплательщика',
    kpp                     Nullable(String) COMMENT 'КПП - код причины постановки на учет',

    -- === Наименование ===
    full_name               String COMMENT 'Полное наименование',
    short_name              Nullable(String) COMMENT 'Сокращенное наименование',
    brand_name              Nullable(String) COMMENT 'Фирменное наименование',

    -- === Организационно-правовая форма ===
    opf_code                Nullable(String) COMMENT 'Код ОПФ',
    opf_name                Nullable(String) COMMENT 'Наименование ОПФ',

    -- === Статус и даты ===
    status                  LowCardinality(String) DEFAULT 'unknown' COMMENT 'Статус юридического лица',
    status_code             Nullable(String) COMMENT 'Код статуса',
    termination_method      Nullable(String) COMMENT 'Способ прекращения деятельности',
    registration_date       Nullable(Date) COMMENT 'Дата регистрации',
    termination_date        Nullable(Date) COMMENT 'Дата прекращения деятельности',
    extract_date            Date DEFAULT toDate('1970-01-01') COMMENT 'Дата выписки (ДатаВып из XML, версия записи)',

    -- === Адрес ===
    postal_code             Nullable(String) COMMENT 'Почтовый индекс',
    region_code             Nullable(String) COMMENT 'Код региона',
    region                  Nullable(String) COMMENT 'Наименование региона',
    district                Nullable(String) COMMENT 'Район',
    city                    Nullable(String) COMMENT 'Город',
    locality                Nullable(String) COMMENT 'Населенный пункт',
    street                  Nullable(String) COMMENT 'Улица',
    house                   Nullable(String) COMMENT 'Дом',
    building                Nullable(String) COMMENT 'Корпус',
    flat                    Nullable(String) COMMENT 'Квартира/Офис',
    full_address            Nullable(String) COMMENT 'Полный адрес одной строкой',
    fias_id                 Nullable(String) COMMENT 'ФИАС код',
    kladr_code              Nullable(String) COMMENT 'Код КЛАДР',
    email                   Nullable(String) COMMENT 'Адрес электронной почты',

    -- === Уставный капитал ===
    capital_amount          Nullable(Decimal(18, 2)) COMMENT 'Размер уставного капитала',
    capital_currency        Nullable(String) DEFAULT 'RUB' COMMENT 'Валюта капитала',
    company_share_percent   Nullable(Decimal(10, 4)) COMMENT 'Процент доли компании в УК',
    company_share_nominal   Nullable(Decimal(18, 2)) COMMENT 'Номинальная стоимость доли компании в УК',

    -- === Регистрация до 01.07.2002 ===
    old_reg_number          Nullable(String) COMMENT 'Регистрационный номер до 01.07.2002',
    old_reg_date            Nullable(Date) COMMENT 'Дата регистрации до 01.07.2002',
    old_reg_authority       Nullable(String) COMMENT 'Орган регистрации до 01.07.2002',

    -- === Руководитель ===
    head_last_name          Nullable(String) COMMENT 'Фамилия руководителя',
    head_first_name         Nullable(String) COMMENT 'Имя руководителя',
    head_middle_name        Nullable(String) COMMENT 'Отчество руководителя',
    head_inn                Nullable(String) COMMENT 'ИНН руководителя',
    head_position           Nullable(String) COMMENT 'Должность руководителя',
    head_position_code      Nullable(String) COMMENT 'Код должности',

    -- === Виды деятельности (ОКВЭД) ===
    okved_main_code         Nullable(String) COMMENT 'Код основного ОКВЭД',
    okved_main_name         Nullable(String) COMMENT 'Наименование основного ОКВЭД',
    okved_additional        Array(String) DEFAULT [] COMMENT 'Коды дополнительных ОКВЭД',
    okved_additional_names  Array(String) DEFAULT [] COMMENT 'Наименования дополнительных ОКВЭД',
    additional_activities   Nullable(String) COMMENT 'Дополнительные ОКВЭД в формате JSON (как в Parquet)',

    -- === Регистрация и налоговый учет ===
    reg_authority_code      Nullable(String) COMMENT 'Код регистрирующего органа',
    reg_authority_name      Nullable(String) COMMENT 'Наименование регистрирующего органа',
    tax_authority_code      Nullable(String) COMMENT 'Код налогового органа',
    tax_authority_name      Nullable(String) COMMENT 'Наименование налогового органа',

    -- === ПФР и ФСС ===
    pfr_reg_number          Nullable(String) COMMENT 'Регистрационный номер в ПФР',
    fss_reg_number          Nullable(String) COMMENT 'Регистрационный номер в ФСС',

    -- === Количественные показатели ===
    founders_count          Nullable(UInt16) DEFAULT 0 COMMENT 'Количество учредителей',
    licenses_count          Nullable(UInt16) DEFAULT 0 COMMENT 'Количество лицензий',
    branches_count          Nullable(UInt16) DEFAULT 0 COMMENT 'Количество филиалов',

    -- === Специальные статусы ===
    is_bankrupt             UInt8 DEFAULT 0 COMMENT 'Признак банкротства (0/1)',
    bankruptcy_stage        Nullable(String) COMMENT 'Стадия банкротства',
    is_liquidating          UInt8 DEFAULT 0 COMMENT 'В процессе ликвидации (0/1)',
    is_reorganizing         UInt8 DEFAULT 0 COMMENT 'В процессе реорганизации (0/1)',

    -- === История изменений (последняя запись ГРН) ===
    last_grn                Nullable(String) COMMENT 'Последний ГРН',
    last_grn_date           Nullable(Date) COMMENT 'Дата последнего ГРН',

    -- === Метаданные ===
    document_id             Nullable(String) COMMENT 'Идентификатор документа в выписке',
    source_file             Nullable(String) COMMENT 'Источник данных (имя файла)',
    version_date            Date DEFAULT today() COMMENT 'Дата версии данных (служебная)',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи',
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/companies_local',
    '{replica}',
    extract_date
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (ogrn, inn)
PRIMARY KEY ogrn
SETTINGS index_granularity = 8192,
         ttl_only_drop_parts = 1;

-- Добавляем индексы
ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_kpp kpp TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_full_name_ngram full_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_short_name short_name TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_status status TYPE set(100) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_region_code region_code TYPE set(100) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_okved_main okved_main_code TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_okved_additional okved_additional TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_registration_date registration_date TYPE minmax GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_capital capital_amount TYPE minmax GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_email email TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_opf_code opf_code TYPE set(500) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_city city TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_head_inn head_inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.companies_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_companies_version_date version_date TYPE minmax GRANULARITY 4;

-- Создаем Distributed таблицу
-- ВАЖНО: Шардирование по OGRN гарантирует, что все версии одной компании попадут на один шард
-- Это позволяет ReplacingMergeTree правильно дедуплицировать записи
CREATE TABLE IF NOT EXISTS egrul.companies ON CLUSTER egrul_cluster
AS egrul.companies_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    companies_local,
    cityHash64(ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: entrepreneurs (Индивидуальные предприниматели - ЕГРИП)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.entrepreneurs ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.entrepreneurs_local ON CLUSTER egrul_cluster
(
    -- === Основные идентификаторы ===
    ogrnip                  String COMMENT 'ОГРНИП - основной государственный регистрационный номер ИП',
    ogrnip_date             Nullable(Date) COMMENT 'Дата присвоения ОГРНИП',
    inn                     String COMMENT 'ИНН',

    -- === Сведения о физическом лице ===
    last_name               String COMMENT 'Фамилия',
    first_name              String COMMENT 'Имя',
    middle_name             Nullable(String) COMMENT 'Отчество',
    gender                  LowCardinality(String) DEFAULT '' COMMENT 'Пол (Мужской/Женский)',

    -- === Гражданство ===
    citizenship_type        LowCardinality(String) DEFAULT '' COMMENT 'Тип гражданства',
    citizenship_country_code Nullable(String) COMMENT 'Код страны (ОКСМ)',
    citizenship_country_name Nullable(String) COMMENT 'Наименование страны',

    -- === Статус и даты ===
    status                  LowCardinality(String) DEFAULT 'unknown' COMMENT 'Статус ИП',
    status_code             Nullable(String) COMMENT 'Код статуса',
    termination_method      Nullable(String) COMMENT 'Способ прекращения деятельности',
    registration_date       Nullable(Date) COMMENT 'Дата регистрации',
    termination_date        Nullable(Date) COMMENT 'Дата прекращения деятельности',
    extract_date            Date DEFAULT toDate('1970-01-01') COMMENT 'Дата выписки (ДатаВып из XML, версия записи)',

    -- === Адрес ===
    postal_code             Nullable(String) COMMENT 'Почтовый индекс',
    region_code             Nullable(String) COMMENT 'Код региона',
    region                  Nullable(String) COMMENT 'Наименование региона',
    district                Nullable(String) COMMENT 'Район',
    city                    Nullable(String) COMMENT 'Город',
    locality                Nullable(String) COMMENT 'Населенный пункт',
    street                  Nullable(String) COMMENT 'Улица',
    house                   Nullable(String) COMMENT 'Дом',
    building                Nullable(String) COMMENT 'Корпус/Строение',
    flat                    Nullable(String) COMMENT 'Квартира',
    full_address            Nullable(String) COMMENT 'Полный адрес',
    fias_id                 Nullable(String) COMMENT 'ФИАС код',
    kladr_code              Nullable(String) COMMENT 'Код КЛАДР',
    email                   Nullable(String) COMMENT 'Адрес электронной почты',

    -- === Виды деятельности (ОКВЭД) ===
    okved_main_code         Nullable(String) COMMENT 'Код основного ОКВЭД',
    okved_main_name         Nullable(String) COMMENT 'Наименование основного ОКВЭД',
    okved_additional        Array(String) DEFAULT [] COMMENT 'Коды дополнительных ОКВЭД',
    okved_additional_names  Array(String) DEFAULT [] COMMENT 'Наименования дополнительных ОКВЭД',
    additional_activities   Nullable(String) COMMENT 'Дополнительные ОКВЭД в формате JSON (как в Parquet)',

    -- === Регистрация и налоговый учет ===
    reg_authority_code      Nullable(String) COMMENT 'Код регистрирующего органа',
    reg_authority_name      Nullable(String) COMMENT 'Наименование регистрирующего органа',
    tax_authority_code      Nullable(String) COMMENT 'Код налогового органа',
    tax_authority_name      Nullable(String) COMMENT 'Наименование налогового органа',

    -- === ПФР и ФСС ===
    pfr_reg_number          Nullable(String) COMMENT 'Регистрационный номер в ПФР',
    fss_reg_number          Nullable(String) COMMENT 'Регистрационный номер в ФСС',

    -- === Лицензии ===
    licenses_count          Nullable(UInt16) DEFAULT 0 COMMENT 'Количество лицензий',

    -- === Специальные статусы ===
    is_bankrupt             UInt8 DEFAULT 0 COMMENT 'Признак банкротства (0/1)',
    bankruptcy_date         Nullable(Date) COMMENT 'Дата признания банкротом',
    bankruptcy_case_number  Nullable(String) COMMENT 'Номер дела о банкротстве',

    -- === История изменений (последняя запись ГРН) ===
    last_grn                Nullable(String) COMMENT 'Последний ГРН',
    last_grn_date           Nullable(Date) COMMENT 'Дата последнего ГРН',

    -- === Метаданные ===
    document_id             Nullable(String) COMMENT 'Идентификатор документа в выписке',
    source_file             Nullable(String) COMMENT 'Источник данных (имя файла)',
    version_date            Date DEFAULT today() COMMENT 'Дата версии данных (служебная)',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи',
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/entrepreneurs_local',
    '{replica}',
    extract_date
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (ogrnip, inn)
PRIMARY KEY ogrnip
SETTINGS index_granularity = 8192,
         ttl_only_drop_parts = 1;

-- Добавляем индексы
ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_last_name last_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_first_name first_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_status status TYPE set(100) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_region_code region_code TYPE set(100) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_okved_main okved_main_code TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_okved_additional okved_additional TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_registration_date registration_date TYPE minmax GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_email email TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.entrepreneurs_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_entrepreneurs_version_date version_date TYPE minmax GRANULARITY 4;

-- Создаем Distributed таблицу
-- ВАЖНО: Шардирование по ОГРНИП гарантирует, что все версии одного ИП попадут на один шард
-- Это позволяет ReplacingMergeTree правильно дедуплицировать записи
CREATE TABLE IF NOT EXISTS egrul.entrepreneurs ON CLUSTER egrul_cluster
AS egrul.entrepreneurs_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    entrepreneurs_local,
    cityHash64(ogrnip)
);

-- ============================================================================
-- ТАБЛИЦА: ownership_graph (Граф собственности)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.ownership_graph ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.ownership_graph_local ON CLUSTER egrul_cluster
(
    -- === Идентификаторы связи ===
    id                      UUID DEFAULT generateUUIDv4() COMMENT 'Уникальный идентификатор связи',

    -- === Владелец (учредитель) ===
    owner_type              LowCardinality(String) COMMENT 'Тип владельца: person/russian_company/foreign_company/public_entity/fund',
    owner_id                String DEFAULT '' COMMENT 'ОГРН/ОГРНИП владельца (если юр.лицо или ИП)',
    owner_inn               String DEFAULT '' COMMENT 'ИНН владельца',
    owner_name              String COMMENT 'Наименование/ФИО владельца',
    owner_country           Nullable(String) COMMENT 'Страна владельца (для иностранных)',
    owner_reg_number        Nullable(String) COMMENT 'Регистрационный номер (для иностранных)',

    -- === Целевая компания (владеемая) ===
    target_ogrn             String COMMENT 'ОГРН целевой компании',
    target_inn              Nullable(String) COMMENT 'ИНН целевой компании',
    target_name             Nullable(String) COMMENT 'Наименование целевой компании',

    -- === Доля владения ===
    share_nominal_value     Nullable(Decimal(18, 2)) COMMENT 'Номинальная стоимость доли',
    share_numerator         Nullable(Int64) COMMENT 'Числитель дроби доли',
    share_denominator       Nullable(Int64) COMMENT 'Знаменатель дроби доли',
    share_percent           Nullable(Decimal(10, 4)) COMMENT 'Процент доли',

    -- === Даты ===
    start_date              Nullable(Date) COMMENT 'Дата начала владения',
    end_date                Nullable(Date) COMMENT 'Дата окончания владения',
    is_active               UInt8 DEFAULT 1 COMMENT 'Признак актуальности (0/1)',

    -- === ГРН записи ===
    grn                     Nullable(String) COMMENT 'ГРН записи об учредителе',
    grn_date                Nullable(Date) COMMENT 'Дата ГРН',

    -- === Метаданные ===
    source_file             Nullable(String) COMMENT 'Источник данных',
    version_date            Date DEFAULT today() COMMENT 'Дата версии данных',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи',
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/ownership_graph_local',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (target_ogrn, owner_type, owner_id, owner_name)
SETTINGS index_granularity = 8192;

-- Добавляем индексы
ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_owner_type owner_type TYPE set(10) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_owner_id owner_id TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_owner_inn owner_inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_owner_name owner_name TYPE ngrambf_v1(3, 256, 2, 0) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_target_inn target_inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_is_active is_active TYPE set(2) GRANULARITY 4;

ALTER TABLE egrul.ownership_graph_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_ownership_share_percent share_percent TYPE minmax GRANULARITY 4;

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.ownership_graph ON CLUSTER egrul_cluster
AS egrul.ownership_graph_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    ownership_graph_local,
    cityHash64(target_ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: founders (Учредители - денормализованная)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.founders ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.founders_local ON CLUSTER egrul_cluster
(
    id                      UUID DEFAULT generateUUIDv4(),

    -- === Целевая компания ===
    company_ogrn            String COMMENT 'ОГРН компании',
    company_inn             Nullable(String) COMMENT 'ИНН компании',
    company_name            Nullable(String) COMMENT 'Наименование компании',

    -- === Учредитель ===
    founder_type            LowCardinality(String) COMMENT 'Тип учредителя',
    founder_ogrn            Nullable(String) COMMENT 'ОГРН учредителя (юр.лицо)',
    founder_inn             String DEFAULT '' COMMENT 'ИНН учредителя',
    founder_name            String COMMENT 'Наименование/ФИО',
    founder_last_name       Nullable(String) COMMENT 'Фамилия (физ.лицо)',
    founder_first_name      Nullable(String) COMMENT 'Имя (физ.лицо)',
    founder_middle_name     Nullable(String) COMMENT 'Отчество (физ.лицо)',
    founder_country         Nullable(String) COMMENT 'Страна (иностранцы)',
    founder_citizenship     Nullable(String) COMMENT 'Гражданство (физ.лицо)',

    -- === Доля ===
    share_nominal_value     Nullable(Decimal(18, 2)),
    share_percent           Nullable(Decimal(10, 4)),

    -- === Метаданные ===
    version_date            Date DEFAULT today(),
    created_at              DateTime64(3) DEFAULT now64(3),
    updated_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/founders_local',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (company_ogrn, founder_type, founder_inn, founder_name)
SETTINGS index_granularity = 8192;

-- Добавляем индексы
ALTER TABLE egrul.founders_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_founders_type founder_type TYPE set(10) GRANULARITY 4;

ALTER TABLE egrul.founders_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_founders_ogrn founder_ogrn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.founders_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_founders_inn founder_inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.founders_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_founders_company_inn company_inn TYPE bloom_filter(0.01) GRANULARITY 4;

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.founders ON CLUSTER egrul_cluster
AS egrul.founders_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    founders_local,
    cityHash64(company_ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: licenses (Лицензии)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.licenses ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.licenses_local ON CLUSTER egrul_cluster
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
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/licenses_local',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (entity_ogrn, entity_type, license_number, activity)
SETTINGS index_granularity = 8192;

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.licenses ON CLUSTER egrul_cluster
AS egrul.licenses_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    licenses_local,
    cityHash64(entity_ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: branches (Филиалы и представительства)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.branches ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.branches_local ON CLUSTER egrul_cluster
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
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/branches_local',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(version_date)
ORDER BY (company_ogrn, branch_type, branch_kpp, full_address)
SETTINGS index_granularity = 8192;

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.branches ON CLUSTER egrul_cluster
AS egrul.branches_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    branches_local,
    cityHash64(company_ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: companies_okved_additional (Дополнительные ОКВЭД компаний)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.companies_okved_additional ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.companies_okved_additional_local ON CLUSTER egrul_cluster
(
    ogrn        String COMMENT 'ОГРН компании',
    inn         Nullable(String) COMMENT 'ИНН компании',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД',
    updated_at  DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/companies_okved_additional_local',
    '{replica}',
    updated_at
)
ORDER BY (ogrn, okved_code);

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.companies_okved_additional ON CLUSTER egrul_cluster
AS egrul.companies_okved_additional_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    companies_okved_additional_local,
    cityHash64(ogrn)
);

-- ============================================================================
-- ТАБЛИЦА: entrepreneurs_okved_additional (Дополнительные ОКВЭД ИП)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.entrepreneurs_okved_additional ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу
CREATE TABLE IF NOT EXISTS egrul.entrepreneurs_okved_additional_local ON CLUSTER egrul_cluster
(
    ogrnip      String COMMENT 'ОГРНИП предпринимателя',
    inn         String COMMENT 'ИНН предпринимателя',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД',
    updated_at  DateTime64(3) DEFAULT now64(3) COMMENT 'Дата обновления записи'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/entrepreneurs_okved_additional_local',
    '{replica}',
    updated_at
)
ORDER BY (ogrnip, okved_code);

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.entrepreneurs_okved_additional ON CLUSTER egrul_cluster
AS egrul.entrepreneurs_okved_additional_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    entrepreneurs_okved_additional_local,
    cityHash64(ogrnip)
);

-- ============================================================================
-- ТАБЛИЦА: company_history (История изменений)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.company_history ON CLUSTER egrul_cluster;

-- Создаем локальную реплицированную таблицу (БЕЗ ReplacingMergeTree - история неизменяема)
CREATE TABLE IF NOT EXISTS egrul.company_history_local ON CLUSTER egrul_cluster
(
    -- === Идентификаторы ===
    id                      UUID DEFAULT generateUUIDv4() COMMENT 'Уникальный идентификатор записи',
    entity_type             LowCardinality(String) COMMENT 'Тип сущности: company/entrepreneur',
    entity_id               String COMMENT 'ОГРН или ОГРНИП',
    inn                     Nullable(String) COMMENT 'ИНН',

    -- === Сведения о записи ГРН ===
    grn                     String COMMENT 'Государственный регистрационный номер записи',
    grn_date                Date COMMENT 'Дата записи',

    -- === Причина изменения ===
    reason_code             Nullable(String) COMMENT 'Код причины внесения записи',
    reason_description      Nullable(String) COMMENT 'Описание причины',

    -- === Регистрирующий орган ===
    authority_code          Nullable(String) COMMENT 'Код регистрирующего органа',
    authority_name          Nullable(String) COMMENT 'Наименование регистрирующего органа',

    -- === Свидетельство ===
    certificate_series      Nullable(String) COMMENT 'Серия свидетельства',
    certificate_number      Nullable(String) COMMENT 'Номер свидетельства',
    certificate_date        Nullable(Date) COMMENT 'Дата выдачи свидетельства',

    -- === Снимок основных данных на момент изменения ===
    snapshot_full_name      Nullable(String) COMMENT 'Наименование на момент изменения',
    snapshot_status         Nullable(String) COMMENT 'Статус на момент изменения',
    snapshot_address        Nullable(String) COMMENT 'Адрес на момент изменения',
    snapshot_json           Nullable(String) COMMENT 'Полный снимок данных в JSON',

    -- === Метаданные для дедупликации ===
    source_files            Array(String) DEFAULT [] COMMENT 'Список файлов-источников',
    extract_date            Date COMMENT 'Дата выписки из XML',
    file_hash               String DEFAULT '' COMMENT 'Хеш файла для отслеживания источника',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи',
    updated_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата последнего обновления'
)
ENGINE = ReplicatedReplacingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/company_history_local',
    '{replica}',
    updated_at
)
PARTITION BY toYYYYMM(grn_date)
ORDER BY (entity_id, grn_date, grn)
SETTINGS index_granularity = 8192;

-- Добавляем индексы
ALTER TABLE egrul.company_history_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_history_entity_type entity_type TYPE set(10) GRANULARITY 4;

ALTER TABLE egrul.company_history_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_history_inn inn TYPE bloom_filter(0.01) GRANULARITY 4;

ALTER TABLE egrul.company_history_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_history_reason_code reason_code TYPE set(500) GRANULARITY 4;

ALTER TABLE egrul.company_history_local ON CLUSTER egrul_cluster
    ADD INDEX IF NOT EXISTS idx_history_grn_date grn_date TYPE minmax GRANULARITY 4;

-- Создаем Distributed таблицу
CREATE TABLE IF NOT EXISTS egrul.company_history ON CLUSTER egrul_cluster
AS egrul.company_history_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    company_history_local,
    cityHash64(entity_id)
);

-- ============================================================================
-- VIEW ДЛЯ ДЕДУПЛИЦИРОВАННЫХ ДАННЫХ ИСТОРИИ
-- ============================================================================

-- Создаем VIEW для автоматической дедупликации при чтении
-- VIEW создается ON CLUSTER и работает поверх Distributed таблицы
CREATE OR REPLACE VIEW egrul.company_history_view ON CLUSTER egrul_cluster AS
SELECT
    id,
    entity_type,
    entity_id,
    inn,
    grn,
    grn_date,
    reason_code,
    reason_description,
    authority_code,
    authority_name,
    certificate_series,
    certificate_number,
    certificate_date,
    snapshot_full_name,
    snapshot_status,
    snapshot_address,
    snapshot_json,
    source_files,
    extract_date,
    file_hash,
    created_at,
    updated_at
FROM egrul.company_history FINAL
ORDER BY entity_id, grn_date DESC, grn DESC;

-- ============================================================================
-- ТАБЛИЦА: import_log (Лог импорта данных)
-- ============================================================================

-- Удаляем старую таблицу
DROP TABLE IF EXISTS egrul.import_log ON CLUSTER egrul_cluster;

-- Создаем локальную таблицу БЕЗ репликации (логи на каждом узле свои)
CREATE TABLE IF NOT EXISTS egrul.import_log_local ON CLUSTER egrul_cluster
(
    id                      UUID DEFAULT generateUUIDv4(),
    import_date             DateTime64(3) DEFAULT now64(3),
    source_file             String,
    entity_type             LowCardinality(String),
    records_total           UInt64 DEFAULT 0,
    records_inserted        UInt64 DEFAULT 0,
    records_updated         UInt64 DEFAULT 0,
    records_failed          UInt64 DEFAULT 0,
    duration_ms             UInt64 DEFAULT 0,
    status                  LowCardinality(String) DEFAULT 'pending',
    error_message           Nullable(String),
    metadata                Nullable(String) COMMENT 'Дополнительные метаданные в JSON'
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(import_date)
ORDER BY (import_date, source_file)
SETTINGS index_granularity = 8192;

-- Создаем Distributed таблицу (random sharding)
CREATE TABLE IF NOT EXISTS egrul.import_log ON CLUSTER egrul_cluster
AS egrul.import_log_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    import_log_local,
    rand()
);

-- ============================================================================
-- Конец миграции
-- ============================================================================
-- После применения этой миграции необходимо:
-- 1. Загрузить данные заново через скрипты импорта
-- 2. Проверить репликацию: SELECT * FROM system.replicas WHERE database = 'egrul'
-- 3. Проверить распределение данных по шардам
-- ============================================================================
