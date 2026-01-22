package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/egrul/services/sync-service/internal/clickhouse"
	"github.com/yourusername/egrul/services/sync-service/internal/config"
	"github.com/yourusername/egrul/services/sync-service/internal/elasticsearch"
	"go.uber.org/zap"
)

type IncrementalSyncer struct {
	chReader     *clickhouse.Reader
	esWriter     *elasticsearch.Writer
	redisClient  *redis.Client
	cfg          config.SyncConfig
	logger       *zap.Logger
}

func NewIncrementalSyncer(
	chReader *clickhouse.Reader,
	esWriter *elasticsearch.Writer,
	redisClient *redis.Client,
	cfg config.SyncConfig,
	logger *zap.Logger,
) *IncrementalSyncer {
	return &IncrementalSyncer{
		chReader:    chReader,
		esWriter:    esWriter,
		redisClient: redisClient,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *IncrementalSyncer) Sync(ctx context.Context) error {
	lastSyncTime, err := s.getLastSyncTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last sync timestamp: %w", err)
	}

	s.logger.Info("Starting incremental sync",
		zap.Time("last_sync", lastSyncTime),
		zap.Int("batch_size", s.cfg.BatchSize))

	totalIndexed := 0

	// Синхронизация обновленных компаний
	companies, err := s.chReader.ReadCompaniesUpdatedAfter(ctx, lastSyncTime, s.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to read updated companies: %w", err)
	}

	if len(companies) > 0 {
		indexed, err := s.esWriter.BulkIndexCompanies(ctx, companies)
		if err != nil {
			s.logger.Error("Failed to index companies", zap.Error(err))
		}
		totalIndexed += indexed
		s.logger.Info("Indexed updated companies", zap.Int("count", indexed))
	}

	// Синхронизация обновленных предпринимателей
	entrepreneurs, err := s.chReader.ReadEntrepreneursUpdatedAfter(ctx, lastSyncTime, s.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to read updated entrepreneurs: %w", err)
	}

	if len(entrepreneurs) > 0 {
		indexed, err := s.esWriter.BulkIndexEntrepreneurs(ctx, entrepreneurs)
		if err != nil {
			s.logger.Error("Failed to index entrepreneurs", zap.Error(err))
		}
		totalIndexed += indexed
		s.logger.Info("Indexed updated entrepreneurs", zap.Int("count", indexed))
	}

	// Сохранить новый timestamp
	if err := s.saveLastSyncTimestamp(ctx, time.Now()); err != nil {
		return fmt.Errorf("failed to save last sync timestamp: %w", err)
	}

	s.logger.Info("Incremental sync completed",
		zap.Int("total_indexed", totalIndexed),
		zap.Int("companies", len(companies)),
		zap.Int("entrepreneurs", len(entrepreneurs)))

	return nil
}

func (s *IncrementalSyncer) getLastSyncTimestamp(ctx context.Context) (time.Time, error) {
	val, err := s.redisClient.Get(ctx, s.cfg.LastTimestampRedisKey).Result()
	if err == redis.Nil {
		// Если ключ не найден, вернуть время начала эпохи Unix
		s.logger.Warn("No last sync timestamp found, using epoch start")
		return time.Unix(0, 0), nil
	}
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	return timestamp, nil
}

func (s *IncrementalSyncer) saveLastSyncTimestamp(ctx context.Context, timestamp time.Time) error {
	return s.redisClient.Set(ctx, s.cfg.LastTimestampRedisKey, timestamp.Format(time.RFC3339), 0).Err()
}
