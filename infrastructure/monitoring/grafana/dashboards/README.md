# Grafana Dashboards для EGRUL/EGRIP

## Быстрый старт - Импорт готовых дашбордов

### Через Grafana UI (рекомендуется)

1. Откройте Grafana: http://localhost:3001 (admin/admin)
2. Перейдите: Dashboards → New → Import
3. Введите ID дашборда и нажмите Load
4. Выберите datasource: **Prometheus**
5. Нажмите Import

### Рекомендуемые дашборды

#### ClickHouse мониторинг
- **ID: 14192** - ClickHouse Cluster Overview (comprehensive)
- **ID: 13606** - ClickHouse Query Performance

#### Go приложения
- **ID: 10826** - Go Metrics (runtime, goroutines, memory)
- **ID: 6671** - Go Processes (detailed process metrics)

#### Docker контейнеры
- **ID: 193** - Docker monitoring (cAdvisor)
- **ID: 14282** - cAdvisor Prometheus

#### Prometheus
- **ID: 3662** - Prometheus 2.0 Overview
- **ID: 11074** - Node Exporter Full

## Custom дашборды для EGRUL

После импорта базовых дашбордов создайте специфичные для проекта:

### 1. EGRUL Services Overview
**Файл**: `egrul-overview.json`

Панели:
- Services Status (up/down)
- Total RPS по сервисам
- Error Rate (%)
- P95/P99 Latency
- Memory Usage
- CPU Usage (из cAdvisor)

**Основные запросы**:
```promql
# Services Up/Down
up{job=~"api-gateway|search-service|change-detection-service|notification-service"}

# RPS по сервисам
sum(rate(http_requests_total[5m])) by (service)

# Error Rate
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
/
sum(rate(http_requests_total[5m])) by (service) * 100

# P95 Latency
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))

# P99 Latency
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))
```

### 2. Business Metrics Dashboard
**Файл**: `business-metrics.json`

Метрики:
- `egrul_company_searches_total` - Поиски компаний
- `egrul_company_exports_total` - Экспорты данных
- `egrul_graphql_queries_total` - GraphQL операции
- `egrul_changes_detected_total` - Обнаруженные изменения
- `egrul_emails_sent_total` - Отправленные email
- `egrul_kafka_consumer_lag` - Kafka consumer lag

### 3. API Gateway Dashboard
Детализация по GraphQL операциям, cache hit rate, DB query performance

### 4. ClickHouse Dashboard
Используйте готовый ID 14192 и добавьте:
- Количество строк в таблицах
- Query duration по операциям
- Replication lag

## Создание custom панелей

### Пример: RPS панель

1. New Dashboard → Add visualization
2. Data source: **Prometheus**
3. Query: `sum(rate(http_requests_total[5m])) by (service)`
4. Visualization: Time series
5. Title: "Requests Per Second"
6. Legend: `{{service}}`

### Пример: Error Rate панель

```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
/
sum(rate(http_requests_total[5m])) by (service)
* 100
```

Visualization: Gauge
Unit: Percent (0-100)
Thresholds: Green < 1%, Yellow < 5%, Red >= 5%

## Экспорт созданных дашбордов

После создания custom дашборда:

1. Dashboard settings (⚙️) → JSON Model
2. Скопировать JSON
3. Сохранить в `infrastructure/monitoring/grafana/dashboards/`
4. Перезапустить Grafana: `docker compose restart grafana`

Дашборды автоматически загрузятся из `/etc/grafana/dashboards` при следующем запуске.

## Troubleshooting

**Дашборд не показывает данные:**
- Проверьте Datasource: Prometheus должен быть доступен
- Проверьте что метрики собираются: http://localhost:9090/targets
- Попробуйте запрос в Prometheus UI: http://localhost:9090/graph

**Панели показывают "No Data":**
- Увеличьте time range (правый верхний угол)
- Проверьте фильтры в queries
- Убедитесь что сервисы работают и генерируют трафик

## Примеры запросов для тестирования

```bash
# Генерация трафика для метрик
curl http://localhost:8080/graphql -d '{"query":"{ company(ogrn:\"1234567890123\") { fullName } }"}'

# Проверка метрик
curl http://localhost:9090/metrics | grep http_requests_total
```
