# Система мониторинга и Observability для EGRUL/EGRIP

## Статус реализации

### ✅ Выполнено (Фаза 1 - частично)

1. **Структура директорий** - создана полная иерархия для всех 3 фаз
2. **Prometheus конфигурация**:
   - `prometheus/prometheus.yml` - конфиг с 11 targets (ClickHouse кластер, Go сервисы, cAdvisor)
   - `prometheus/rules/alerts.yml` - 17 alert rules (availability, performance, resources, business)
3. **Shared пакет для метрик**:
   - `services/shared/pkg/observability/metrics/prometheus.go` - HTTP, DB, Cache метрики
   - `services/shared/pkg/observability/metrics/middleware.go` - Chi middleware
   - `services/shared/pkg/observability/metrics/business.go` - Business метрики для всех сервисов
4. **Docker Compose** - Prometheus volumes раскомментированы в `docker-compose.cluster.yml`

### 📋 Оставшиеся задачи Фазы 1

#### 1. Добавить cAdvisor в docker-compose.yml

Добавить перед секцией volumes:

```yaml
  # ==================== Мониторинг ====================

  # cAdvisor - метрики Docker контейнеров
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:v0.47.0
    container_name: egrul-cadvisor
    hostname: cadvisor
    privileged: true
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:ro
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
      - /dev/disk/:/dev/disk:ro
    ports:
      - "8085:8080"
    networks:
      - egrul-network
    profiles:
      - monitoring
      - full
    restart: unless-stopped
```

#### 2. Добавить Prometheus client в services/shared/go.mod

```bash
cd services/shared
go get github.com/prometheus/client_golang@v1.19.0
```

#### 3. Добавить /metrics endpoint в API Gateway

Файл: `services/api-gateway/cmd/server/main.go`

После строки 56 (после инициализации logger), добавить:

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
    sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)

// После создания роутера r:

// Prometheus metrics на отдельном порту
go func() {
    metricsRouter := chi.NewRouter()
    metricsRouter.Handle("/metrics", promhttp.Handler())
    logger.Info("Starting metrics server", zap.String("addr", ":9090"))
    if err := http.ListenAndServe(":9090", metricsRouter); err != nil {
        logger.Fatal("Failed to start metrics server", zap.Error(err))
    }
}()

// Добавить middleware для метрик (ПЕРЕД другими middleware)
r.Use(sharedMetrics.HTTPMiddleware("api-gateway"))
```

Также добавить в go.mod:
```bash
cd services/api-gateway
go get github.com/prometheus/client_golang@v1.19.0
```

#### 4. Аналогично для остальных 3 сервисов

- **Search Service**: порт 9091, service name "search-service"
- **Change Detection**: порт 9092, service name "change-detection"
- **Notification**: порт 9093, service name "notification"

#### 5. Создать Grafana provisioning конфиги

**Файл**: `infrastructure/monitoring/grafana/provisioning/datasources.yml`

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: "15s"
      queryTimeout: "60s"

  - name: ClickHouse
    type: grafana-clickhouse-datasource
    access: proxy
    url: http://clickhouse-01:8123
    database: egrul
    editable: true
    jsonData:
      server: clickhouse-01
      port: 9000
      username: egrul_app
    secureJsonData:
      password: test
```

**Файл**: `infrastructure/monitoring/grafana/provisioning/dashboards.yml`

```yaml
apiVersion: 1

providers:
  - name: 'EGRUL Dashboards'
    orgId: 1
    folder: 'EGRUL/EGRIP'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    allowUiUpdates: true
    options:
      path: /etc/grafana/dashboards
```

#### 6. Обновить Grafana в docker-compose.yml

Добавить volumes:

```yaml
grafana:
  image: grafana/grafana:11.3.0
  volumes:
    - grafana_data:/var/lib/grafana
    - ./infrastructure/monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
    - ./infrastructure/monitoring/grafana/dashboards:/etc/grafana/dashboards:ro
  depends_on:
    - prometheus
  # остальное без изменений
```

#### 7. Создать 7 Grafana дашбордов

В директории `infrastructure/monitoring/grafana/dashboards/` создать JSON файлы:

1. `overview.json` - общий обзор (можно использовать готовые шаблоны с Grafana.com)
2. `api-gateway.json` - метрики API Gateway
3. `search-service.json` - метрики поиска
4. `change-detection.json` - метрики детектора изменений
5. `notification.json` - метрики уведомлений
6. `clickhouse-cluster.json` - метрики ClickHouse кластера
7. `business-metrics.json` - бизнес метрики

**Примеры готовых дашбордов** (можно импортировать):
- ClickHouse Overview: https://grafana.com/grafana/dashboards/14192
- Go Processes: https://grafana.com/grafana/dashboards/6671
- Docker cAdvisor: https://grafana.com/grafana/dashboards/193

#### 8. Обновить .env.example

Добавить в конец файла:

