# Система отслеживания изменений контрагентов

## Обзор

Система для отслеживания изменений в данных компаний и ИП из ЕГРЮЛ/ЕГРИП с отправкой Email уведомлений. Пользователи могут подписываться на изменения конкретных организаций и получать уведомления о статусе, руководителях, учредителях, адресе, капитале и видах деятельности.

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                      EGRUL/EGRIP System                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐         ┌────────────────────────┐        │
│  │  XML Import  │─────1──>│ Change Detection       │        │
│  │  (Existing)  │         │ Service                │        │
│  └──────────────┘         └────────────┬───────────┘        │
│                                        │                     │
│                                        │ 2. Kafka events    │
│                                        ▼                     │
│                           ┌────────────────────────┐        │
│                           │ Notification Service   │        │
│                           │ - Email Channel (SMTP) │        │
│                           └────────────┬───────────┘        │
│                                        │                     │
│  ┌──────────────┐                     │ 3. Email           │
│  │  Frontend    │                     └───────────────>     │
│  │  (Watchlist) │<──────GraphQL──────API Gateway            │
│  └──────────────┘                    (PostgreSQL)           │
│                                                               │
│  Данные:                                                     │
│  - PostgreSQL: subscriptions, notification_log               │
│  - ClickHouse: company_changes, entrepreneur_changes         │
│  - Kafka: company-changes, entrepreneur-changes topics       │
└─────────────────────────────────────────────────────────────┘
```

## Компоненты

### 1. API Gateway (GraphQL API)

**Файлы:**
- `services/api-gateway/internal/graph/subscription.graphqls` - GraphQL схема
- `services/api-gateway/internal/graph/subscription.resolvers.go` - Resolvers
- `services/api-gateway/internal/repository/postgresql/subscription_repo.go` - PostgreSQL repository

**GraphQL API:**

```graphql
# Создание подписки
mutation CreateSubscription {
  createSubscription(
    email: "user@example.com"
    entityType: COMPANY
    entityId: "1234567890123"
    entityName: "ООО ПРИМЕР"
    changeFilters: {
      status: true
      director: true
      founders: true
      address: true
      capital: true
      activities: true
    }
    notificationChannels: {
      email: true
    }
  ) {
    id
    userEmail
    isActive
  }
}

# Получение подписок пользователя
query MySubscriptions {
  mySubscriptions(email: "user@example.com") {
    id
    entityType
    entityId
    entityName
    isActive
    changeFilters {
      status
      director
      founders
      address
      capital
      activities
    }
  }
}

# Проверка наличия подписки
query HasSubscription {
  hasSubscription(
    email: "user@example.com"
    entityType: COMPANY
    entityId: "1234567890123"
  )
}

# Удаление подписки
mutation DeleteSubscription {
  deleteSubscription(id: "uuid")
}

# Приостановка/возобновление подписки
mutation ToggleSubscription {
  toggleSubscription(id: "uuid", isActive: false) {
    id
    isActive
  }
}
```

**База данных (PostgreSQL):**

Схема: `subscriptions`

Таблицы:
- `entity_subscriptions` - подписки пользователей
- `notification_log` - история отправленных уведомлений

### 2. Change Detection Service

**Порт:** 8082
**Язык:** Go
**Файлы:** `services/change-detection-service/`

**Функции:**
- Сравнение старых и новых версий данных компаний/ИП
- Классификация изменений (status, director, founders, address, capital, activities)
- Запись изменений в ClickHouse (`company_changes`, `entrepreneur_changes`)
- Отправка событий в Kafka

**HTTP API:**

```bash
# Запуск детектирования изменений
POST http://localhost:8082/detect
Content-Type: application/json

{
  "entity_type": "company",
  "entity_ids": ["1234567890123", "9876543210987"]
}
```

**ClickHouse таблицы:**

```sql
-- Таблица изменений компаний
CREATE TABLE company_changes (
    id UUID,
    ogrn String,
    change_type String,  -- status, director, founder_added, address, capital, activity
    field_name String,
    old_value String,
    new_value String,
    detected_at DateTime,
    is_significant UInt8
) ENGINE = MergeTree()
ORDER BY (ogrn, detected_at);

-- Аналогично для entrepreneur_changes
```

**Интеграция с импортом:**

После импорта Parquet в ClickHouse вызывается Change Detection Service:

```bash
# В infrastructure/scripts/import-data.sh
curl -X POST http://change-detection-service:8082/detect \
  -H "Content-Type: application/json" \
  -d '{"entity_type": "company", "entity_ids": ["1234567890123", ...]}'
