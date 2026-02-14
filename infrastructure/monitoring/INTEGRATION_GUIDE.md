# Инструкция по интеграции Prometheus метрик в Go сервисы

## Прогресс Фазы 1

### ✅ Выполнено (10 из 13 задач)

1. ✅ Структура директорий создана
2. ✅ Prometheus конфигурация ([prometheus.yml](prometheus.yml), [alerts.yml](prometheus/rules/alerts.yml))
3. ✅ Shared пакет для метрик ([services/shared/pkg/observability/metrics/](../../services/shared/pkg/observability/metrics/))
4. ✅ Docker Compose обновлен (Prometheus volumes, cAdvisor)
5. ✅ Grafana provisioning ([provisioning/](grafana/provisioning/))
6. ✅ .env.example обновлен
7. ✅ Makefile обновлен (9 новых команд для мониторинга)

### 📋 Осталось (3 задачи)

1. **Добавить /metrics endpoints в 4 Go сервиса** ⬅️ КРИТИЧНО
2. Создать Grafana дашборды (можно импортировать готовые)
3. Тестирование

---

## Шаг 1: Установка зависимостей

### 1.1. Обновить services/shared/go.mod

```bash
cd services/shared
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy
```

### 1.2. Обновить go.mod в каждом сервисе

```bash
cd services/api-gateway
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy

cd ../search-service
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy

cd ../change-detection-service
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy

cd ../notification-service
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy
```

---

## Шаг 2: Интеграция метрик в API Gateway

### 2.1. Обновить импорты в services/api-gateway/cmd/server/main.go

После строки 31 добавить:

```go
import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)
```

### 2.2. Добавить /metrics server

После строки 56 (после `defer logger.Sync()`), добавить:

```go
// Prometheus metrics server на отдельном порту
go func() {
	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	metricsAddr := ":9090"
	logger.Info("Starting metrics server", zap.String("addr", metricsAddr))
	if err := http.ListenAndServe(metricsAddr, metricsRouter); err != nil {
		logger.Fatal("Failed to start metrics server", zap.Error(err))
	}
}()
```

### 2.3. Добавить metrics middleware

Найти строку где создается роутер `r := chi.NewRouter()` (примерно строка 140).

ПЕРЕД другими middleware добавить:

```go
// Prometheus metrics middleware (должен быть ПЕРВЫМ)
r.Use(sharedMetrics.HTTPMiddleware("api-gateway"))
```

### 2.4. Пример использования business метрик

В GraphQL resolvers можно использовать business метрики.

Например, в `internal/graph/schema.resolvers.go`:

```go
import (
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)

func (r *queryResolver) Company(ctx context.Context, ogrn string) (*model.Company, error) {
	// Инкрементируем счетчик поисков компаний
	sharedMetrics.CompanySearchesTotal.Inc()

	// ... существующий код ...
}
```

---

## Шаг 3: Интеграция метрик в Search Service

### 3.1. Обновить services/search-service/main.go

После строки 12 добавить импорты:

```go
import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)
```

### 3.2. Заменить Gin на Chi router (опционально)

Если хотите использовать Chi для единообразия:

```go
func main() {
	// ... logger init ...

	// Prometheus metrics server
	go func() {
		metricsRouter := chi.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		log.Info().Str("addr", ":9091").Msg("Starting metrics server")
		if err := http.ListenAndServe(":9091", metricsRouter); err != nil {
			log.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	// Main router с middleware
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(sharedMetrics.HTTPMiddleware("search-service"))

	// Routes
	setupRoutes(router)

	// ... rest of code ...
}
```

### 3.3. Или добавить в существующий Gin router

Если оставляете Gin, добавьте отдельный metrics server:

```go
// После gin.SetMode(gin.ReleaseMode)

// Prometheus metrics server
go func() {
	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	log.Info().Str("addr", ":9091").Msg("Starting metrics server")
	if err := http.ListenAndServe(":9091", metricsRouter); err != nil {
		log.Fatal().Err(err).Msg("Failed to start metrics server")
	}
}()

// Gin router с middleware для метрик (если есть возможность)
// Для Gin нужен custom middleware, так как sharedMetrics.HTTPMiddleware для Chi
```

