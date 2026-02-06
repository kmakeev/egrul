// Package config содержит конфигурацию API Gateway
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config - основная структура конфигурации
type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	ClickHouse      ClickHouseConfig      `mapstructure:"clickhouse"`
	Elasticsearch   ElasticConfig         `mapstructure:"elasticsearch"`
	PostgreSQL      PostgreSQLConfig      `mapstructure:"postgresql"`
	Redis           RedisConfig           `mapstructure:"redis"`
	Kafka           KafkaConfig           `mapstructure:"kafka"`
	NotificationHub NotificationHubConfig `mapstructure:"notification_hub"`
	Log             LogConfig             `mapstructure:"log"`
	GraphQL         GraphQLConfig         `mapstructure:"graphql"`
	Auth            AuthConfig            `mapstructure:"auth"`
}

// ServerConfig - конфигурация HTTP сервера
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// ClickHouseConfig - конфигурация ClickHouse
type ClickHouseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	HTTPPort     int           `mapstructure:"http_port"`
	Database     string        `mapstructure:"database"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	Debug        bool          `mapstructure:"debug"`
}

// ElasticConfig - конфигурация Elasticsearch
type ElasticConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

// RedisConfig - конфигурация Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// PostgreSQLConfig - конфигурация PostgreSQL
type PostgreSQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
	Schema   string `mapstructure:"schema"` // subscriptions schema
}

// LogConfig - конфигурация логирования
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	TimeFormat string `mapstructure:"time_format"`
}

// GraphQLConfig - конфигурация GraphQL
type GraphQLConfig struct {
	PlaygroundEnabled bool `mapstructure:"playground_enabled"`
	IntrospectionEnabled bool `mapstructure:"introspection_enabled"`
	MaxDepth          int  `mapstructure:"max_depth"`
	MaxComplexity     int  `mapstructure:"max_complexity"`
}

// AuthConfig - конфигурация аутентификации
type AuthConfig struct {
	JWTSecretKey     string        `mapstructure:"jwt_secret_key"`
	JWTTokenDuration time.Duration `mapstructure:"jwt_token_duration"`
}

// KafkaConfig - конфигурация Kafka
type KafkaConfig struct {
	Brokers              []string `mapstructure:"brokers"`
	CompanyTopic         string   `mapstructure:"company_topic"`
	EntrepreneurTopic    string   `mapstructure:"entrepreneur_topic"`
	ConsumerGroup        string   `mapstructure:"consumer_group"`
}

// NotificationHubConfig - конфигурация Notification Hub
type NotificationHubConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	BufferSize        int           `mapstructure:"buffer_size"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	MaxClients        int           `mapstructure:"max_clients"`
}

// Load загружает конфигурацию из файла и переменных окружения
func Load() (*Config, error) {
	v := viper.New()

	// Установка значений по умолчанию
	setDefaults(v)

	// Настройка чтения из переменных окружения
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Биндинг переменных окружения
	bindEnvVariables(v)

	// Чтение конфигурационного файла (если есть)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/api-gateway/")
	v.AddConfigPath("$HOME/.api-gateway")

	// Игнорируем ошибку если файл не найден - используем переменные окружения
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Post-processing: split KAFKA_BROKERS если это строка
	// Viper не умеет автоматически разбивать строки в массивы при чтении из env
	if brokerStr := v.GetString("kafka.brokers"); brokerStr != "" && brokerStr != "localhost:9092" {
		// Если строка не пустая и не default, разбить её
		cfg.Kafka.Brokers = strings.Split(brokerStr, ",")
		for i := range cfg.Kafka.Brokers {
			cfg.Kafka.Brokers[i] = strings.TrimSpace(cfg.Kafka.Brokers[i])
		}
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)

	// ClickHouse
	v.SetDefault("clickhouse.host", "localhost")
	v.SetDefault("clickhouse.port", 9000)
	v.SetDefault("clickhouse.http_port", 8123)
	v.SetDefault("clickhouse.database", "egrul")
	v.SetDefault("clickhouse.username", "default")
	v.SetDefault("clickhouse.password", "")
	v.SetDefault("clickhouse.max_open_conns", 10)
	v.SetDefault("clickhouse.max_idle_conns", 5)
	v.SetDefault("clickhouse.dial_timeout", 10*time.Second)
	v.SetDefault("clickhouse.debug", false)

	// Elasticsearch
	v.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	v.SetDefault("elasticsearch.username", "")
	v.SetDefault("elasticsearch.password", "")

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// PostgreSQL
	v.SetDefault("postgresql.host", "localhost")
	v.SetDefault("postgresql.port", 5432)
	v.SetDefault("postgresql.database", "egrul")
	v.SetDefault("postgresql.user", "postgres")
	v.SetDefault("postgresql.password", "")
	v.SetDefault("postgresql.sslmode", "disable")
	v.SetDefault("postgresql.schema", "subscriptions")

	// Logging
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.time_format", "2006-01-02T15:04:05.000Z07:00")

	// GraphQL
	v.SetDefault("graphql.playground_enabled", true)
	v.SetDefault("graphql.introspection_enabled", true)
	v.SetDefault("graphql.max_depth", 15)
	v.SetDefault("graphql.max_complexity", 1000)

	// Auth
	v.SetDefault("auth.jwt_secret_key", "CHANGE_ME_IN_PRODUCTION_MIN_32_CHARS")
	v.SetDefault("auth.jwt_token_duration", 24*time.Hour)

	// Kafka
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.company_topic", "company-changes")
	v.SetDefault("kafka.entrepreneur_topic", "entrepreneur-changes")
	v.SetDefault("kafka.consumer_group", "api-gateway-notifications")

	// Notification Hub
	v.SetDefault("notification_hub.enabled", true)
	v.SetDefault("notification_hub.buffer_size", 100)
	v.SetDefault("notification_hub.heartbeat_interval", 30*time.Second)
	v.SetDefault("notification_hub.max_clients", 1000)
}

