// Package clickhouse содержит репозиторий для работы с ClickHouse
package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/egrul-system/services/api-gateway/internal/config"
	"go.uber.org/zap"
)

// Client клиент для работы с ClickHouse
type Client struct {
	conn   driver.Conn
	logger *zap.Logger
}

// NewClient создает новый клиент ClickHouse
func NewClient(cfg *config.ClickHouseConfig, logger *zap.Logger) (*Client, error) {
	opts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     cfg.DialTimeout,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: time.Hour,
		Debug:           cfg.Debug,
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse connection: %w", err)
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	logger.Info("Connected to ClickHouse",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &Client{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// Conn возвращает соединение
func (c *Client) Conn() driver.Conn {
	return c.conn
}

// Ping проверяет соединение
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