---

## Шаг 4: Интеграция метрик в Change Detection Service

### 4.1. Обновить services/change-detection-service/cmd/server/main.go

Аналогично API Gateway:

```go
// Импорты
import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)

// После logger init
go func() {
	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	logger.Info("Starting metrics server", zap.String("addr", ":9092"))
	if err := http.ListenAndServe(":9092", metricsRouter); err != nil {
		logger.Fatal("Failed to start metrics server", zap.Error(err))
	}
}()

// В роутере
r.Use(sharedMetrics.HTTPMiddleware("change-detection"))
```

### 4.2. Использование business метрик

В обработчике детекции изменений:

```go
// Инкрементируем счетчик обнаруженных изменений
sharedMetrics.ChangesDetectedTotal.WithLabelValues(entityType, changeType).Inc()

// Инкрементируем счетчик Kafka сообщений
sharedMetrics.KafkaMessagesProducedTotal.WithLabelValues(topic, "success").Inc()
```

---

## Шаг 5: Интеграция метрик в Notification Service

### 5.1. Обновить services/notification-service/cmd/server/main.go

Аналогично:

```go
// Metrics server
go func() {
	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	logger.Info("Starting metrics server", zap.String("addr", ":9093"))
	if err := http.ListenAndServe(":9093", metricsRouter); err != nil {
		logger.Fatal("Failed to start metrics server", zap.Error(err))
	}
}()

// Middleware
r.Use(sharedMetrics.HTTPMiddleware("notification"))
```

### 5.2. Использование business метрик

В email sender:

```go
// При успешной отправке
sharedMetrics.EmailsSentTotal.WithLabelValues("success").Inc()

// При ошибке
sharedMetrics.EmailsSentTotal.WithLabelValues("error").Inc()
sharedMetrics.SMTPErrorsTotal.Inc()
```

В Kafka consumer:

```go
// Обновление лага консьюмера
sharedMetrics.KafkaConsumerLag.WithLabelValues(topic, partition).Set(float64(lag))
```

---

## Шаг 6: Пересборка сервисов

### 6.1. Локальная пересборка

```bash
# Пересобрать все Go сервисы
make services-build

# Или по отдельности
cd services/api-gateway && go build -o ../../bin/api-gateway ./cmd/server
cd services/search-service && go build -o ../../bin/search-service .
cd services/change-detection-service && go build -o ../../bin/change-detection-service ./cmd/server
cd services/notification-service && go build -o ../../bin/notification-service ./cmd/server
```

### 6.2. Docker пересборка

```bash
# Пересобрать Docker образы
docker compose build api-gateway search-service change-detection-service notification-service

# Или через Makefile (если есть target)
make docker-build
```

---

## Шаг 7: Запуск и проверка

### 7.1. Запуск мониторинга

```bash
# Запустить ClickHouse кластер (если еще не запущен)
make cluster-up

# Запустить мониторинг
make monitoring-up

# Запустить приложения
make up
```

### 7.2. Проверка /metrics endpoints

```bash
# API Gateway
curl http://localhost:9090/metrics | grep http_requests_total

# Search Service
curl http://localhost:9091/metrics | grep http_requests_total

# Change Detection
curl http://localhost:9092/metrics | grep http_requests_total

# Notification
curl http://localhost:9093/metrics | grep http_requests_total
```

### 7.3. Проверка Prometheus targets

Откройте: http://localhost:9090/targets

Должны быть UP (15 targets):
- ClickHouse cluster: 6 nodes
- ClickHouse Keeper: 3 nodes
- api-gateway, search-service, change-detection-service, notification-service: 4
- cAdvisor: 1
- Prometheus self: 1

### 7.4. Проверка метрик в Prometheus

Откройте: http://localhost:9090/graph

Попробуйте queries:
```promql
# Все HTTP запросы
http_requests_total

# Latency P95 по сервисам
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
```