func bindEnvVariables(v *viper.Viper) {
	// Server
	_ = v.BindEnv("server.port", "PORT")

	// ClickHouse
	_ = v.BindEnv("clickhouse.host", "CLICKHOUSE_HOST")
	_ = v.BindEnv("clickhouse.port", "CLICKHOUSE_PORT")
	_ = v.BindEnv("clickhouse.http_port", "CLICKHOUSE_HTTP_PORT")
	_ = v.BindEnv("clickhouse.database", "CLICKHOUSE_DATABASE")
	_ = v.BindEnv("clickhouse.username", "CLICKHOUSE_USER")
	_ = v.BindEnv("clickhouse.password", "CLICKHOUSE_PASSWORD")

	// Elasticsearch
	_ = v.BindEnv("elasticsearch.addresses", "ELASTICSEARCH_URL")

	// Redis
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")

	// PostgreSQL
	_ = v.BindEnv("postgresql.host", "POSTGRES_HOST")
	_ = v.BindEnv("postgresql.port", "POSTGRES_PORT")
	_ = v.BindEnv("postgresql.database", "POSTGRES_DB")
	_ = v.BindEnv("postgresql.user", "POSTGRES_USER")
	_ = v.BindEnv("postgresql.password", "POSTGRES_PASSWORD")
	_ = v.BindEnv("postgresql.sslmode", "POSTGRES_SSLMODE")
	_ = v.BindEnv("postgresql.schema", "POSTGRES_SUBSCRIPTION_SCHEMA")

	// Logging
	_ = v.BindEnv("log.level", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "LOG_FORMAT")

	// Auth
	_ = v.BindEnv("auth.jwt_secret_key", "JWT_SECRET_KEY")
	_ = v.BindEnv("auth.jwt_token_duration", "JWT_TOKEN_DURATION")

	// Kafka
	_ = v.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	_ = v.BindEnv("kafka.company_topic", "KAFKA_COMPANY_CHANGES_TOPIC")
	_ = v.BindEnv("kafka.entrepreneur_topic", "KAFKA_ENTREPRENEUR_CHANGES_TOPIC")
	_ = v.BindEnv("kafka.consumer_group", "NOTIFICATION_HUB_KAFKA_GROUP")

	// Notification Hub
	_ = v.BindEnv("notification_hub.enabled", "NOTIFICATION_HUB_ENABLED")
	_ = v.BindEnv("notification_hub.buffer_size", "NOTIFICATION_HUB_BUFFER_SIZE")
	_ = v.BindEnv("notification_hub.heartbeat_interval", "NOTIFICATION_HUB_HEARTBEAT_INTERVAL")
	_ = v.BindEnv("notification_hub.max_clients", "NOTIFICATION_HUB_MAX_CLIENTS")
}

// Addr возвращает адрес сервера
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DSN возвращает строку подключения к ClickHouse
func (c *ClickHouseConfig) DSN() string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=%s&max_open_conns=%d&max_idle_conns=%d",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.DialTimeout.String(),
		c.MaxOpenConns,
		c.MaxIdleConns,
	)
}

// URL возвращает первый адрес Elasticsearch (для обратной совместимости)
func (c *ElasticConfig) URL() string {
	if len(c.Addresses) > 0 {
		return c.Addresses[0]
	}
	return ""
}