```

### 3. Notification Service

**Порт:** 8083
**Язык:** Go
**Файлы:** `services/notification-service/`

**Функции:**
- Kafka consumer для событий изменений (topics: `company-changes`, `entrepreneur-changes`)
- Чтение подписок из PostgreSQL
- Фильтрация изменений по `change_filters`
- Отправка Email уведомлений через SMTP
- Запись логов в `notification_log`

**Email шаблон:**

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Изменение в данных компании</title>
</head>
<body>
    <h2>Обнаружено изменение: {{.CompanyName}}</h2>
    <p><strong>ОГРН:</strong> {{.OGRN}}</p>
    <p><strong>Тип изменения:</strong> {{.ChangeType}}</p>
    <p><strong>Старое значение:</strong> {{.OldValue}}</p>
    <p><strong>Новое значение:</strong> {{.NewValue}}</p>
    <p><strong>Дата изменения:</strong> {{.DetectedAt}}</p>

    <a href="{{.CompanyURL}}">Перейти к карточке компании</a>

    <hr>
    <p><small>Чтобы отписаться, перейдите в настройки профиля.</small></p>
</body>
</html>
```

**SMTP конфигурация (переменные окружения):**

```bash
SMTP_HOST=smtp.company.ru
SMTP_PORT=587
SMTP_USERNAME=notifications@company.ru
SMTP_PASSWORD=CHANGE_ME_IN_SECRETS
SMTP_FROM=noreply@egrul.company.ru
SMTP_FROM_NAME=ЕГРЮЛ/ЕГРИП Мониторинг
SMTP_TLS=true
```

Для разработки используется **MailHog** (профиль: tools):
```bash
SMTP_HOST=mailhog
SMTP_PORT=1025
# Web UI: http://localhost:8025
```

### 4. Frontend (Next.js)

**Файлы:**
- `frontend/src/components/subscriptions/subscription-form.tsx` - форма создания подписки
- `frontend/src/components/subscriptions/subscriptions-list.tsx` - список подписок
- `frontend/src/app/(dashboard)/watchlist/page.tsx` - страница управления подписками
- `frontend/src/store/user-store.ts` - Zustand store для email пользователя
- `frontend/src/lib/api/subscription-hooks.ts` - React Query hooks

**UI компоненты:**

1. **Кнопка "Отслеживать изменения"** в карточках компаний/ИП:
   - Показывает статус подписки (активна/неактивна)
   - Открывает форму создания/редактирования подписки

2. **Форма подписки (SubscriptionForm):**
   - Ввод email
   - Выбор типов изменений (чекбоксы)
   - Select all / Deselect all

3. **Список подписок (SubscriptionsList):**
   - Карточки с информацией о подписке
   - Кнопки: Приостановить/Возобновить, Удалить
   - Ссылка на карточку компании/ИП

4. **Страница Watchlist:**
   - Email prompt (если не задан)
   - Статистика подписок (всего, активных, приостановлено)
   - Список всех подписок
   - Инструкции по использованию

**Временное решение для аутентификации:**
- Email сохраняется в localStorage через Zustand persist
- Используется как идентификатор пользователя
- В будущем будет заменено на полную систему аутентификации

## Kafka Topics

**company-changes:**
```json
{
  "id": "uuid",
  "ogrn": "1234567890123",
  "change_type": "director",
  "field_name": "ceo",
  "old_value": "{\"fullName\": \"Иванов Иван Иванович\"}",
  "new_value": "{\"fullName\": \"Петров Петр Петрович\"}",
  "detected_at": "2024-01-15T10:30:00Z",
  "is_significant": true
}
```

**entrepreneur-changes:**
Аналогичная структура с `ogrnip` вместо `ogrn`.

## Docker Services

### Запуск всей инфраструктуры

```bash
# Запуск с Kafka и Notification services
make docker-up --profile full

# Или через docker compose
docker compose --profile full up -d
```

**Профили:**
- `default` - базовые сервисы (без Kafka, MinIO)
- `full` - все сервисы включая notification system
- `tools` - UI инструменты (Adminer, RedisInsight, MailHog)

### Проверка статуса

```bash
# Проверка всех сервисов
docker compose ps

# Логи notification сервисов
docker compose logs -f change-detection-service notification-service

# Kafka topics
docker compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

## Миграции

### PostgreSQL (subscriptions schema)

```bash
# Автоматически применяются при первом запуске postgres контейнера
# Файл: infrastructure/migrations/postgresql/001_subscriptions.sql

# Или вручную:
docker compose exec postgres psql -U postgres -d egrul -f /migrations/001_subscriptions.sql
```

### ClickHouse (change_tracking)

```bash
# Single-node режим
make ch-migrate
# Файл: infrastructure/migrations/clickhouse/single-node/012_change_tracking.sql

# Кластерный режим
make cluster-reset
# Файл: infrastructure/migrations/clickhouse/cluster/011_distributed_cluster.sql
```

## Environment Variables

Ключевые переменные в `.env`:

```bash
# Change Detection Service
CHANGE_DETECTION_SERVICE_PORT=8082
CHANGE_DETECTION_SERVICE_LOG_LEVEL=info

# Notification Service
NOTIFICATION_SERVICE_PORT=8083
NOTIFICATION_SERVICE_LOG_LEVEL=info

