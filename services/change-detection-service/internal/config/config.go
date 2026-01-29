package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config представляет конфигурацию сервиса
type Config struct {
	Server     ServerConfig
	ClickHouse ClickHouseConfig
	Kafka      KafkaConfig
	Log        LogConfig
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// ClickHouseConfig конфигурация подключения к ClickHouse
type ClickHouseConfig struct {
	Host             string
	Port             int
	Database         string
	User             string
	Password         string
	MaxOpenConns     int
	MaxIdleConns     int
	MaxExecutionTime int
	Debug            bool
}

// KafkaConfig конфигурация Kafka
type KafkaConfig struct {
	Brokers                 []string
	CompanyChangesTopic     string
	EntrepreneurChangesTopic string
}

// LogConfig конфигурация логирования
type LogConfig struct {
	Level  string
	Format string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Настройка Viper
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Значения по умолчанию
	setDefaults(v)

	cfg := &Config{
		Server: ServerConfig{
			Port:         v.GetInt("PORT"),
			ReadTimeout:  v.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout: v.GetDuration("SERVER_WRITE_TIMEOUT"),
			IdleTimeout:  v.GetDuration("SERVER_IDLE_TIMEOUT"),
		},
		ClickHouse: ClickHouseConfig{
			Host:             v.GetString("CLICKHOUSE_HOST"),
			Port:             v.GetInt("CLICKHOUSE_PORT"),
			Database:         v.GetString("CLICKHOUSE_DATABASE"),
			User:             v.GetString("CLICKHOUSE_USER"),
			Password:         v.GetString("CLICKHOUSE_PASSWORD"),
			MaxOpenConns:     v.GetInt("CLICKHOUSE_MAX_OPEN_CONNS"),
			MaxIdleConns:     v.GetInt("CLICKHOUSE_MAX_IDLE_CONNS"),
			MaxExecutionTime: v.GetInt("CLICKHOUSE_MAX_EXECUTION_TIME"),
			Debug:            v.GetBool("CLICKHOUSE_DEBUG"),
		},
		Kafka: KafkaConfig{
			Brokers:                 v.GetStringSlice("KAFKA_BROKERS"),
			CompanyChangesTopic:     v.GetString("KAFKA_COMPANY_CHANGES_TOPIC"),
			EntrepreneurChangesTopic: v.GetString("KAFKA_ENTREPRENEUR_CHANGES_TOPIC"),
		},
		Log: LogConfig{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults устанавливает значения по умолчанию
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("PORT", 8082)
	v.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 60*time.Second)

	// ClickHouse
	v.SetDefault("CLICKHOUSE_HOST", "localhost")
	v.SetDefault("CLICKHOUSE_PORT", 9000)
	v.SetDefault("CLICKHOUSE_DATABASE", "egrul")
	v.SetDefault("CLICKHOUSE_USER", "default")
	v.SetDefault("CLICKHOUSE_PASSWORD", "")
	v.SetDefault("CLICKHOUSE_MAX_OPEN_CONNS", 10)
	v.SetDefault("CLICKHOUSE_MAX_IDLE_CONNS", 5)
	v.SetDefault("CLICKHOUSE_MAX_EXECUTION_TIME", 300)
	v.SetDefault("CLICKHOUSE_DEBUG", false)

	// Kafka
	v.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	v.SetDefault("KAFKA_COMPANY_CHANGES_TOPIC", "company-changes")
	v.SetDefault("KAFKA_ENTREPRENEUR_CHANGES_TOPIC", "entrepreneur-changes")

	// Log
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.ClickHouse.Host == "" {
		return fmt.Errorf("clickhouse host is required")
	}

	if c.ClickHouse.Database == "" {
		return fmt.Errorf("clickhouse database is required")
	}

	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	if c.Kafka.CompanyChangesTopic == "" {
		return fmt.Errorf("kafka company changes topic is required")
	}

	if c.Kafka.EntrepreneurChangesTopic == "" {
		return fmt.Errorf("kafka entrepreneur changes topic is required")
	}

	return nil
}
