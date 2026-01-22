package sync

import (
	"context"
	"fmt"

	"github.com/yourusername/egrul/services/sync-service/internal/clickhouse"
	"github.com/yourusername/egrul/services/sync-service/internal/elasticsearch"
	"go.uber.org/zap"
)

type InitialSyncer struct {
	chReader *clickhouse.Reader
	esWriter *elasticsearch.Writer
	logger   *zap.Logger
}

func NewInitialSyncer(chReader *clickhouse.Reader, esWriter *elasticsearch.Writer, logger *zap.Logger) *InitialSyncer {
	return &InitialSyncer{
		chReader: chReader,
		esWriter: esWriter,
		logger:   logger,
	}
}

func (s *InitialSyncer) Sync(ctx context.Context, batchSize int) error {
	s.logger.Info("Starting initial sync", zap.Int("batch_size", batchSize))

	// Синхронизация компаний
	if err := s.syncCompanies(ctx, batchSize); err != nil {
		return fmt.Errorf("failed to sync companies: %w", err)
	}

	// Синхронизация предпринимателей
	if err := s.syncEntrepreneurs(ctx, batchSize); err != nil {
		return fmt.Errorf("failed to sync entrepreneurs: %w", err)
	}

	s.logger.Info("Initial sync completed successfully")
	return nil
}

func (s *InitialSyncer) syncCompanies(ctx context.Context, batchSize int) error {
	totalCount, err := s.chReader.CountCompanies(ctx)
	if err != nil {
		return fmt.Errorf("failed to count companies: %w", err)
	}

	s.logger.Info("Syncing companies", zap.Uint64("total", totalCount))

	var offset int
	totalIndexed := 0

	for {
		companies, err := s.chReader.ReadCompanies(ctx, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to read companies at offset %d: %w", offset, err)
		}

		if len(companies) == 0 {
			break
		}

		indexed, err := s.esWriter.BulkIndexCompanies(ctx, companies)
		if err != nil {
			s.logger.Error("Failed to bulk index companies",
				zap.Int("offset", offset),
				zap.Int("batch_size", len(companies)),
				zap.Error(err))
		}

		totalIndexed += indexed
		offset += len(companies)

		s.logger.Info("Progress",
			zap.String("type", "companies"),
			zap.Int("offset", offset),
			zap.Uint64("total", totalCount),
			zap.Float64("percent", float64(offset)/float64(totalCount)*100))

		if len(companies) < batchSize {
			break
		}
	}

	s.logger.Info("Companies sync completed",
		zap.Int("total_indexed", totalIndexed),
		zap.Uint64("total_count", totalCount))

	return nil
}

func (s *InitialSyncer) syncEntrepreneurs(ctx context.Context, batchSize int) error {
	totalCount, err := s.chReader.CountEntrepreneurs(ctx)
	if err != nil {
		return fmt.Errorf("failed to count entrepreneurs: %w", err)
	}

	s.logger.Info("Syncing entrepreneurs", zap.Uint64("total", totalCount))

	var offset int
	totalIndexed := 0

	for {
		entrepreneurs, err := s.chReader.ReadEntrepreneurs(ctx, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to read entrepreneurs at offset %d: %w", offset, err)
		}

		if len(entrepreneurs) == 0 {
			break
		}

		indexed, err := s.esWriter.BulkIndexEntrepreneurs(ctx, entrepreneurs)
		if err != nil {
			s.logger.Error("Failed to bulk index entrepreneurs",
				zap.Int("offset", offset),
				zap.Int("batch_size", len(entrepreneurs)),
				zap.Error(err))
		}

		totalIndexed += indexed
		offset += len(entrepreneurs)

		s.logger.Info("Progress",
			zap.String("type", "entrepreneurs"),
			zap.Int("offset", offset),
			zap.Uint64("total", totalCount),
			zap.Float64("percent", float64(offset)/float64(totalCount)*100))

		if len(entrepreneurs) < batchSize {
			break
		}
	}

	s.logger.Info("Entrepreneurs sync completed",
		zap.Int("total_indexed", totalIndexed),
		zap.Uint64("total_count", totalCount))

	return nil
}
