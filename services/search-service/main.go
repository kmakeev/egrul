package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedLogging "github.com/egrul-system/services/shared/pkg/observability/logging"
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)

// prometheusMiddleware - Gin middleware для сбора Prometheus метрик
func prometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Обработка запроса
		c.Next()

		// Сбор метрик
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		sharedMetrics.HTTPRequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.Request.URL.Path,
			status,
		).Inc()

		sharedMetrics.HTTPRequestDuration.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.Request.URL.Path,
		).Observe(duration)
	}
}

func main() {
	// Настройка логирования
	logger, err := sharedLogging.NewLogger(sharedLogging.Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      "json",
		ServiceName: "search-service",
	})
	if err != nil {
		panic("Failed to init logger: " + err.Error())
	}
	defer logger.Sync()

	// Prometheus metrics server на отдельном порту
	go func() {
		metricsRouter := chi.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":9091"
		logger.Info("Starting metrics server", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, metricsRouter); err != nil {
			logger.Fatal("Failed to start metrics server", zap.Error(err))
		}
	}()

	// Настройка Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(sharedLogging.GinMiddleware(logger))
	router.Use(prometheusMiddleware("search-service"))

	// Роуты
	setupRoutes(router)

	// Сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Запуск Search Service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка запуска сервера", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Ошибка остановки сервера", zap.Error(err))
	}

	logger.Info("Сервер остановлен")
}

func setupRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "search-service",
		})
	})

	// Search API
	r.POST("/search", handleSearch)
	r.POST("/index", handleIndex)
	r.DELETE("/index/:id", handleDeleteFromIndex)
}

// SearchRequest - запрос на поиск
type SearchRequest struct {
	Query    string   `json:"query" binding:"required"`
	Type     string   `json:"type,omitempty"` // legal_entity, entrepreneur, all
	Filters  []Filter `json:"filters,omitempty"`
	Page     int      `json:"page,omitempty"`
	PageSize int      `json:"page_size,omitempty"`
}

// Filter - фильтр поиска
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, in, contains
	Value    interface{} `json:"value"`
}

// IndexRequest - запрос на индексацию
type IndexRequest struct {
	ID   string      `json:"id" binding:"required"`
	Type string      `json:"type" binding:"required"`
	Data interface{} `json:"data" binding:"required"`
}

func handleSearch(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Реализовать поиск через Elasticsearch

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": []interface{}{},
		"total":   0,
		"page":    req.Page,
	})
}

func handleIndex(c *gin.Context) {
	var req IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Реализовать индексацию в Elasticsearch

	c.JSON(http.StatusOK, gin.H{
		"status": "indexed",
		"id":     req.ID,
	})
}

func handleDeleteFromIndex(c *gin.Context) {
	id := c.Param("id")

	// TODO: Реализовать удаление из индекса

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"id":     id,
	})
}

// getEnv - helper для получения env переменной с дефолтным значением
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

