package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egrul/notification-service/internal/channels"
	"github.com/egrul/notification-service/internal/config"
	"github.com/egrul/notification-service/internal/consumer"
	pgRepo "github.com/egrul/notification-service/internal/repository/postgresql"
	"github.com/egrul/notification-service/internal/service"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sharedLogging "github.com/egrul-system/services/shared/pkg/observability/logging"
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
		ServiceName: "notification",
	})
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Notification Service",
		zap.String("version", "1.0.0"),
		zap.Int("port", cfg.Server.Port),
	)

	// Prometheus metrics server на отдельном порту
	go func() {
		metricsRouter := chi.NewRouter()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":9093"
		logger.Info("Starting metrics server", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, metricsRouter); err != nil {
			logger.Fatal("Failed to start metrics server", zap.Error(err))
		}
	}()

	// Подключение к PostgreSQL
	db, err := connectPostgreSQL(cfg.PostgreSQL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.PostgreSQL.Host),
		zap.Int("port", cfg.PostgreSQL.Port),
	)

	// Инициализация repositories
	subscriptionRepo := pgRepo.NewSubscriptionRepository(db, cfg.PostgreSQL.Schema, logger)
	notificationLogRepo := pgRepo.NewNotificationLogRepository(db, cfg.PostgreSQL.Schema, logger)

	// Загрузка email шаблонов
	htmlTemplate, textTemplate, err := loadEmailTemplates()
	if err != nil {
		logger.Fatal("Failed to load email templates", zap.Error(err))
	}
	logger.Info("Email templates loaded successfully")

	// Инициализация Email channel
	emailConfig := channels.EmailConfig{
		Host:          cfg.SMTP.Host,
		Port:          cfg.SMTP.Port,
		Username:      cfg.SMTP.Username,
		Password:      cfg.SMTP.Password,
		From:          cfg.SMTP.From,
		FromName:      cfg.SMTP.FromName,
		TLS:           cfg.SMTP.TLS,
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
		DryRun:        cfg.SMTP.DryRun,
	}
	emailChannel := channels.NewEmailChannel(emailConfig, htmlTemplate, textTemplate, logger)
	defer emailChannel.Close()
	logger.Info("Email channel initialized",
		zap.String("smtp_host", cfg.SMTP.Host),
		zap.Int("smtp_port", cfg.SMTP.Port),
		zap.String("from", cfg.SMTP.From),
		zap.Bool("dry_run", cfg.SMTP.DryRun),
	)

	// Инициализация service
	notificationService := service.NewNotificationService(
		subscriptionRepo,
		notificationLogRepo,
		emailChannel,
		logger,
	)

	// Инициализация Kafka consumer
	consumerConfig := consumer.ConsumerConfig{
		Brokers:           cfg.Kafka.Brokers,
		CompanyTopic:      cfg.Kafka.CompanyChangesTopic,
		EntrepreneurTopic: cfg.Kafka.EntrepreneurChangesTopic,
		GroupID:           cfg.Kafka.ConsumerGroup,
	}
	kafkaConsumer := consumer.NewKafkaConsumer(consumerConfig, notificationService, logger)
	defer kafkaConsumer.Close()
	logger.Info("Kafka consumer initialized",
		zap.Strings("brokers", cfg.Kafka.Brokers),
		zap.String("company_topic", cfg.Kafka.CompanyChangesTopic),
		zap.String("entrepreneur_topic", cfg.Kafka.EntrepreneurChangesTopic),
		zap.String("consumer_group", cfg.Kafka.ConsumerGroup),
	)

	// Запускаем Kafka consumer в отдельной горутине
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("Starting Kafka consumer...")
		if err := kafkaConsumer.Start(ctx); err != nil && err != context.Canceled {
			logger.Fatal("Kafka consumer error", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down service...")

	// Graceful shutdown
	cancel()

	// Даем время на завершение обработки текущих сообщений
	time.Sleep(5 * time.Second)

	logger.Info("Service stopped gracefully")
}

// initLogger инициализирует zap logger
// connectPostgreSQL создает подключение к PostgreSQL
func connectPostgreSQL(cfg config.PostgreSQLConfig, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Настройка connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// loadEmailTemplates загружает email шаблоны
func loadEmailTemplates() (*template.Template, *template.Template, error) {
	// HTML шаблон
	htmlTemplate, err := template.ParseFiles("internal/templates/email_change_notification.html")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Текстовый шаблон
	textTemplate, err := template.ParseFiles("internal/templates/email_change_notification.txt")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse text template: %w", err)
	}

	return htmlTemplate, textTemplate, nil
}
