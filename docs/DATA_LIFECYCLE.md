# Управление жизненным циклом данных

Документ описывает стратегию хранения версий данных и автоматическую очистку в системе ЕГРЮЛ/ЕГРИП.

## Архитектура хранения

### ReplacingMergeTree

Таблицы используют движок `ReplicatedReplacingMergeTree(extract_date)`:
- Хранит **все версии** записей до слияния
- При слиянии оставляет версию с **максимальной** `extract_date`
- Слияние происходит **асинхронно** в фоне
- `OPTIMIZE FINAL` принудительно запускает слияние

### Текущее состояние

По умолчанию в системе:
- ✅ История хранится (слияние отложено)
- ✅ Можно детектировать изменения между версиями
- ⚠️ Дубли накапливаются (артефакты парсинга + реальные изменения)
- ⚠️ Размер БД растет с каждым импортом

## Workflow импорта и детектирования

### 1. Импорт данных

```bash
make cluster-import
```

**Что происходит:**

**Шаг 1: Импорт**
1. Парсинг XML → Parquet (если нужно)
2. INSERT данных в ClickHouse (все версии сохраняются)

**Шаг 2: Детектирование изменений** (если `AUTO_DETECT_CHANGES=true`)
1. Получает список OGRN с множественными версиями (разные `extract_date`)
2. Вызывает `change-detection-service` через HTTP
3. Сравнивает старую и новую версии
4. Создает записи в `company_changes`
5. Отправляет события в Kafka → уведомления по email

**Шаг 3: Автоматическая очистка** (если `AUTO_OPTIMIZE_AFTER_DETECT=true`)
1. `OPTIMIZE TABLE ... ON CLUSTER ... FINAL` для всех таблиц
2. Удаление старых версий (остается только последняя)
3. Удаление артефактов парсинга

### 2. Ручная очистка старых версий

**Когда использовать:**
- ✅ Если `AUTO_OPTIMIZE_AFTER_DETECT=false` (Вариант 2: скользящее окно)
- ✅ Периодически через cron (раз в неделю/месяц)
- ✅ По требованию для освобождения места

**Команды:**

```bash
# Показать статистику (сколько дублей)
make cluster-optimize-stats

# Очистка с подтверждением
make cluster-optimize

# Очистка без подтверждения (для cron)
make cluster-optimize-force
```

**Что происходит:**
1. Статистика ДО очистки (дубли, версии)
2. `OPTIMIZE TABLE ... ON CLUSTER ... FINAL` для всех таблиц
3. Удаление старых версий (остается только последняя)
4. Статистика ПОСЛЕ очистки

**ВАЖНО:** Детектирование всегда запускается автоматически ДО очистки (если `AUTO_DETECT_CHANGES=true`)

## Стратегии управления данными

### Вариант 1: Минимальная история (рекомендуется для production)

**Частота очистки:** После каждого импорта (автоматически)

```bash
# В .env или .env.production
AUTO_DETECT_CHANGES=true
AUTO_OPTIMIZE_AFTER_DETECT=true

# Запуск импорта (очистка произойдет автоматически)
make cluster-import
```

**Плюсы:**
- ✅ Минимальный размер БД
- ✅ Нет артефактов парсинга
- ✅ Полностью автоматизировано

**Минусы:**
- ❌ История удаляется после обработки
- ❌ Нельзя пересчитать изменения за прошлые периоды

### Вариант 2: Скользящее окно (рекомендуется для dev/staging)

**Частота очистки:** Раз в неделю (через cron)

```bash
# В .env или .env.development
AUTO_DETECT_CHANGES=true
AUTO_OPTIMIZE_AFTER_DETECT=false  # очистка вручную

# Cron: каждое воскресенье в 03:00
0 3 * * 0 cd /path/to/egrul && make cluster-optimize-force
```

**Плюсы:**
- ✅ История доступна несколько дней
- ✅ Можно пересчитать изменения за неделю
- ✅ Разумный баланс размера и гибкости
- ✅ Гибкий контроль над очисткой

**Минусы:**
- ⚠️ Размер БД растет в течение недели
- ⚠️ Требует настройки cron

### Вариант 3: Полная история

**Частота очистки:** Никогда или очень редко (раз в год)

```bash
# В .env
AUTO_DETECT_CHANGES=true
AUTO_OPTIMIZE_AFTER_DETECT=false  # никогда не очищать автоматически

# Очистка только при необходимости (вручную)
make cluster-optimize
```

**Плюсы:**
- ✅ Вся история доступна всегда
- ✅ Можно анализировать тренды

**Минусы:**
- ❌ Большой размер БД
- ❌ Медленнее запросы (больше данных для сканирования)
- ❌ Много артефактов парсинга

## Настройка автоматизации

### Автоматическое детектирование и очистка

**Рекомендуемая конфигурация для production:**

```bash
# .env.production
AUTO_DETECT_CHANGES=true              # Автоматически детектировать изменения
AUTO_OPTIMIZE_AFTER_DETECT=true       # Автоматически очищать старые версии
```

После каждого `make cluster-import`:
1. ✅ Импорт данных в ClickHouse
2. ✅ Автоматическое детектирование изменений
3. ✅ Отправка событий в Kafka → Email уведомления
4. ✅ Автоматическая очистка старых версий

