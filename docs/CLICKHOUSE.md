# ClickHouse Infrastructure для ЕГРЮЛ/ЕГРИП

## Обзор

ClickHouse используется как основное аналитическое хранилище данных ЕГРЮЛ/ЕГРИП. Текущая конфигурация оптимизирована для single-node MVP развертывания.

## Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         ClickHouse                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │  companies  │  │entrepreneurs│  │    company_history      │ │
│  │   (ЕГРЮЛ)   │  │   (ЕГРИП)   │  │  (История изменений)    │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│                                                                  │
│  ┌─────────────────────────┐  ┌────────────────────────────┐   │
│  │    ownership_graph      │  │   Materialized Views       │   │
│  │   (Граф собственности)  │  │  (Агрегированные данные)   │   │
│  └─────────────────────────┘  └────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Быстрый старт

### Запуск ClickHouse

```bash
# Запуск только ClickHouse
docker-compose up -d clickhouse

# Запуск с применением миграций
docker-compose --profile setup up clickhouse-migrations

# Проверка статуса
docker-compose ps clickhouse
```

### Применение миграций

```bash
# Через Docker Compose
docker-compose --profile setup up clickhouse-migrations

# Через скрипт
./infrastructure/scripts/clickhouse-migrate.sh migrate

# Проверка статуса базы
./infrastructure/scripts/clickhouse-migrate.sh status
```

### Подключение к ClickHouse

```bash
# Через Docker
docker exec -it egrul-clickhouse clickhouse-client

# Через clickhouse-client напрямую
clickhouse-client --host localhost --port 9000 --user admin --password admin

# HTTP API
curl 'http://localhost:8123/?query=SELECT%201'
```

## Структура таблиц

### companies (Юридические лица)

| Поле | Тип | Описание |
|------|-----|----------|
| ogrn | String | ОГРН (PRIMARY KEY) |
| inn | String | ИНН |
| kpp | String | КПП |
| full_name | String | Полное наименование |
| short_name | String | Сокращенное наименование |
| status | LowCardinality(String) | Статус (Действующее, Ликвидировано и т.д.) |
| opf_code | String | Код ОПФ |
| opf_name | String | Наименование ОПФ |
| region_code | String | Код региона |
| city | String | Город |
| full_address | String | Полный адрес |
| capital_amount | Decimal(18,2) | Уставный капитал |
| okved_main_code | String | Код основного ОКВЭД |
| okved_additional | Array(String) | Дополнительные ОКВЭД |
| head_last_name, head_first_name | String | ФИО руководителя |
| registration_date | Date | Дата регистрации |
| termination_date | Date | Дата ликвидации |
| version_date | Date | Дата версии данных |

### entrepreneurs (Индивидуальные предприниматели)

| Поле | Тип | Описание |
|------|-----|----------|
| ogrnip | String | ОГРНИП (PRIMARY KEY) |
| inn | String | ИНН |
| last_name | String | Фамилия |
| first_name | String | Имя |
| middle_name | String | Отчество |
| status | LowCardinality(String) | Статус |
| region_code | String | Код региона |
| okved_main_code | String | Код основного ОКВЭД |
| okved_additional | Array(String) | Дополнительные ОКВЭД |
| registration_date | Date | Дата регистрации |
| termination_date | Date | Дата прекращения |

### ownership_graph (Граф собственности)

| Поле | Тип | Описание |
|------|-----|----------|
| owner_type | LowCardinality(String) | Тип владельца |
| owner_id | String | ОГРН/ОГРНИП владельца |
| owner_inn | String | ИНН владельца |
| owner_name | String | Наименование/ФИО |
| target_ogrn | String | ОГРН целевой компании |
| share_percent | Decimal(10,4) | Процент доли |
| is_active | UInt8 | Признак актуальности |

### company_history (История изменений)

| Поле | Тип | Описание |
|------|-----|----------|
| entity_type | LowCardinality(String) | Тип: company/entrepreneur |
| entity_id | String | ОГРН или ОГРНИП |
| grn | String | ГРН записи |
| grn_date | Date | Дата записи |
| reason_code | String | Код причины изменения |
| reason_description | String | Описание причины |

## Индексы и оптимизации

### Типы индексов

- **bloom_filter** - для точного поиска по ИНН, ОГРН, email
- **ngrambf_v1** - для полнотекстового поиска по наименованию
- **set** - для фильтрации по статусу, региону, ОПФ
- **minmax** - для диапазонных запросов по датам и суммам

### Проекции

Созданы проекции для оптимизации частых запросов:
- `proj_companies_by_inn` - поиск по ИНН
- `proj_companies_by_region` - агрегация по регионам
- `proj_companies_by_okved` - агрегация по ОКВЭД
- `proj_ownership_by_owner` - поиск владений

## Materialized Views

### Статистика по регионам

```sql
SELECT 
    region_code,
    region,
    companies_count,
    total_capital,
    avg_capital
FROM egrul.v_stats_companies_by_region
WHERE status = 'Действующее'
ORDER BY companies_count DESC
LIMIT 10;
```

