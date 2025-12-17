package graph

// This file contains resolvers for Statistics type fields

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// ByActivity is the resolver for the byActivity field on Statistics.
func (r *statisticsResolver) ByActivity(ctx context.Context, obj *model.Statistics, limit *int) ([]*model.ActivityStatistics, error) {
	l := 20
	if limit != nil {
		l = *limit
	}

	stats, err := r.StatisticsService.GetActivityStats(ctx, l)
	if err != nil {
		r.Logger.Error("failed to get activity stats", zap.Error(err))
		return nil, err
	}
	return stats, nil
}

// Statistics returns StatisticsResolver implementation.
func (r *Resolver) Statistics() StatisticsResolver { return &statisticsResolver{r} }

type statisticsResolver struct{ *Resolver }

