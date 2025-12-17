-- ============================================================================
-- ЕГРЮЛ/ЕГРИП ClickHouse Schema Migration
-- Version: 001_initial_schema
-- Description: Создание основных таблиц для хранения данных ЕГРЮЛ/ЕГРИП
-- ============================================================================

-- Создание базы данных
CREATE DATABASE IF NOT EXISTS egrul
    ENGINE = Atomic
    COMMENT 'База данных ЕГРЮЛ/ЕГРИП - Единый государственный реестр юридических лиц и индивидуальных предпринимателей';

-- ============================================================================
-- Таблица: companies (Юридические лица - ЕГРЮЛ)
-- ============================================================================
-- Соответствует структуре EgrulRecord из parser/src/models/egrul.rs
-- и схеме parquet из parser/src/output/parquet_writer.rs

CREATE TABLE IF NOT EXISTS egrul.companies
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
    email                   Nullable(String) COMMENT 'Адрес электронной почты',
    
    -- === Уставный капитал ===
    capital_amount          Nullable(Decimal(18, 2)) COMMENT 'Размер уставного капитала',
    capital_currency        Nullable(String) DEFAULT 'RUB' COMMENT 'Валюта капитала',
    
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
ENGINE = ReplacingMergeTree(extract_date)
PARTITION BY toYYYYMM(version_date)
ORDER BY (ogrn, inn)
PRIMARY KEY ogrn
SETTINGS index_granularity = 8192,
         ttl_only_drop_parts = 1;

-- ============================================================================
-- Таблица: entrepreneurs (Индивидуальные предприниматели - ЕГРИП)
-- ============================================================================
-- Соответствует структуре EgripRecord из parser/src/models/egrip.rs

CREATE TABLE IF NOT EXISTS egrul.entrepreneurs
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
    full_address            Nullable(String) COMMENT 'Полный адрес',
    fias_id                 Nullable(String) COMMENT 'ФИАС код',
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
ENGINE = ReplacingMergeTree(extract_date)
PARTITION BY toYYYYMM(version_date)
ORDER BY (ogrnip, inn)
PRIMARY KEY ogrnip
SETTINGS index_granularity = 8192,
         ttl_only_drop_parts = 1;

-- ============================================================================
-- Таблица: companies_okved_additional (Дополнительные ОКВЭД компаний)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.companies_okved_additional
(
    ogrn        String COMMENT 'ОГРН компании',
    inn         Nullable(String) COMMENT 'ИНН компании',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД'
)
ENGINE = MergeTree()
ORDER BY (ogrn, okved_code);

-- ============================================================================
-- Таблица: entrepreneurs_okved_additional (Дополнительные ОКВЭД ИП)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.entrepreneurs_okved_additional
(
    ogrnip      String COMMENT 'ОГРНИП предпринимателя',
    inn         String COMMENT 'ИНН предпринимателя',
    okved_code  String COMMENT 'Код дополнительного ОКВЭД',
    okved_name  Nullable(String) COMMENT 'Наименование дополнительного ОКВЭД'
)
ENGINE = MergeTree()
ORDER BY (ogrnip, okved_code);

-- ============================================================================
-- Таблица: company_history (История изменений)
-- ============================================================================
-- Хранит все записи ГРН (государственный регистрационный номер записи)

CREATE TABLE IF NOT EXISTS egrul.company_history
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
    
    -- === Метаданные ===
    source_file             Nullable(String) COMMENT 'Источник данных',
    created_at              DateTime64(3) DEFAULT now64(3) COMMENT 'Дата создания записи'
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(grn_date)
ORDER BY (entity_id, grn_date, grn)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: ownership_graph (Граф собственности)
-- ============================================================================
-- Связи владения для построения графа собственности

CREATE TABLE IF NOT EXISTS egrul.ownership_graph
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
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(version_date)
ORDER BY (target_ogrn, owner_type, owner_id, owner_name)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: founders (Учредители - денормализованная)
-- ============================================================================
-- Детальная информация об учредителях для быстрых запросов

CREATE TABLE IF NOT EXISTS egrul.founders
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
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(version_date)
ORDER BY (company_ogrn, founder_type, founder_inn, founder_name)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: licenses (Лицензии)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.licenses
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
    activity                Nullable(String) COMMENT 'Вид деятельности',
    start_date              Nullable(Date) COMMENT 'Дата начала действия',
    end_date                Nullable(Date) COMMENT 'Дата окончания действия',
    authority               Nullable(String) COMMENT 'Лицензирующий орган',
    status                  LowCardinality(String) DEFAULT '' COMMENT 'Статус лицензии',
    
    -- === Метаданные ===
    version_date            Date DEFAULT today(),
    created_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(version_date)
ORDER BY (entity_ogrn, license_number)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: branches (Филиалы и представительства)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.branches
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
    full_address            Nullable(String),
    
    -- === ГРН ===
    grn                     Nullable(String),
    grn_date                Nullable(Date),
    
    -- === Метаданные ===
    version_date            Date DEFAULT today(),
    created_at              DateTime64(3) DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(version_date)
ORDER BY (company_ogrn, branch_type, branch_kpp)
SETTINGS index_granularity = 8192;

-- ============================================================================
-- Таблица: import_log (Лог импорта данных)
-- ============================================================================

CREATE TABLE IF NOT EXISTS egrul.import_log
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