### 7.5. Проверка Grafana

Откройте: http://localhost:3001 (admin/admin)

1. Configuration → Data Sources → проверить Prometheus (должен быть зеленый)
2. Explore → выбрать Prometheus → попробовать query `up`
3. Создать/импортировать дашборды

---

## Шаг 8: Импорт готовых дашбордов в Grafana

### 8.1. Рекомендуемые дашборды

Зайдите в Grafana → Dashboards → Import

**ClickHouse:**
- ID: 14192 - ClickHouse Cluster Overview
- ID: 13606 - ClickHouse Query Performance

**Go Applications:**
- ID: 6671 - Go Processes
- ID: 10826 - Go Metrics

**Docker:**
- ID: 193 - Docker monitoring (cAdvisor)
- ID: 14282 - cAdvisor Prometheus

**Prometheus:**
- ID: 3662 - Prometheus 2.0 Overview
- ID: 6671 - Prometheus Stats

### 8.2. Создание custom дашборда

1. Dashboards → New Dashboard → Add visualization
2. Data source: Prometheus
3. Примеры панелей:

**RPS (Requests Per Second):**
```promql
sum(rate(http_requests_total[5m])) by (service)
```

**Error Rate:**
```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
/
sum(rate(http_requests_total[5m])) by (service)
```

**P95 Latency:**
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (service, le))
```

---

## Шаг 9: Тестирование алертов

### 9.1. Проверка alert rules

```bash
# Проверить что правила загружены
make prometheus-rules-check

# Или в UI
open http://localhost:9090/alerts
```

Должно быть 17 alerts:
- ServiceDown
- ClickHouseNodeDown
- ClickHouseKeeperDown
- HighErrorRate
- HighLatencyP95
- HighLatencyP99
- HighMemoryUsage
- ClickHouseQueryErrors
- ClickHouseReplicationLag
- SMTPErrorsSpike
- KafkaConsumerLagHigh

### 9.2. Тестирование алерта ServiceDown

```bash
# Остановить сервис
docker compose stop api-gateway

# Через 5 минут в Prometheus Alerts должен появиться FIRING alert
# Проверить: http://localhost:9090/alerts

# Запустить обратно
docker compose start api-gateway

# Alert должен перейти в состояние RESOLVED
```

---

## Полезные команды

```bash
# Мониторинг
make monitoring-up          # Запуск мониторинга
make monitoring-down        # Остановка
make monitoring-status      # Статус
make monitoring-logs        # Логи

# Prometheus
make prometheus-reload      # Перезагрузка конфига
make prometheus-check       # Проверка конфига
make prometheus-rules-check # Проверка alert rules
make prometheus-open        # Открыть UI

# Grafana
make grafana-open          # Открыть UI

# Проверка
curl http://localhost:9090/-/healthy  # Prometheus health
curl http://localhost:3001/api/health # Grafana health
```

---

## Troubleshooting

### Метрики не появляются в Prometheus

1. Проверьте targets: http://localhost:9090/targets
2. Если target DOWN:
   - Проверьте что сервис запущен: `docker compose ps`
   - Проверьте логи: `docker compose logs api-gateway`
   - Проверьте что /metrics endpoint доступен: `curl http://localhost:9090/metrics`

### Grafana не показывает метрики

1. Проверьте Data Source: Configuration → Data Sources → Prometheus → Test
2. URL должен быть: `http://prometheus:9090`
3. Проверьте что Prometheus собирает метрики: http://localhost:9090/graph

### Alert rules не загружаются

```bash
# Проверить синтаксис
make prometheus-rules-check

# Проверить логи Prometheus
docker compose logs prometheus | grep -i error
```

---

## Следующие шаги

После завершения Фазы 1:

- **Фаза 2**: Loki Logging + Zap унификация
- **Фаза 3**: Jaeger Tracing + OpenTelemetry

См. детальный план: `/Users/konstantin/.claude/plans/foamy-snuggling-porcupine.md`
