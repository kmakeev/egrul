package service

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/egrul-system/services/api-gateway/internal/repository/clickhouse"
	"go.uber.org/zap"
)

// StatisticsService сервис для работы со статистикой
type StatisticsService struct {
	statsRepo *clickhouse.StatisticsRepository
	logger    *zap.Logger
}

// NewStatisticsService создает новый сервис статистики
func NewStatisticsService(
	statsRepo *clickhouse.StatisticsRepository,
	logger *zap.Logger,
) *StatisticsService {
	return &StatisticsService{
		statsRepo: statsRepo,
		logger:    logger.Named("statistics_service"),
	}
}

// GetStatistics получает общую статистику
func (s *StatisticsService) GetStatistics(ctx context.Context, filter *model.StatsFilter) (*model.Statistics, error) {
	return s.statsRepo.GetStatistics(ctx, filter)
}

// GetActivityStats получает статистику по видам деятельности
func (s *StatisticsService) GetActivityStats(ctx context.Context, limit int) ([]*model.ActivityStatistics, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.statsRepo.GetActivityStats(ctx, limit)
}