**Рекомендуемая конфигурация для dev/staging:**

```bash
# .env.development
AUTO_DETECT_CHANGES=true              # Автоматически детектировать изменения
AUTO_OPTIMIZE_AFTER_DETECT=false      # Очистка вручную (через cron или по требованию)
```

### Отключение автоматического детектирования

Если нужно запускать детектирование вручную:

```bash
# В .env
AUTO_DETECT_CHANGES=false

# Затем вручную:
make cluster-detect-changes
```

### Настройка cron для периодической очистки

Для **Варианта 2** (скользящее окно) с `AUTO_OPTIMIZE_AFTER_DETECT=false`:

```bash
# Редактируем crontab
crontab -e

# Добавляем задачу (раз в неделю в воскресенье в 03:00)
0 3 * * 0 cd /path/to/egrul && make cluster-optimize-force >> /var/log/egrul-cleanup.log 2>&1
```

### Мониторинг размера БД

```bash
# Размер таблиц
docker exec egrul-clickhouse-01 clickhouse-client --query "
SELECT
    table,
    formatReadableSize(sum(bytes)) as size,
    formatReadableQuantity(sum(rows)) as rows,
    formatReadableQuantity(sum(rows) / uniqExact(rows)) as avg_versions
FROM system.parts
WHERE database = 'egrul' AND active
GROUP BY table
ORDER BY sum(bytes) DESC
"

# Статистика дублей
make cluster-optimize-stats
```

## Рекомендации

### Для разработки/тестирования:
- ✅ Используйте **Вариант 2** (скользящее окно)
- ✅ Конфигурация: `AUTO_DETECT_CHANGES=true`, `AUTO_OPTIMIZE_AFTER_DETECT=false`
- ✅ Очистка раз в неделю через cron или по требованию
- ✅ Возможность пересчета изменений за период

### Для production:
- ✅ Используйте **Вариант 1** (минимальная история) - рекомендуется
- ✅ Конфигурация: `AUTO_DETECT_CHANGES=true`, `AUTO_OPTIMIZE_AFTER_DETECT=true`
- ✅ Полностью автоматический workflow
- ✅ Мониторинг размера БД
- ✅ Алерты при превышении порогов

**Альтернатива для production:**
- Вариант 2 с еженедельной очисткой через cron
- Конфигурация: `AUTO_DETECT_CHANGES=true`, `AUTO_OPTIMIZE_AFTER_DETECT=false`

### Важные моменты:

1. **Детектирование ВСЕГДА запускается ДО очистки**

   С `AUTO_OPTIMIZE_AFTER_DETECT=true`:
   ```bash
   make cluster-import
   # Автоматически: импорт → детектирование → очистка
   ```

   С `AUTO_OPTIMIZE_AFTER_DETECT=false`:
   ```bash
   make cluster-import  # импорт + детектирование
   make cluster-optimize-force  # ручная очистка когда нужно
   ```

2. **Мониторьте размер БД**
   - Настройте алерты при превышении 80% дискового пространства
   - Проверяйте статистику дублей перед очисткой

3. **Backup перед очисткой**
   - OPTIMIZE необратимо удаляет старые версии
   - Сделайте backup если нужна полная история:
   ```bash
   make cluster-backup
   make cluster-optimize-force
   ```

## Переменные окружения

```bash
# Автоматическое детектирование изменений после импорта
AUTO_DETECT_CHANGES=true  # default: true

# Автоматическая очистка старых версий после детектирования
# true - минимальный размер БД, нет истории (Вариант 1)
# false - скользящее окно, еженедельная очистка через cron (Вариант 2, рекомендуется)
AUTO_OPTIMIZE_AFTER_DETECT=false  # default: false для dev, true для production

# УСТАРЕВШЕЕ: Используйте AUTO_OPTIMIZE_AFTER_DETECT вместо этого
# OPTIMIZE_AFTER_IMPORT=false
```

## Команды быстрого доступа

```bash
# Импорт с автоматическим детектированием и опциональной очисткой
make cluster-import
# Использует: AUTO_DETECT_CHANGES, AUTO_OPTIMIZE_AFTER_DETECT из .env

# Детектирование вручную (если AUTO_DETECT_CHANGES=false)
make cluster-detect-changes

# Статистика дублей и версий
make cluster-optimize-stats

# Очистка с подтверждением (если AUTO_OPTIMIZE_AFTER_DETECT=false)
make cluster-optimize

# Очистка без подтверждения (для cron)
make cluster-optimize-force
```

## Быстрая настройка

**Минимальная история (production):**
```bash
# В .env.production
AUTO_DETECT_CHANGES=true
AUTO_OPTIMIZE_AFTER_DETECT=true

# Просто запускаем импорт - все остальное автоматически
make cluster-import
```

**Скользящее окно (dev/staging):**
```bash
# В .env.development
AUTO_DETECT_CHANGES=true
AUTO_OPTIMIZE_AFTER_DETECT=false

# Импорт с детектированием
make cluster-import

# Очистка по расписанию (cron)
0 3 * * 0 make cluster-optimize-force
```
