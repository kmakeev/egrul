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
	PostgreSQL PostgreSQLConfig
	Kafka      KafkaConfig
	SMTP       SMTPConfig
	Log        LogConfig
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// PostgreSQLConfig конфигурация подключения к PostgreSQL
type PostgreSQLConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
	Schema   string // subscriptions schema
}

// KafkaConfig конфигурация Kafka consumer
type KafkaConfig struct {
	Brokers                  []string
	CompanyChangesTopic      string
	EntrepreneurChangesTopic string
	ConsumerGroup            string
}

// SMTPConfig конфигурация SMTP сервера
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	TLS      bool
	DryRun   bool // Режим тестирования - только логировать, не отправлять
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
		PostgreSQL: PostgreSQLConfig{
			Host:     v.GetString("POSTGRES_HOST"),
			Port:     v.GetInt("POSTGRES_PORT"),
			Database: v.GetString("POSTGRES_DB"),
			User:     v.GetString("POSTGRES_USER"),
			Password: v.GetString("POSTGRES_PASSWORD"),
			SSLMode:  v.GetString("POSTGRES_SSLMODE"),
			Schema:   v.GetString("POSTGRES_SUBSCRIPTION_SCHEMA"),
		},
		Kafka: KafkaConfig{
			Brokers:                  v.GetStringSlice("KAFKA_BROKERS"),
			CompanyChangesTopic:      v.GetString("KAFKA_COMPANY_CHANGES_TOPIC"),
			EntrepreneurChangesTopic: v.GetString("KAFKA_ENTREPRENEUR_CHANGES_TOPIC"),
			ConsumerGroup:            v.GetString("KAFKA_CONSUMER_GROUP"),
		},
		SMTP: SMTPConfig{
			Host:     v.GetString("SMTP_HOST"),
			Port:     v.GetInt("SMTP_PORT"),
			Username: v.GetString("SMTP_USERNAME"),
			Password: v.GetString("SMTP_PASSWORD"),
			From:     v.GetString("SMTP_FROM"),
			FromName: v.GetString("SMTP_FROM_NAME"),
			TLS:      v.GetBool("SMTP_TLS"),
			DryRun:   v.GetBool("EMAIL_DRY_RUN"),
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
	v.SetDefault("PORT", 8083)
	v.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 60*time.Second)

	// PostgreSQL
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_DB", "egrul")
	v.SetDefault("POSTGRES_USER", "postgres")
	v.SetDefault("POSTGRES_PASSWORD", "")
	v.SetDefault("POSTGRES_SSLMODE", "disable")
	v.SetDefault("POSTGRES_SUBSCRIPTION_SCHEMA", "subscriptions")

	// Kafka
	v.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	v.SetDefault("KAFKA_COMPANY_CHANGES_TOPIC", "company-changes")
	v.SetDefault("KAFKA_ENTREPRENEUR_CHANGES_TOPIC", "entrepreneur-changes")
	v.SetDefault("KAFKA_CONSUMER_GROUP", "notification-service-group")

	// SMTP
	v.SetDefault("SMTP_HOST", "localhost")
	v.SetDefault("SMTP_PORT", 1025)
	v.SetDefault("SMTP_USERNAME", "")
	v.SetDefault("SMTP_PASSWORD", "")
	v.SetDefault("SMTP_FROM", "noreply@egrul.local")
	v.SetDefault("SMTP_FROM_NAME", "ЕГРЮЛ/ЕГРИП Мониторинг")
	v.SetDefault("SMTP_TLS", false)
	v.SetDefault("EMAIL_DRY_RUN", false)

	// Log
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.PostgreSQL.Host == "" {
		return fmt.Errorf("postgres host is required")
	}

	if c.PostgreSQL.Database == "" {
		return fmt.Errorf("postgres database is required")
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

	if c.Kafka.ConsumerGroup == "" {
		return fmt.Errorf("kafka consumer group is required")
	}

	if c.SMTP.Host == "" {
		return fmt.Errorf("smtp host is required")
	}

	if c.SMTP.From == "" {
		return fmt.Errorf("smtp from address is required")
	}

	return nil
}

// GetPostgreSQLDSN возвращает строку подключения к PostgreSQL
func (c *PostgreSQLConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}
