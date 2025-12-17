// Package config содержит общие конфигурации
package config

import "os"

// DatabaseConfig - конфигурация базы данных
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// ElasticsearchConfig - конфигурация Elasticsearch
type ElasticsearchConfig struct {
	Addresses []string
	Username  string
	Password  string
}

// GetDatabaseConfig возвращает конфигурацию БД из переменных окружения
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "egrul"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// GetElasticsearchConfig возвращает конфигурацию Elasticsearch из переменных окружения
func GetElasticsearchConfig() ElasticsearchConfig {
	return ElasticsearchConfig{
		Addresses: []string{getEnv("ELASTICSEARCH_URL", "http://localhost:9200")},
		Username:  getEnv("ELASTICSEARCH_USER", ""),
		Password:  getEnv("ELASTICSEARCH_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