# SMTP Configuration
SMTP_HOST=smtp.company.ru
SMTP_PORT=587
SMTP_USERNAME=notifications@company.ru
SMTP_PASSWORD=CHANGE_ME_IN_SECRETS
SMTP_FROM=noreply@egrul.company.ru
SMTP_FROM_NAME=ЕГРЮЛ/ЕГРИП Мониторинг
SMTP_TLS=true

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_COMPANY_CHANGES_TOPIC=company-changes
KAFKA_ENTREPRENEUR_CHANGES_TOPIC=entrepreneur-changes
KAFKA_CONSUMER_GROUP=notification-service-group

# PostgreSQL для subscriptions
POSTGRES_SUBSCRIPTION_SCHEMA=subscriptions
```

## Testing End-to-End

### 1. Инициализация

```bash
# Запустить полную инфраструктуру
docker compose --profile full up -d

# Проверить что все сервисы работают
docker compose ps
# Должны быть: postgres, clickhouse, kafka,
# change-detection-service, notification-service, api-gateway, frontend

# Проверить Kafka topics
docker compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
# Должны быть: company-changes, entrepreneur-changes

# Проверить PostgreSQL схему
docker compose exec postgres psql -U postgres -d egrul -c "\dt subscriptions.*"
# Должны быть: entity_subscriptions, notification_log
```

### 2. Создание подписки через UI

```bash
# Открыть frontend
open http://localhost:3000

# Перейти на страницу watchlist
open http://localhost:3000/watchlist

# Ввести email: test@example.com

# Найти компанию через поиск
# Перейти на карточку компании

# Нажать "Отслеживать изменения"
# Выбрать типы изменений (статус, руководитель)
# Нажать "Подписаться"
```

### 3. Проверка подписки в БД

```bash
# Проверить что подписка создана
docker compose exec postgres psql -U postgres -d egrul -c \
  "SELECT * FROM subscriptions.entity_subscriptions WHERE user_email = 'test@example.com';"
```

### 4. Импорт обновлённых данных

```bash
# Используйте XML файл с изменениями в той же компании
make import INPUT=./data/updated
```

### 5. Проверка детектирования изменений

```bash
# Проверить что изменения детектированы в ClickHouse
docker compose exec clickhouse clickhouse-client --query \
  "SELECT * FROM company_changes WHERE ogrn = '...' ORDER BY detected_at DESC LIMIT 5;"
```

### 6. Проверка Kafka событий

```bash
# Просмотр событий в topic
docker compose exec kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic company-changes \
  --from-beginning \
  --max-messages 5
```

### 7. Проверка логов Notification Service

```bash
# Должны увидеть: "Sending email notification to test@example.com"
docker compose logs -f notification-service | grep "Sending email"
```

### 8. Проверка email (MailHog)

```bash
# Открыть MailHog Web UI
open http://localhost:8025

# Должно быть письмо с уведомлением об изменении
```

### 9. Проверка notification_log

```bash
# Проверить что уведомление записано
docker compose exec postgres psql -U postgres -d egrul -c \
  "SELECT * FROM subscriptions.notification_log WHERE status = 'SENT' ORDER BY sent_at DESC LIMIT 5;"
```

## Troubleshooting

### Проблема: Email не отправляются

```bash
# 1. Проверить логи Notification Service
docker compose logs notification-service | grep -i error

# 2. Проверить SMTP подключение
docker compose exec notification-service telnet smtp.company.ru 587

# 3. Проверить переменные окружения
docker compose exec notification-service env | grep SMTP
```

### Проблема: Изменения не детектируются

```bash
# 1. Проверить Change Detection Service
docker compose logs change-detection-service

# 2. Проверить что данные импортированы
docker compose exec clickhouse clickhouse-client --query \
  "SELECT count() FROM companies WHERE ogrn = '...';"

# 3. Проверить вызов детектирования
curl -X POST http://localhost:8082/detect \
  -H "Content-Type: application/json" \
  -d '{"entity_type": "company", "entity_ids": ["1234567890123"]}'
```

### Проблема: Kafka lag

```bash
# Проверить consumer lag
docker compose exec kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --describe \
  --group notification-service-group
```

## Security & Performance

### Idempotency

Notification Service проверяет `notification_log` перед отправкой email по `change_event_id` для предотвращения дублей.

### Rate Limiting

Email отправка имеет retry механизм (3 попытки с exponential backoff: 1s, 5s, 30s).

### Батчирование

При массовом импорте уведомления группируются по пользователю (максимум N уведомлений в час).

### Мониторинг

Рекомендуется настроить alerts:
- Kafka consumer lag > 1000
- Email failure rate > 10%
- Change Detection Service response time > 5s

## Future Enhancements

### Фаза 2: Расширение каналов
- WebSocket real-time уведомления (GraphQL Subscriptions)
- Telegram Bot для уведомлений

### Фаза 3: Улучшения UX
- Батчирование уведомлений (digest раз в день)
- Умные фильтры ("только значимые изменения")
- Аналитика подписок (dashboards в Grafana)

### Фаза 4: Масштабирование
- Кластер Kafka (вместо single-node)
- PostgreSQL replication (master-slave)
- Rate limiting на API Gateway
