# ClickHouse Infrastructure для ЕГРЮЛ/ЕГРИП

## Обзор

ClickHouse используется как основное аналитическое хранилище данных ЕГРЮЛ/ЕГРИП. Система работает в **кластерном режиме**: 6 нод ClickHouse + 3 ноды ClickHouse Keeper.

> **Single-node режим отключён.** Все таблицы создаются как `_local` (ReplicatedReplacingMergeTree) с Distributed-обёртками поверх.

## Архитектура кластера

```
┌─────────────────────────────────────────────────────────────────────┐
│                     ClickHouse Cluster (egrul_cluster)              │
├──────────────────────────┬──────────────────────────────────────────┤
│  Shard 1 (ОГРН % 3 = 0) │  Shard 2 (ОГРН % 3 = 1)                 │
│  clickhouse-01 (primary) │  clickhouse-03 (primary)                 │
│  clickhouse-02 (replica) │  clickhouse-04 (replica)                 │
├──────────────────────────┴──────────────────────────────────────────┤
│  Shard 3 (ОГРН % 3 = 2)                                             │
│  clickhouse-05 (primary)                                            │
│  clickhouse-06 (replica)                                            │
└─────────────────────────────────────────────────────────────────────┘

Keeper (Raft): keeper-01, keeper-02, keeper-03

Шардирование: по хэшу ОГРН/ОГРНИП (обеспечивает дедупликацию в пределах одного шарда)
```

**Двухуровневая структура таблиц:**
- `egrul.companies_local` — `ReplicatedReplacingMergeTree` на каждой ноде
- `egrul.companies` — `Distributed` поверх `_local`, прозрачный интерфейс для API

## Быстрый старт

### Запуск кластера

```bash
# Запуск всей системы (кластер + все сервисы) — рекомендуется
make up

# Запуск только кластера
make cluster-up

# Проверка состояния кластера
make cluster-verify

# Статус нод
make cluster-ps
```

### Подключение к ClickHouse

```bash
# Через Makefile (подключается к node-01)
make ch-shell

# Напрямую через Docker
docker exec -it egrul-clickhouse-01 clickhouse-client --user egrul_app --password test

# HTTP API
curl 'http://localhost:8123/?query=SELECT%201&user=egrul_app&password=test'
```

### Управление базой данных

```bash
# Пересоздать БД на всех нодах + применить миграции 011-018
make cluster-reset

# Очистить все таблицы (без удаления структуры)
make cluster-truncate

# Импорт данных + заполнение Materialized Views
make cluster-import

# Импорт только дополнительных ОКВЭД
make cluster-import-okved
```

## Структура таблиц

### companies / companies_local (Юридические лица)

| Поле | Тип | Описание |
|------|-----|----------|
| ogrn | String | ОГРН (часть PRIMARY KEY) |
| inn | String | ИНН |
| kpp | String | КПП |
| full_name | String | Полное наименование |
| short_name | String | Сокращенное наименование |
| status_code | String | Код статуса |
| status_name | String | Наименование статуса |
| opf_code | String | Код ОПФ |
| opf_name | String | Наименование ОПФ |
| region_code | String | Код региона |
| region | String | Наименование региона |
| city | String | Город |
| full_address | String | Полный адрес |
| capital_amount | Decimal(18,2) | Уставный капитал |
| okved_main_code | String | Код основного ОКВЭД |
| additional_activities | Nullable(String) | Доп. ОКВЭД (JSON) |
| head_last_name, head_first_name | String | ФИО руководителя |
| registration_date | Date | Дата регистрации |
| termination_date | Nullable(Date) | Дата ликвидации |
| extract_date | Date | Дата выписки (версия записи) |

### entrepreneurs / entrepreneurs_local (Индивидуальные предприниматели)

| Поле | Тип | Описание |
|------|-----|----------|
| ogrnip | String | ОГРНИП (часть PRIMARY KEY) |
| inn | String | ИНН |
| last_name | String | Фамилия |
| first_name | String | Имя |
| middle_name | String | Отчество |
| status_code | String | Код статуса |
| region_code | String | Код региона |
| okved_main_code | String | Код основного ОКВЭД |
| additional_activities | Nullable(String) | Доп. ОКВЭД (JSON) |
| registration_date | Date | Дата регистрации |
| termination_date | Nullable(Date) | Дата прекращения |
| extract_date | Date | Дата выписки (версия записи) |