### Топ владельцев

```sql
SELECT 
    owner_name,
    owner_inn,
    companies_count,
    avg_share
FROM egrul.v_top_owners
WHERE owner_type = 'russian_company'
ORDER BY companies_count DESC
LIMIT 20;
```

### Динамика регистраций

```sql
SELECT 
    registration_month,
    registrations_count
FROM egrul.v_registration_dynamics
WHERE entity_type = 'company'
ORDER BY registration_month DESC
LIMIT 24;
```

## Примеры запросов

### Поиск компании по ИНН

```sql
SELECT 
    ogrn, inn, full_name, status, 
    registration_date, capital_amount
FROM egrul.companies
WHERE inn = '7707083893';
```

### Поиск по наименованию (полнотекстовый)

```sql
SELECT 
    ogrn, inn, full_name, status
FROM egrul.companies
WHERE full_name ILIKE '%сбербанк%'
LIMIT 10;
```

### Компании с определенным ОКВЭД в регионе

```sql
SELECT 
    ogrn, inn, full_name, city, capital_amount
FROM egrul.companies
WHERE okved_main_code = '62.01'
  AND region_code = '77'
  AND status = 'Действующее'
ORDER BY capital_amount DESC
LIMIT 100;
```

### Граф владения компанией

```sql
-- Все владельцы компании
SELECT 
    owner_type, owner_name, owner_inn,
    share_percent
FROM egrul.ownership_graph
WHERE target_ogrn = '1027700132195'
  AND is_active = 1
ORDER BY share_percent DESC;

-- Все компании владельца
SELECT 
    target_ogrn, target_name, share_percent
FROM egrul.ownership_graph
WHERE owner_inn = '7707083893'
  AND is_active = 1;
```

### Связанные компании (через общих учредителей)

```sql
SELECT 
    company1_ogrn, company1_name,
    company2_ogrn, company2_name,
    owner_name
FROM egrul.v_related_companies
WHERE company1_ogrn = '1027700132195'
LIMIT 100;
```

## Пользователи

| Пользователь | Пароль | Назначение |
|--------------|--------|------------|
| admin | admin | Администрирование |
| egrul_app | test | Приложение (полный доступ) |
| egrul_import | 123 | Импорт данных |
| egrul_reader | password | Только чтение |
| egrul_api | password123 | API сервис |

⚠️ **Важно**: Замените пароли в production!

## Мониторинг

### Системные таблицы

```sql
-- Размер таблиц
SELECT 
    table,
    formatReadableSize(sum(bytes_on_disk)) AS size,
    sum(rows) AS rows
FROM system.parts
WHERE database = 'egrul'
GROUP BY table
ORDER BY sum(bytes_on_disk) DESC;

-- Активные запросы
SELECT query, elapsed, read_rows, memory_usage
FROM system.processes;

-- История запросов
SELECT 
    query_start_time,
    query_duration_ms,
    read_rows,
    result_rows,
    query
FROM system.query_log
WHERE type = 'QueryFinish'
  AND query_start_time > now() - INTERVAL 1 HOUR
ORDER BY query_start_time DESC
LIMIT 20;
```

### Grafana

Для мониторинга доступна опциональная интеграция с Grafana:

```bash
docker-compose --profile monitoring up -d grafana
```

Grafana доступен на http://localhost:3001 (admin/admin)

## Резервное копирование

```bash
# Создание бэкапа
clickhouse-client --query "BACKUP DATABASE egrul TO Disk('backups', 'egrul_backup.zip')"

# Восстановление
clickhouse-client --query "RESTORE DATABASE egrul FROM Disk('backups', 'egrul_backup.zip')"
```

## Масштабирование

Для production рекомендуется:

1. **Репликация**: Настроить ReplicatedMergeTree для отказоустойчивости
2. **Шардирование**: Распределить данные по нескольким нодам
3. **Keeper**: Использовать ClickHouse Keeper вместо ZooKeeper

Пример конфигурации кластера см. в документации ClickHouse.

## Troubleshooting

### Медленные запросы

```sql
-- Проверка использования индексов
EXPLAIN indexes = 1
SELECT * FROM egrul.companies WHERE inn = '7707083893';
```

### Проблемы с памятью

```sql
-- Настройка лимитов
SET max_memory_usage = 10000000000;
SET max_bytes_before_external_group_by = 5000000000;
```

### Очистка устаревших данных

```sql
-- Принудительное слияние партиций
OPTIMIZE TABLE egrul.companies FINAL;

-- Удаление старых партиций
ALTER TABLE egrul.companies DROP PARTITION '202301';
```

## Версионность данных и поле `extract_date` (ДатаВып)

### Общие принципы

- Таблицы `egrul.companies` и `egrul.entrepreneurs` хранят **только актуальное состояние** сущностей:
  - для каждой компании (`ogrn, inn`) в таблице `companies` хранится одна строка;
  - для каждого ИП (`ogrnip, inn`) в таблице `entrepreneurs` хранится одна строка.
