# TODO: Миграция Materialized Views для кластера

## Статус: ОТЛОЖЕНО

**Дата создания:** 2026-01-20
**Приоритет:** Средний
**Ответственный:** Backend team

---

## Текущее состояние

### ✅ Что реализовано в кластерной миграции 011

**Базовые таблицы (все с Replicated*MergeTree + Distributed):**
- `companies_local` / `companies`
- `entrepreneurs_local` / `entrepreneurs`
- `founders_local` / `founders`
- `company_history_local` / `company_history` + VIEW `company_history_view`
- `licenses_local` / `licenses`
- `branches_local` / `branches`
- `ownership_graph_local` / `ownership_graph`
- `companies_okved_additional_local` / `companies_okved_additional`
- `entrepreneurs_okved_additional_local` / `entrepreneurs_okved_additional`
- `import_log_local` / `import_log`

**Все поля из миграций 004, 007, 009, 010:**
- ✅ `additional_activities` (JSON для ОКВЭД)
- ✅ Детализированные поля адреса для ИП
- ✅ ReplacingMergeTree для дедупликации связанных таблиц
- ✅ Поля `company_share_percent`, `company_share_nominal`
- ✅ Поля `old_reg_number`, `old_reg_date`, `old_reg_authority`

### ❌ Что пропущено

**Миграция 003_materialized_views.sql НЕ адаптирована для кластера!**

Отсутствуют Materialized Views для аналитики в реальном времени:

1. **Статистика по регионам:**
   - `stats_companies_by_region` + `mv_stats_companies_by_region`
   - `stats_entrepreneurs_by_region` + `mv_stats_entrepreneurs_by_region`

2. **Статистика по ОКВЭД:**
   - `stats_companies_by_okved` + `mv_stats_companies_by_okved`
   - `stats_entrepreneurs_by_okved` + `mv_stats_entrepreneurs_by_okved`

3. **Статистика по ОПФ:**
   - `stats_companies_by_opf` + `mv_stats_companies_by_opf`

4. **Статистика регистраций по месяцам:**
   - `stats_registrations_by_month`
   - `mv_stats_company_registrations`
   - `mv_stats_entrepreneur_registrations`

5. **Статистика прекращений по месяцам:**
   - `stats_terminations_by_month`
   - `mv_stats_company_terminations`
   - `mv_stats_entrepreneur_terminations`

6. **Статистика собственников:**
   - `stats_top_owners` + `mv_stats_top_owners`
   - `stats_ownership_by_type` + `mv_stats_ownership_by_type`

7. **Статистика изменений:**
   - `stats_changes_by_month` + `mv_stats_changes_by_month`

8. **Итоговая статистика:**
   - `stats_summary` (таблица с общей статистикой)

**Итого: 11 агрегирующих таблиц + 14 Materialized Views**

---

## План доработки

### Миграция 012_materialized_views_cluster.sql

**Цель:** Адаптировать все Materialized Views из миграции 003 для работы в кластере.

### Шаг 1: Анализ использования (перед началом)

**Проверить использование в коде:**
```bash
# API Gateway
grep -r "stats_" services/api-gateway/
grep -r "mv_stats_" services/api-gateway/

# Frontend
grep -r "stats_" frontend/src/
```

**Выяснить:**
- Какие MV активно используются в UI/API
- Какие запросы к статистике выполняются
- Нужны ли все 14 MV или можно начать с критичных

### Шаг 2: Адаптация для кластера

**Для каждой агрегирующей таблицы создать:**

1. **_local версию** с ReplicatedAggregatingMergeTree:
```sql
CREATE TABLE egrul.stats_companies_by_region_local ON CLUSTER egrul_cluster
(
    region_code LowCardinality(String),
    region Nullable(String),
    status LowCardinality(String),
    count AggregateFunction(count),
    total_capital AggregateFunction(sum, Nullable(Decimal(18, 2))),
    ...
)
ENGINE = ReplicatedAggregatingMergeTree(
    '/clickhouse/tables/{cluster}/{shard}/stats_companies_by_region_local',
    '{replica}'
)
PARTITION BY region_code
ORDER BY (region_code, status);
```

2. **Distributed обертку:**
```sql
CREATE TABLE egrul.stats_companies_by_region ON CLUSTER egrul_cluster
AS egrul.stats_companies_by_region_local
ENGINE = Distributed(
    egrul_cluster,
    egrul,
    stats_companies_by_region_local,
    cityHash64(region_code)
);
```

3. **Materialized View на каждом шарде:**
```sql
CREATE MATERIALIZED VIEW egrul.mv_stats_companies_by_region_local ON CLUSTER egrul_cluster
TO egrul.stats_companies_by_region_local
AS SELECT
    region_code,
    any(region) AS region,
    status,
    countState() AS count,
    sumState(capital_amount) AS total_capital,
    ...
FROM egrul.companies_local  -- ВАЖНО: читаем из _local
WHERE region_code IS NOT NULL AND region_code != ''
GROUP BY region_code, status;
```

### Шаг 3: Ключевые моменты кластерной реализации

**Шардирование:**
- Использовать те же ключи шардирования, что и у базовых таблиц
- `region_code` для региональной статистики
- `okved_main_code` для статистики по ОКВЭД
- `owner_id` для статистики собственников

**Репликация:**
- ReplicatedAggregatingMergeTree для автоматической репликации агрегатов
- RF=2 (как и у базовых таблиц)

