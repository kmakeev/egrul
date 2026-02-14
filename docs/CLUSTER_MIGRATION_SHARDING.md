# Миграция шардирования кластера: region_code → OGRN/OGRNIP

## Проблема

При шардировании по `region_code` возникали дубликаты записей, которые не удалялись через `OPTIMIZE FINAL`:

1. **Основная проблема (99%):** Компании/ИП с одинаковым OGRN/OGRNIP, но **разными region_code** попадали на разные шарды
   - 208 компаний имели разные region_code
   - 355 ИП имели разные region_code
   - ReplacingMergeTree работает только внутри одного шарда → дубли не объединялись

2. **Второстепенная проблема (1%):** Даже с одинаковым region_code записи иногда попадали на разные шарды

## Решение

Изменено шардирование:
- **Было:** `cityHash64(coalesce(region_code, '00'))`
- **Стало:** `cityHash64(ogrn)` для companies, `cityHash64(ogrnip)` для entrepreneurs

**Преимущества:**
- ✅ Все версии одной сущности гарантированно попадают на один шард
- ✅ ReplacingMergeTree корректно дедуплицирует записи
- ✅ OPTIMIZE FINAL правильно работает
- ✅ Не зависит от region_code (который может меняться или быть некорректным)

**Недостатки:**
- ⚠️ Теряется географическое распределение данных
- ⚠️ Требуется полное пересоздание кластера и реимпорт данных

## План миграции

### Шаг 1: Подготовка

```bash
# Убедитесь, что у вас есть актуальные Parquet файлы
ls -lh output/*.parquet

# Проверьте текущее состояние кластера
make cluster-verify
make cluster-optimize-stats
```

### Шаг 2: Backup (ОПЦИОНАЛЬНО, если нужны данные)

```bash
# Создать backup текущих данных
make cluster-backup
```

### Шаг 3: Полное пересоздание кластера

```bash
# Остановка всех сервисов
make down

# Пересоздание кластера с новым шардированием
make cluster-reset

# Запуск всех сервисов
make up

# Проверка кластера
make cluster-verify
```

### Шаг 4: Реимпорт данных

```bash
# Полный импорт с автоматическим детектированием изменений
make cluster-import

# Статистика после импорта
make cluster-optimize-stats
```

### Шаг 5: Проверка результатов

```bash
# Проверка отсутствия дублей
make cluster-optimize-stats

# Должно показать:
# Companies: Duplicates = 0 (или минимум)
# Entrepreneurs: Duplicates = 0 (или минимум)
```

### Шаг 6: Проверка подписок (если есть)

```bash
# Подписки хранятся в PostgreSQL и НЕ удаляются при очистке ClickHouse
# Проверьте, что подписки на месте
docker compose exec postgres psql -U postgres -d egrul -c "SELECT * FROM subscriptions.entity_subscriptions LIMIT 5;"

# Загрузите модифицированный файл ИП для проверки
# (если у вас есть test/subs/modified/VO_RIGFO_0000_9965_20240615_old_version.xml)
make parser-run INPUT=test/subs/modified OUTPUT=output/modified
make cluster-import

# Проверьте логи notification-service
docker compose logs -f notification-service
```

## Ожидаемый результат

После миграции:
- **0 дублей** по OGRN/OGRNIP (все версии объединены)
- **Быстрое детектирование изменений** (данные на одном шарде)
- **Корректная работа подписок** (уведомления отправляются)
- **Эффективное использование дискового пространства**

## Откат (если что-то пошло не так)

```bash
# Восстановление из backup
make cluster-restore BACKUP_NAME=backup_20240131_xxx

# ИЛИ: Вернуться к старой версии миграции
git checkout HEAD~1 infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql
make cluster-reset
make cluster-import
```

## Технические детали

### Изменения в миграции 011

**Файл:** `infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql`

**Изменено:**

```sql
-- БЫЛО (companies):
ENGINE = Distributed(egrul_cluster, egrul, companies_local, cityHash64(coalesce(region_code, '00')))

-- СТАЛО:
ENGINE = Distributed(egrul_cluster, egrul, companies_local, cityHash64(ogrn))
```

```sql
-- БЫЛО (entrepreneurs):
ENGINE = Distributed(egrul_cluster, egrul, entrepreneurs_local, cityHash64(coalesce(region_code, '00')))

-- СТАЛО:
ENGINE = Distributed(egrul_cluster, egrul, entrepreneurs_local, cityHash64(ogrnip))
```

### Почему OGRN/OGRNIP - правильный ключ шардирования

1. **OGRN/ОГРНИП уникальны и постоянны** - не меняются в течение жизни компании/ИП
2. **Первые 2 цифры OGRN указывают регион** - частично сохраняется географическое распределение
3. **Гарантия консистентности** - все версии одной сущности всегда на одном шарде
4. **Эффективная дедупликация** - ReplacingMergeTree может правильно работать

## FAQ

**Q: Потеряются ли подписки после пересоздания кластера?**
A: Нет! Подписки хранятся в PostgreSQL (таблица `subscriptions.entity_subscriptions`), а не в ClickHouse.

**Q: Можно ли сделать миграцию без простоя?**
A: Нет, требуется полное пересоздание Distributed таблиц. Рекомендуется делать в нерабочее время.

**Q: Сколько времени займет реимпорт?**
A: Зависит от объема данных. Для ~100k компаний и ~300k ИП - примерно 10-30 минут.

**Q: Что делать с данными в production?**
A:
1. Создать backup (`make cluster-backup`)
2. Запланировать окно обслуживания
3. Выполнить миграцию
4. Проверить результаты
5. При проблемах - откатиться к backup

**Q: Изменится ли API для клиентов?**
A: Нет, API останется прежним. Изменения только внутри кластера ClickHouse.

## Проверка после миграции

```bash
# 1. Проверка отсутствия дублей
docker exec egrul-clickhouse-01 clickhouse-client --query "
SELECT
    'Companies' as entity,
    count(*) as total_records,
    uniqExact(ogrn) as unique_ids,
    count(*) - uniqExact(ogrn) as duplicates
FROM egrul.companies
UNION ALL
SELECT
    'Entrepreneurs',
    count(*),
    uniqExact(ogrnip),
    count(*) - uniqExact(ogrnip)
FROM egrul.entrepreneurs
FORMAT PrettyCompact"

# Ожидаемый результат:
# Companies:     total_records ≈ unique_ids, duplicates = 0
# Entrepreneurs: total_records ≈ unique_ids, duplicates = 0

# 2. Проверка правильности шардирования
docker exec egrul-clickhouse-01 clickhouse-client --query "
SELECT
    ogrn,
    count() as versions,
    uniqExact(_shard_num) as shards_count
FROM egrul.companies
WHERE ogrn IN (SELECT ogrn FROM egrul.companies GROUP BY ogrn HAVING count() > 1 LIMIT 10)
GROUP BY ogrn
FORMAT PrettyCompact"

# Ожидаемый результат:
# shards_count = 1 для всех OGRN (все версии на одном шарде)
```
