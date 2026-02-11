package graph

// This file contains resolvers for DashboardStatistics type fields

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// RegistrationsByMonth is the resolver for the registrationsByMonth field on DashboardStatistics.
func (r *dashboardStatisticsResolver) RegistrationsByMonth(ctx context.Context, obj *model.DashboardStatistics, dateFrom, dateTo *string, entityType *model.EntityType, filter *model.StatsFilter) ([]*model.TimeSeriesPoint, error) {
	r.Logger.Info("getting registrations by month",
		zap.Any("dateFrom", dateFrom),
		zap.Any("dateTo", dateTo),
		zap.Any("entityType", entityType),
		zap.Any("filter", filter))

	timeSeries, err := r.StatisticsService.GetRegistrationsByMonth(ctx, dateFrom, dateTo, entityType, filter)
	if err != nil {
		r.Logger.Error("failed to get registrations by month", zap.Error(err))
		return nil, err
	}

	return timeSeries, nil
}

// RegionHeatmap is the resolver for the regionHeatmap field on DashboardStatistics.
func (r *dashboardStatisticsResolver) RegionHeatmap(ctx context.Context, obj *model.DashboardStatistics) ([]*model.RegionStatistics, error) {
	r.Logger.Info("getting region heatmap")

	regions, err := r.StatisticsService.GetRegionHeatmap(ctx)
	if err != nil {
		r.Logger.Error("failed to get region heatmap", zap.Error(err))
		return nil, err
	}

	return regions, nil
}

// DashboardStatistics returns DashboardStatisticsResolver implementation.
func (r *Resolver) DashboardStatistics() DashboardStatisticsResolver {
	return &dashboardStatisticsResolver{r}
}

type dashboardStatisticsResolver struct{ *Resolver }