```bash
# ==============================================================================
# Prometheus Monitoring
# ==============================================================================
PROMETHEUS_PORT=9090
PROMETHEUS_RETENTION=30d

# Metrics endpoints для Go сервисов
API_GATEWAY_METRICS_PORT=9090
SEARCH_SERVICE_METRICS_PORT=9091
CHANGE_DETECTION_METRICS_PORT=9092
NOTIFICATION_SERVICE_METRICS_PORT=9093

# cAdvisor
CADVISOR_PORT=8085
```

#### 9. Обновить Makefile

Добавить targets для мониторинга:

```makefile
# ==============================================================================
# Мониторинг
# ==============================================================================

monitoring-up: ## Запуск Prometheus + Grafana
	@echo "$(CYAN)📊 Запуск мониторинга...$(NC)"
	@$(DOCKER_COMPOSE) --profile monitoring up -d prometheus grafana cadvisor
	@echo "$(GREEN)✅ Мониторинг запущен!$(NC)"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3001 (admin/admin)"
	@echo "  - cAdvisor: http://localhost:8085"

monitoring-down: ## Остановка мониторинга
	@$(DOCKER_COMPOSE) --profile monitoring down

prometheus-reload: ## Перезагрузка конфигурации Prometheus
	@echo "$(CYAN)🔄 Перезагрузка Prometheus...$(NC)"
	@curl -X POST http://localhost:9090/-/reload
	@echo "$(GREEN)✅ Prometheus перезагружен$(NC)"

grafana-open: ## Открыть Grafana UI
	@open http://localhost:3001
```

#### 10. Тестирование Фазы 1

```bash
# 1. Запуск кластера (если еще не запущен)
make cluster-up

# 2. Запуск мониторинга
make monitoring-up

# 3. Проверка Prometheus targets
open http://localhost:9090/targets
# Должны быть UP: clickhouse-cluster (6), clickhouse-keeper (3),
# api-gateway, search-service, change-detection-service, notification-service,
# cadvisor, prometheus (всего 15 targets)

# 4. Проверка метрик от Go сервисов
curl http://localhost:9090/metrics | grep http_requests_total

# 5. Проверка Grafana
open http://localhost:3001
# Логин: admin/admin
# Проверить datasources: Configuration → Data Sources
# Импортировать/создать дашборды

# 6. Проверка алертов
open http://localhost:9090/alerts
# Должны быть loaded: 17 alerts из alerts.yml

# 7. Тестовый алерт (остановить сервис)
docker compose stop api-gateway
# Через 5 минут alert ServiceDown должен сработать
docker compose start api-gateway
```

---

## Фаза 2: Loki Logging + Zap унификация

См. детальный план в `/Users/konstantin/.claude/plans/foamy-snuggling-porcupine.md`

Основные задачи:
1. Создать shared пакет `services/shared/pkg/observability/logging/`
2. Мигрировать Search Service с Zerolog на Zap
3. Обновить остальные сервисы на shared logging
4. Настроить Loki + Promtail
5. Добавить Loki datasource в Grafana

---

## Фаза 3: Jaeger Tracing + OpenTelemetry

См. детальный план в `/Users/konstantin/.claude/plans/foamy-snuggling-porcupine.md`

Основные задачи:
1. Создать shared пакет `services/shared/pkg/observability/tracing/`
2. Добавить Jaeger в docker-compose.yml
3. Интегрировать OpenTelemetry во все Go сервисы
4. Добавить трейсинг в repository слой
5. Настроить Trace-to-Logs корреляцию в Grafana

---

## Полезные ссылки

- **Детальный план**: `/Users/konstantin/.claude/plans/foamy-snuggling-porcupine.md`
- **Prometheus документация**: https://prometheus.io/docs/
- **Grafana дашборды**: https://grafana.com/grafana/dashboards/
- **Prometheus client Go**: https://github.com/prometheus/client_golang
- **Chi middleware**: https://github.com/go-chi/chi
- **ClickHouse Prometheus metrics**: https://clickhouse.com/docs/en/operations/server-configuration-parameters/settings#server_configuration_parameters-prometheus

---

## Troubleshooting

### Prometheus не собирает метрики с Go сервисов

1. Проверить что `/metrics` endpoint доступен:
   ```bash
   curl http://localhost:9090/metrics
   ```

2. Проверить логи Prometheus:
   ```bash
   docker compose logs prometheus
   ```

3. Проверить targets в Prometheus UI:
   ```
   http://localhost:9090/targets
   ```

### Grafana не показывает данные

1. Проверить Datasource: Configuration → Data Sources → Test
2. Проверить что Prometheus URL правильный: `http://prometheus:9090`
3. Проверить логи Grafana:
   ```bash
   docker compose logs grafana
   ```

### Alert rules не загружаются

1. Проверить синтаксис alerts.yml:
   ```bash
   docker compose exec prometheus promtool check rules /etc/prometheus/rules/alerts.yml
   ```

2. Перезагрузить Prometheus:
   ```bash
   make prometheus-reload
   ```
