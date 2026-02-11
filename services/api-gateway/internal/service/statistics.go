package service

import (
	"context"
	"fmt"
	"time"

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

// GetDashboardStatistics получает расширенную статистику для дашборда
func (s *StatisticsService) GetDashboardStatistics(ctx context.Context, filter *model.StatsFilter) (*model.DashboardStatistics, error) {
	dashboard := &model.DashboardStatistics{}

	s.logger.Info("получение статистики для дашборда")

	return dashboard, nil
}

// GetRegistrationsByMonth получает временной ряд регистраций и ликвидаций
func (s *StatisticsService) GetRegistrationsByMonth(ctx context.Context, dateFrom, dateTo *string, entityType *model.EntityType, filter *model.StatsFilter) ([]*model.TimeSeriesPoint, error) {
	var from, to *time.Time

	// Парсинг дат если указаны
	if dateFrom != nil && *dateFrom != "" {
		parsed, err := time.Parse("2006-01-02", *dateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid dateFrom format: %w", err)
		}
		from = &parsed
	}

	if dateTo != nil && *dateTo != "" {
		parsed, err := time.Parse("2006-01-02", *dateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid dateTo format: %w", err)
		}
		to = &parsed
	}

	return s.statsRepo.GetRegistrationsByMonth(ctx, from, to, entityType, filter)
}

// GetRegionHeatmap получает статистику для всех регионов (тепловая карта)
func (s *StatisticsService) GetRegionHeatmap(ctx context.Context) ([]*model.RegionStatistics, error) {
	return s.statsRepo.GetRegionHeatmap(ctx)
}
