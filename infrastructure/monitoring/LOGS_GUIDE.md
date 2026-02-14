# ЕГРЮЛ/ЕГРИП Logs Guide

## Доступ к логам

**Grafana Logs Dashboard:** http://localhost:3001/d/egrul-logs

**Логин:** admin / admin

## Возможности Dashboard

### 1. Фильтрация по сервисам
В верхней части dashboard есть dropdown **"service"** для выбора сервисов:
- `All` - логи всех сервисов
- `api-gateway` - только API Gateway
- `change-detection` - только Change Detection Service
- `notification` - только Notification Service
- `search-service` - только Search Service

Можно выбрать несколько сервисов одновременно.

### 2. Поиск по содержимому
Текстовое поле **"search"** позволяет фильтровать логи по содержимому:
- `GraphQL` - найти все логи содержащие "GraphQL"
- `error` - найти логи с ошибками
- `POST /api/` - найти POST запросы к API
- Пустое значение - показать все логи

### 3. Панели Dashboard

**Logs by Level** (Bar Chart)
- Распределение логов по уровням (info, warn, error, debug)
- Цветовая кодировка: info=синий, warn=оранжевый, error=красный, debug=зелёный

**Total Errors** (Stat)
- Общее количество ошибок за выбранный период
- Красный цвет если есть ошибки

**Total Logs** (Stat)
- Общее количество логов за период

**Logs Rate by Service** (Time Series)
- График частоты логов по каждому сервису во времени

**Errors & Warnings Over Time** (Time Series)
- График ошибок и предупреждений во времени

**Service Logs** (Logs Panel)
- Основная панель с логами
- Показывает полный текст каждого лога
- JSON форматирование
- Можно раскрыть для просмотра всех полей

**Errors Only** (Logs Panel)
- Только логи с level="error"
- Быстрый доступ к ошибкам

**Top 10 Endpoints** (Pie Chart)
- 10 наиболее часто вызываемых endpoint'ов
- Процентное распределение запросов

## Примеры использования

### Поиск ошибок за последний час
1. Установите временной диапазон: `Last 1 hour`
2. Выберите сервис: `All`
3. Посмотрите панель **"Errors Only"**

### Анализ производительности конкретного сервиса
1. Выберите сервис: `api-gateway`
2. Установите временной диапазон: `Last 6 hours`
3. Посмотрите **"Top 10 Endpoints"** для определения горячих точек
4. Посмотрите **"Service Logs"** для анализа latency

### Поиск конкретной операции
1. Выберите сервис: `api-gateway`
2. В поле search введите: `GraphQL`
3. Посмотрите **"Service Logs"**

### Мониторинг в реальном времени
1. Включите auto-refresh: кнопка справа вверху, выберите `5s` или `10s`
2. Выберите временной диапазон: `Last 5 minutes`
3. Наблюдайте обновление логов в режиме реального времени

## Структура JSON логов

Каждый лог содержит следующие поля:

```json
{
  "level": "info",              // Уровень: debug, info, warn, error
  "timestamp": "2026-02-12...", // ISO 8601 timestamp
  "caller": "file.go:123",      // Файл и строка вызова
  "msg": "HTTP request",        // Сообщение
  "service": "api-gateway",     // Имя сервиса
  "request_id": "...",          // ID запроса (для трейсинга)
  "method": "GET",              // HTTP метод
  "path": "/health",            // Путь запроса
  "status": 200,                // HTTP статус
  "latency": 0.001,             // Время выполнения (секунды)
  "remote_addr": "...",         // IP клиента
  "user_agent": "..."           // User Agent
}
```

## LogQL Queries (для Explore)

Базовые запросы:
```logql
{service="api-gateway"}                          # Все логи сервиса
{service="api-gateway",level="error"}            # Только ошибки
{service=~"api-gateway|search-service"}          # Несколько сервисов
```

Фильтрация по содержимому:
```logql
{service="api-gateway"} |= "GraphQL"             # Содержит "GraphQL"
{service="api-gateway"} |~ "GET|POST"            # Regex поиск
{service="api-gateway"} != "health"              # Не содержит "health"
```

JSON парсинг и фильтрация:
```logql
{service="api-gateway"} | json                   # Парсить JSON
{service="api-gateway"} | json | status >= 400   # HTTP статус >= 400
{service="api-gateway"} | json | latency > 1.0   # Latency > 1 сек
{service="api-gateway"} | json | method="POST"   # POST запросы
```

Агрегация:
```logql
sum(count_over_time({service="api-gateway"}[5m]))                    # Количество логов за 5 мин
sum by (level) (count_over_time({service="api-gateway"}[1h]))        # Группировка по level
rate({service="api-gateway"}[5m])                                    # Частота логов/сек
topk(10, sum by (path) (count_over_time({service="api-gateway"} | json [$__range])))  # Top 10 paths
```

## Troubleshooting

**Логи не появляются:**
1. Проверьте что сервисы запущены: `docker compose ps`
2. Проверьте что Promtail собирает логи: `make promtail-logs`
3. Проверьте что Loki работает: `curl http://localhost:3100/ready`
4. Сгенерируйте тестовый трафик: `for i in {1..20}; do curl -s http://localhost:8080/health > /dev/null; done`

**Dashboard показывает "No data":**
1. Проверьте временной диапазон - возможно выбран период без активности
2. Проверьте фильтры service и search
3. Убедитесь что datasource "Loki" доступен в Grafana

**Медленная загрузка dashboard:**
1. Уменьшите временной диапазон (например, с "Last 24h" на "Last 1h")
2. Выберите конкретный сервис вместо "All"
3. Увеличьте интервал обновления (или отключите auto-refresh)

## Полезные ссылки

- **Grafana Logs Dashboard:** http://localhost:3001/d/egrul-logs
- **Grafana Explore (для ad-hoc queries):** http://localhost:3001/explore
- **Loki API:** http://localhost:3100
- **Loki Labels:** http://localhost:3100/loki/api/v1/labels
- **LogQL Documentation:** https://grafana.com/docs/loki/latest/query/

## Makefile команды

```bash
make monitoring-up      # Запустить весь стек мониторинга
make monitoring-logs    # Просмотр логов мониторинга
make loki-logs          # Логи Loki
make promtail-logs      # Логи Promtail
make loki-labels        # Доступные labels в Loki
make loki-query QUERY='{service="api-gateway"}'  # Запрос логов через API
```
