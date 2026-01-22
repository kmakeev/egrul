package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ClickHouse   ClickHouseConfig
	Elasticsearch ElasticsearchConfig
	Redis        RedisConfig
	Sync         SyncConfig
	LogLevel     string
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type ElasticsearchConfig struct {
	URL string
}

type RedisConfig struct {
	Host string
	Port int
}

type SyncConfig struct {
	Mode                  string        // initial, incremental, daemon
	BatchSize             int           // количество записей для обработки за раз
	Interval              time.Duration // интервал для daemon mode
	LastTimestampRedisKey string        // redis key для хранения timestamp
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("CLICKHOUSE_PORT", "9000"))
	if err != nil {
		return nil, fmt.Errorf("invalid CLICKHOUSE_PORT: %w", err)
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	batchSize, err := strconv.Atoi(getEnv("SYNC_BATCH_SIZE", "10000"))
	if err != nil {
		return nil, fmt.Errorf("invalid SYNC_BATCH_SIZE: %w", err)
	}

	intervalStr := getEnv("SYNC_INTERVAL", "5m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SYNC_INTERVAL: %w", err)
	}

	return &Config{
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "clickhouse"),
			Port:     port,
			User:     getEnv("CLICKHOUSE_USER", "admin"),
			Password: getEnv("CLICKHOUSE_PASSWORD", "admin"),
			Database: getEnv("CLICKHOUSE_DATABASE", "egrul"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL: getEnv("ELASTICSEARCH_URL", "http://elasticsearch:9200"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "redis"),
			Port: redisPort,
		},
		Sync: SyncConfig{
			Mode:                  getEnv("SYNC_MODE", "incremental"),
			BatchSize:             batchSize,
			Interval:              interval,
			LastTimestampRedisKey: getEnv("SYNC_LAST_TIMESTAMP_KEY", "es:last_sync"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
