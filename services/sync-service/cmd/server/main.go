package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/egrul/services/sync-service/internal/clickhouse"
	"github.com/yourusername/egrul/services/sync-service/internal/config"
	"github.com/yourusername/egrul/services/sync-service/internal/elasticsearch"
	"github.com/yourusername/egrul/services/sync-service/internal/sync"
	"go.uber.org/zap"
)

func main() {
	// CLI flags
	mode := flag.String("mode", "incremental", "Sync mode: initial, incremental, or daemon")
	flag.Parse()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация logger
	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting sync-service",
		zap.String("mode", *mode),
		zap.String("clickhouse_host", cfg.ClickHouse.Host),
		zap.String("elasticsearch_url", cfg.Elasticsearch.URL))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Инициализация компонентов
	chReader, err := clickhouse.NewReader(cfg.ClickHouse, logger)
	if err != nil {
		logger.Fatal("Failed to create ClickHouse reader", zap.Error(err))
	}
	defer chReader.Close()

	esWriter, err := elasticsearch.NewWriter(cfg.Elasticsearch, logger)
	if err != nil {
		logger.Fatal("Failed to create Elasticsearch writer", zap.Error(err))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
	})
	defer redisClient.Close()

	// Проверка подключения к Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Выполнение синхронизации в зависимости от режима
	switch *mode {
	case "initial":
		if err := runInitialSync(ctx, chReader, esWriter, cfg, logger); err != nil {
			logger.Fatal("Initial sync failed", zap.Error(err))
		}

	case "incremental":
		if err := runIncrementalSync(ctx, chReader, esWriter, redisClient, cfg, logger); err != nil {
			logger.Fatal("Incremental sync failed", zap.Error(err))
		}

	case "daemon":
		runDaemon(ctx, chReader, esWriter, redisClient, cfg, logger)

	default:
		logger.Fatal("Invalid mode", zap.String("mode", *mode))
	}

	logger.Info("Sync-service finished")
}

func runInitialSync(ctx context.Context, chReader *clickhouse.Reader, esWriter *elasticsearch.Writer, cfg *config.Config, logger *zap.Logger) error {
	syncer := sync.NewInitialSyncer(chReader, esWriter, logger)
	return syncer.Sync(ctx, cfg.Sync.BatchSize)
}

func runIncrementalSync(ctx context.Context, chReader *clickhouse.Reader, esWriter *elasticsearch.Writer, redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) error {
	syncer := sync.NewIncrementalSyncer(chReader, esWriter, redisClient, cfg.Sync, logger)
	return syncer.Sync(ctx)
}

func runDaemon(ctx context.Context, chReader *clickhouse.Reader, esWriter *elasticsearch.Writer, redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) {
	syncer := sync.NewIncrementalSyncer(chReader, esWriter, redisClient, cfg.Sync, logger)

	logger.Info("Starting daemon mode", zap.Duration("interval", cfg.Sync.Interval))

	ticker := time.NewTicker(cfg.Sync.Interval)
	defer ticker.Stop()

	// Выполнить первую синхронизацию сразу
	if err := syncer.Sync(ctx); err != nil {
		logger.Error("Sync failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Daemon stopped")
			return
		case <-ticker.C:
			logger.Info("Running scheduled sync")
			if err := syncer.Sync(ctx); err != nil {
				logger.Error("Sync failed", zap.Error(err))
			}
		}
	}
}

func initLogger(logLevel string) (*zap.Logger, error) {
	var cfg zap.Config

	if logLevel == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	// Установить уровень логирования
	level, err := zap.ParseAtomicLevel(logLevel)
	if err == nil {
		cfg.Level = level
	}

	return cfg.Build()
}