### ownership_graph / ownership_graph_local (Граф собственности)

| Поле | Тип | Описание |
|------|-----|----------|
| owner_type | LowCardinality(String) | Тип владельца |
| owner_id | String | ОГРН/ОГРНИП владельца |
| owner_inn | String | ИНН владельца |
| owner_name | String | Наименование/ФИО |
| target_ogrn | String | ОГРН целевой компании |
| share_percent | Decimal(10,4) | Процент доли |
| is_active | UInt8 | Признак актуальности |

### company_history / company_history_local (История изменений)

| Поле | Тип | Описание |
|------|-----|----------|
| entity_type | LowCardinality(String) | Тип: company/entrepreneur |
| entity_id | String | ОГРН или ОГРНИП |
| grn | String | ГРН записи |
| grn_date | Date | Дата записи |
| reason_code | String | Код причины изменения |
| reason_description | String | Описание причины |

## Движки таблиц и версионность

### Принцип работы ReplicatedReplacingMergeTree

```sql
-- _local таблицы на каждой ноде
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/egrul/companies', '{replica}', extract_date)
ORDER BY (ogrn, inn);   -- для companies

ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/egrul/entrepreneurs', '{replica}', extract_date)
ORDER BY (ogrnip, inn); -- для entrepreneurs

-- Distributed поверх _local (прозрачный интерфейс)
ENGINE = Distributed(egrul_cluster, egrul, companies_local, cityHash64(ogrn))
```

**Как выбирается актуальная запись:**
- При загрузке нескольких строк с одинаковым ключом сортировки и разными `extract_date` ClickHouse хранит все версии до слияния частей
- При слиянии `ReplicatedReplacingMergeTree(extract_date)` оставляет **строку с максимальным `extract_date`**
- Таким образом, можно безопасно загружать данные из разных периодов в любом порядке

### Поле `extract_date` (ДатаВып)

- Источник: атрибут `ДатаВып` из XML-выписок ЕГРЮЛ/ЕГРИП
- Парсер записывает в колонку `extract_date`
- Если `ДатаВып` отсутствует — используется `toDate('1970-01-01')` (запись будет «перебита» любой записью с валидной датой)

## Индексы и оптимизации

### Типы индексов

- **bloom_filter** — для точного поиска по ИНН, ОГРН
- **ngrambf_v1** — для полнотекстового поиска по наименованию
- **set** — для фильтрации по статусу, региону, ОПФ
- **minmax** — для диапазонных запросов по датам и суммам

## Materialized Views (статистика)

Materialized Views хранят агрегаты для дашборда аналитики:
- `stats_companies_by_region` — компании по регионам и статусу
- `stats_entrepreneurs_by_region` — ИП по регионам и статусу
- `stats_registrations_by_month` — динамика регистраций
- `stats_terminations_by_month` — динамика ликвидаций

Заполнение MV выполняется при импорте:
```bash
make cluster-fill-mv    # заполнение из существующих данных
make cluster-import     # импорт + автоматическое заполнение MV
```

## Примеры запросов

### Поиск компании по ИНН

```sql
SELECT ogrn, inn, full_name, status_name, registration_date, capital_amount
FROM egrul.companies
WHERE inn = '7707083893';
```

### Поиск по наименованию (полнотекстовый)

```sql
SELECT ogrn, inn, full_name, status_name
FROM egrul.companies
WHERE full_name ILIKE '%сбербанк%'
LIMIT 10;
```

### Компании с определенным ОКВЭД в регионе

```sql
SELECT ogrn, inn, full_name, city, capital_amount
FROM egrul.companies
WHERE okved_main_code = '62.01'
  AND region_code = '77'
  AND termination_date IS NULL
ORDER BY capital_amount DESC
LIMIT 100;
```

### Граф владения компанией

