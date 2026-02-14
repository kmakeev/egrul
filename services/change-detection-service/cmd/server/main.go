package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/egrul/change-detection-service/internal/config"
	"github.com/egrul/change-detection-service/internal/detector"
	"github.com/egrul/change-detection-service/internal/handler"
	"github.com/egrul/change-detection-service/internal/kafka"
	"github.com/egrul/change-detection-service/internal/middleware"
	chRepo "github.com/egrul/change-detection-service/internal/repository/clickhouse"
	"github.com/egrul/change-detection-service/internal/service"
	"github.com/go-chi/chi/v5"
	chMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedLogging "github.com/egrul-system/services/shared/pkg/observability/logging"
	sharedMetrics "github.com/egrul-system/services/shared/pkg/observability/metrics"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация logger
	logger, err := sharedLogging.NewLogger(sharedLogging.Config{
		Level:       cfg.Log.Level,
		Format:      "json",
		ServiceName: "change-detection",
	})
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Change Detection Service",
		zap.String("version", "1.0.0"),
		zap.Int("port", cfg.Server.Port),
	)

	// Prometheus metrics server на отдельном порту
	go func() {
		metricsRouter := chi.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":9092"
		logger.Info("Starting metrics server", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, metricsRouter); err != nil {
			logger.Fatal("Failed to start metrics server", zap.Error(err))
		}
	}()

	// Подключение к ClickHouse
	conn, err := connectClickHouse(cfg.ClickHouse, logger)
	if err != nil {
		logger.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}
	defer conn.Close()
	logger.Info("Connected to ClickHouse",
		zap.String("host", cfg.ClickHouse.Host),
		zap.Int("port", cfg.ClickHouse.Port),
	)

	// Инициализация repositories
	companyRepo := chRepo.NewCompanyRepository(conn, logger)
	entrepreneurRepo := chRepo.NewEntrepreneurRepository(conn, logger)
	changeRepo := chRepo.NewChangeRepository(conn, logger)

	// Инициализация детекторов
	classifier := detector.NewClassifier(logger)
	comparator := detector.NewComparator(logger, classifier)

	// Инициализация Kafka producer
	kafkaProducer := initKafkaProducer(cfg.Kafka, logger)
	defer kafkaProducer.Close()
	logger.Info("Kafka producer initialized",
		zap.Strings("brokers", cfg.Kafka.Brokers),
		zap.String("company_topic", cfg.Kafka.CompanyChangesTopic),
		zap.String("entrepreneur_topic", cfg.Kafka.EntrepreneurChangesTopic),
	)

	// Инициализация service
	detectionService := service.NewDetectionService(
		companyRepo,
		entrepreneurRepo,
		changeRepo,
		comparator,
		kafkaProducer,
		logger,
	)

	// Инициализация HTTP handlers
	h := handler.NewHandler(detectionService, logger)

	// Настройка роутера
	r := chi.NewRouter()

	// Middleware
	r.Use(sharedMetrics.HTTPMiddleware("change-detection"))
	r.Use(chMiddleware.RequestID)
	r.Use(chMiddleware.RealIP)
	r.Use(sharedLogging.HTTPMiddleware(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORS())
	r.Use(chMiddleware.Timeout(60 * time.Second))

	// Маршруты
	r.Get("/health", h.HandleHealth)
	r.Get("/stats", h.HandleStats)
	r.Post("/detect", h.HandleDetect)
	r.Get("/company/{ogrn}/changes", h.HandleGetCompanyChanges)
	r.Get("/entrepreneur/{ogrnip}/changes", h.HandleGetEntrepreneurChanges)
	r.Get("/changes/recent", h.HandleGetRecentChanges)

	// HTTP сервер
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		logger.Info("HTTP server started", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// connectClickHouse создает подключение к ClickHouse
func connectClickHouse(cfg config.ClickHouseConfig, logger *zap.Logger) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": cfg.MaxExecutionTime,
		},
		DialTimeout:      30 * time.Second,
		MaxOpenConns:     cfg.MaxOpenConns,
		MaxIdleConns:     cfg.MaxIdleConns,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		Debug:            cfg.Debug,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return conn, nil
}

// initKafkaProducer инициализирует Kafka producer
func initKafkaProducer(cfg config.KafkaConfig, logger *zap.Logger) *kafka.Producer {
	producerCfg := kafka.ProducerConfig{
		Brokers:               cfg.Brokers,
		CompanyTopic:          cfg.CompanyChangesTopic,
		EntrepreneurTopic:     cfg.EntrepreneurChangesTopic,
		RequiredAcks:          1, // Ждем подтверждения от лидера
		BatchSize:             100,
		BatchTimeout:          100 * time.Millisecond,
	}

	return kafka.NewProducer(producerCfg, logger)
}