**Materialized Views:**
- Создавать с суффиксом `_local` и привязывать к `_local` таблицам
- MV будет работать локально на каждом шарде
- Distributed таблица автоматически агрегирует результаты со всех шардов

**Чтение данных:**
```sql
-- Пример запроса к агрегированным данным
SELECT
    region_code,
    any(region) as region,
    status,
    countMerge(count) as total_count,
    sumMerge(total_capital) as total_capital
FROM egrul.stats_companies_by_region
GROUP BY region_code, status;
```

### Шаг 4: Порядок создания

1. Создать все `*_local` таблицы с ReplicatedAggregatingMergeTree
2. Создать все Distributed обертки
3. Создать все Materialized Views (они начнут работать с новых данных)
4. Опционально: заполнить исторические данные через `INSERT INTO ... SELECT ...`

### Шаг 5: Тестирование

**После миграции проверить:**
```bash
# 1. Созданы ли все таблицы на всех нодах
docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 \
  --query "SHOW TABLES FROM egrul LIKE 'stats_%' FORMAT Pretty"

# 2. Работает ли репликация
docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 \
  --query "SELECT * FROM system.replicas WHERE database = 'egrul' AND table LIKE 'stats_%'"

# 3. Наполняются ли MV (после импорта данных)
docker exec egrul-clickhouse-01 clickhouse-client --user egrul_import --password 123 \
  --query "SELECT count() FROM egrul.stats_companies_by_region"
```

---

## Технические детали

### Отличия от single-node версии

| Аспект | Single-node | Cluster |
|--------|-------------|---------|
| Engine агрегатов | AggregatingMergeTree | ReplicatedAggregatingMergeTree |
| Суффикс таблиц | Нет | `_local` + Distributed обертка |
| MV источник | `companies`, `entrepreneurs` | `companies_local`, `entrepreneurs_local` |
| Шардирование | Нет | По ключу (region_code, okved_code, и т.д.) |
| Репликация | Нет | RF=2 через ZooKeeper/Keeper |

### Потенциальные проблемы

1. **Производительность при импорте:**
   - MV увеличивают нагрузку при массовых INSERT
   - Решение: отключать MV на время больших импортов через `DETACH`

2. **Консистентность между шардами:**
   - Данные могут быть временно несогласованы
   - Решение: использовать `FINAL` в критичных запросах

3. **Миграция исторических данных:**
   - Новые MV работают только с новыми данными
   - Решение: запустить пересчет через `INSERT INTO ... SELECT ...` после создания

### Оценка трудозатрат

- **Написание миграции:** 4-6 часов
- **Тестирование на тестовых данных:** 2-3 часа
- **Применение к продакшену:** 1 час
- **Мониторинг после запуска:** 1-2 дня

**Итого:** ~1-2 рабочих дня

---

## Приоритет и риски

### Приоритет: СРЕДНИЙ

**Можно отложить, если:**
- Статистика пока не критична для бизнеса
- Можно считать статистику по требованию (без MV)
- Планируется доработка аналитики в будущем

**Нужно сделать срочно, если:**
- UI активно использует эндпоинты статистики
- Запросы без MV тормозят на больших объемах
- Нужна статистика в реальном времени

### Риски

1. **Низкий риск:**
   - Отсутствие MV не ломает основной функционал
   - Данные компаний/ИП доступны через обычные запросы

2. **Средний риск:**
   - Медленная статистика может ухудшить UX
   - Пересчет больших объемов данных может нагрузить кластер

3. **Минимизация рисков:**
   - Начать с критичных MV (региональная статистика)
   - Добавлять остальные постепенно
   - Мониторить производительность кластера

---

## Примеры использования (после реализации)

### Запрос региональной статистики
```sql
SELECT
    region_code,
    any(region) as region_name,
    status,
    countMerge(count) as companies_count,
    sumMerge(total_capital) as total_capital_rub
FROM egrul.stats_companies_by_region
WHERE region_code IN ('77', '78', '50')
GROUP BY region_code, status
ORDER BY companies_count DESC;
```

### Топ-10 регионов по количеству компаний
```sql
SELECT
    region_code,
    any(region) as region_name,
    countMerge(count) as total_companies
FROM egrul.stats_companies_by_region
GROUP BY region_code
ORDER BY total_companies DESC
LIMIT 10;
```

### Динамика регистраций по месяцам
```sql
SELECT
    month,
    countMerge(count) as new_companies
FROM egrul.stats_registrations_by_month
WHERE entity_type = 'company'
  AND month >= '2024-01-01'
GROUP BY month
ORDER BY month;
```

---

## Чек-лист перед началом работы

- [ ] Проверить использование MV в API Gateway (`grep -r "stats_"`)
- [ ] Проверить использование MV во Frontend (`grep -r "stats_"`)
- [ ] Определить приоритет MV (какие критичны, какие можно отложить)
- [ ] Подготовить тестовые данные для проверки
- [ ] Убедиться что кластер готов к дополнительной нагрузке
- [ ] Создать резервную копию текущих данных
- [ ] Запланировать окно для миграции (если критично)

---

## Связанные файлы

- **Базовая миграция:** `infrastructure/migrations/clickhouse/single-node/003_materialized_views.sql`
- **Кластерная миграция:** `infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql`
- **Будущая миграция:** `infrastructure/migrations/clickhouse/cluster/012_materialized_views_cluster.sql` (создать)

---

## Контакты для уточнений

- **Вопросы по архитектуре:** Backend team lead
- **Вопросы по использованию:** Frontend team / Product owner
- **Вопросы по производительности:** DevOps team