```sql
-- Все владельцы компании
SELECT owner_type, owner_name, owner_inn, share_percent
FROM egrul.ownership_graph
WHERE target_ogrn = '1027700132195' AND is_active = 1
ORDER BY share_percent DESC;
```

### Статистика по регионам

```sql
SELECT region_code, status, sumMerge(count) AS cnt
FROM egrul.stats_companies_by_region
WHERE status = 'active'
GROUP BY region_code, status
ORDER BY cnt DESC
LIMIT 20;
```

## Пользователи кластера

| Пользователь | Пароль | Назначение |
|--------------|--------|------------|
| egrul_app | test | API Gateway (профиль: default) |
| egrul_import | 123 | Импорт данных (профиль: import_profile) |
| egrul_api | password123 | API с аналитикой (профиль: analytics) |
| egrul_reader | password | Только чтение (профиль: readonly) |

> ⚠️ **Важно**: Замените пароли в production! Конфиг пользователей: `infrastructure/docker/clickhouse-cluster/shared/users.xml`

## Дополнительные ОКВЭД и батч-импорт

### Хранение дополнительных ОКВЭД

- В parquet-файлах дополнительные виды деятельности хранятся в JSON-поле `additional_activities`
- При импорте попадают в основные таблицы в сыром виде: `additional_activities Nullable(String)`
- Для аналитики используются нормализованные таблицы:
  - `egrul.companies_okved_additional (ogrn, inn, okved_code, okved_name)`
  - `egrul.entrepreneurs_okved_additional (ogrnip, inn, okved_code, okved_name)`

### Батч-процедура

```bash
make okved-extra          # Развернуть additional_activities в нормализованные таблицы
make cluster-import-okved # То же, с явными credentials для кластера
```

После `make cluster-reset && make cluster-import` API возвращает полный набор видов деятельности.

## Версионность данных (жизненный цикл)

См. подробное описание в [DATA_LIFECYCLE.md](DATA_LIFECYCLE.md).

```bash
make cluster-detect-changes  # Детектирование изменений по версиям
make cluster-optimize        # OPTIMIZE FINAL (с подтверждением)
make cluster-optimize-force  # OPTIMIZE FINAL (без подтверждения, для автоматизации)
make cluster-optimize-stats  # Статистика дублей без очистки
```

## Мониторинг

### Системные таблицы

```sql
-- Размер таблиц на кластере
SELECT table, formatReadableSize(sum(bytes_on_disk)) AS size, sum(rows) AS rows
FROM clusterAllReplicas('egrul_cluster', system.parts)
WHERE database = 'egrul'
GROUP BY table ORDER BY sum(bytes_on_disk) DESC;

-- Активные запросы
SELECT query, elapsed, read_rows, memory_usage
FROM system.processes;

-- История запросов за последний час
SELECT query_start_time, query_duration_ms, read_rows, result_rows, query
FROM system.query_log
WHERE type = 'QueryFinish' AND query_start_time > now() - INTERVAL 1 HOUR
ORDER BY query_start_time DESC LIMIT 20;
```

### Grafana + Prometheus

```bash
make monitoring-up
# Grafana: http://localhost:3001 (admin/admin)
# Prometheus: http://localhost:9090
```

## Резервное копирование

```bash
# Создание backup в MinIO
make cluster-backup

# Восстановление из backup
make cluster-restore BACKUP_NAME=backup_YYYYMMDD_HHMMSS
```

Backup хранится в MinIO bucket `backups`. Конфиг: `infrastructure/docker/clickhouse-cluster/shared/backup_disk.xml`.

## Troubleshooting

### Медленные запросы

```sql
-- Проверка использования индексов
EXPLAIN indexes = 1
SELECT * FROM egrul.companies WHERE inn = '7707083893';
```

### Проблемы с памятью

```sql
SET max_memory_usage = 10000000000;
SET max_bytes_before_external_group_by = 5000000000;
```

### Проверка репликации

```sql
SELECT database, table, is_leader, total_replicas, active_replicas
FROM system.replicas WHERE database = 'egrul';
```

### Ноды кластера недоступны

```bash
make cluster-verify    # Быстрая проверка
make cluster-test      # Полный набор тестов
make cluster-logs      # Просмотр логов кластера
```
