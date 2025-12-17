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
	Server     ServerConfig     `mapstructure:"server"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	Elastic    ElasticConfig    `mapstructure:"elasticsearch"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Log        LogConfig        `mapstructure:"log"`
	GraphQL    GraphQLConfig    `mapstructure:"graphql"`
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

	// Logging
	_ = v.BindEnv("log.level", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "LOG_FORMAT")
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