- Исторические изменения (старые состояния) в этих таблицах не сохраняются.

### Источник актуальности — `ДатаВып` из XML

- В XML‑выписках источником «актуальности» записи является поле `ДатаВып`:
  - для юридических лиц — атрибут `ДатаВып` тега `СвЮЛ`;
  - для ИП — атрибут `ДатаВып` тега `СвИП`.
- Парсер извлекает это поле и записывает его в модели как:
  - `EgrulRecord.extract_date` (ЮЛ),
  - `EgripRecord.extract_date` (ИП).
- При записи в Parquet это поле уходит в колонку `extract_date` (строка).
- При импорте в ClickHouse оно преобразуется в тип `Date` и попадает в колонку `extract_date` в таблицах:

```sql
-- companies
extract_date Date DEFAULT toDate('1970-01-01') COMMENT 'Дата выписки (ДатаВып из XML, версия записи)',

-- entrepreneurs
extract_date Date DEFAULT toDate('1970-01-01') COMMENT 'Дата выписки (ДатаВып из XML, версия записи)',
```

### Как ClickHouse выбирает актуальную запись

- Таблицы настроены на движок:

```sql
ENGINE = ReplacingMergeTree(extract_date)
ORDER BY (ogrn, inn);          -- для companies

ENGINE = ReplacingMergeTree(extract_date)
ORDER BY (ogrnip, inn);        -- для entrepreneurs
```

- Это означает:
  - при загрузке нескольких строк с одинаковым ключом сортировки (`ogrn, inn` или `ogrnip, inn`)
    и разными `extract_date` ClickHouse хранит их все до слияния частей;
  - при слиянии партиций `ReplacingMergeTree(extract_date)` оставляет **строку с максимальной датой `extract_date`**.
- Таким образом:
  - можно безопасно загружать данные:
    - из разных периодов (исторические + актуальные),
    - в любом порядке (не обязательно по возрастанию дат),
    - сколько угодно раз (повторные parquet‑файлы);
  - в итоге в таблицах `companies` и `entrepreneurs` всегда будет по одной записи на ОГРН/ОГРНИП — с максимально возможной `ДатаВып` из всех загруженных XML.

### Поведение при отсутствии `ДатаВып`

- Если в исходном XML атрибут `ДатаВып` отсутствует или не парсится:
  - при импорте в ClickHouse используется значение `toDate('1970-01-01')`;
  - такие записи считаются наименее актуальными и будут «перебиты» любой записью с валидной датой выписки.
  
## Дополнительные ОКВЭД и батч‑импорт

### Хранение дополнительных ОКВЭД

- В parquet‑файлах дополнительные виды деятельности хранятся в JSON‑поле `additional_activities`.
- При импорте в ClickHouse это поле попадает в основные таблицы в сыром виде:
  - `egrul.companies.additional_activities Nullable(String)`;
  - `egrul.entrepreneurs.additional_activities Nullable(String)`.
- Для аналитики по дополнительным ОКВЭД используются нормализованные таблицы:
  - `egrul.companies_okved_additional (ogrn, inn, okved_code, okved_name)`;
  - `egrul.entrepreneurs_okved_additional (ogrnip, inn, okved_code, okved_name)`.

Основная загрузка в `companies` / `entrepreneurs` **не разворачивает** JSON сразу в массивы кодов, чтобы не упираться в лимиты памяти ClickHouse на больших объёмах данных.

### Батч‑процедура `make okved-extra`

Для развертывания `additional_activities` в нормализованные таблицы используется отдельная батч‑процедура:

- цель в корневом `Makefile`:

```bash
make okved-extra
```

- под капотом вызывается скрипт `infrastructure/scripts/import-okved-extra.sh`, который:
  - по чанкам (bucket‑ам) проходит по данным в `egrul.companies` и `egrul.entrepreneurs`;
  - читает поле `additional_activities` как JSON‑массив;
  - извлекает из каждого элемента `code` и `name`;
  - вставляет строки в:
    - `egrul.companies_okved_additional`,
    - `egrul.entrepreneurs_okved_additional`;
  - использует настройки типа `OKVED_BUCKETS` и `OKVED_MAX_MEMORY` для контроля объёма памяти.

Процедура идемпотентна относительно основных сущностей: дополнительные ОКВЭД всегда вычисляются из актуального состояния `companies` / `entrepreneurs` и могут быть безопасно перезапущены после полного `make ch-reset && make import`.

### Использование в API

- GraphQL‑API (`services/api-gateway`) при выборке компаний и ИП:
  - загружает основные данные из `egrul.companies` / `egrul.entrepreneurs`;
  - дополнительно подтягивает дополнительные ОКВЭД из:
    - `egrul.companies_okved_additional`,
    - `egrul.entrepreneurs_okved_additional`;
  - объединяет основной и дополнительные виды деятельности в поле `activities` GraphQL‑моделей.

Таким образом, после выполнения `make import` **и** `make okved-extra` API возвращает полный набор видов деятельности (основной + все дополнительные) для компаний и ИП.
