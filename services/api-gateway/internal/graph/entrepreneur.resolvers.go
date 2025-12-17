package graph

// This file contains resolvers for Entrepreneur type fields that require additional data loading

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// Licenses is the resolver for the licenses field on Entrepreneur.
func (r *entrepreneurResolver) Licenses(ctx context.Context, obj *model.Entrepreneur) ([]*model.License, error) {
	licenses, err := r.EntrepreneurService.GetLicenses(ctx, obj.Ogrnip)
	if err != nil {
		r.Logger.Error("failed to get licenses", zap.String("ogrnip", obj.Ogrnip), zap.Error(err))
		return nil, err
	}
	return licenses, nil
}

// History is the resolver for the history field on Entrepreneur.
func (r *entrepreneurResolver) History(ctx context.Context, obj *model.Entrepreneur, limit *int, offset *int) ([]*model.HistoryRecord, error) {
	l := 50
	if limit != nil {
		l = *limit
	}
	o := 0
	if offset != nil {
		o = *offset
	}

	history, err := r.EntrepreneurService.GetHistory(ctx, obj.Ogrnip, l, o)
	if err != nil {
		r.Logger.Error("failed to get history", zap.String("ogrnip", obj.Ogrnip), zap.Error(err))
		return nil, err
	}
	return history, nil
}

// Entrepreneur returns EntrepreneurResolver implementation.
func (r *Resolver) Entrepreneur() EntrepreneurResolver { return &entrepreneurResolver{r} }

type entrepreneurResolver struct{ *Resolver }

