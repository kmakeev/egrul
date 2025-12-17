// Package graph содержит GraphQL резолверы
package graph

import (
	"github.com/egrul-system/services/api-gateway/internal/cache"
	"github.com/egrul-system/services/api-gateway/internal/service"
	"go.uber.org/zap"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver содержит зависимости для GraphQL резолверов
type Resolver struct {
	CompanyService      *service.CompanyService
	EntrepreneurService *service.EntrepreneurService
	StatisticsService   *service.StatisticsService
	SearchService       *service.SearchService
	Cache               cache.Cache
	Logger              *zap.Logger
}

// NewResolver создает новый резолвер
func NewResolver(
	companyService *service.CompanyService,
	entrepreneurService *service.EntrepreneurService,
	statisticsService *service.StatisticsService,
	searchService *service.SearchService,
	cache cache.Cache,
	logger *zap.Logger,
) *Resolver {
	return &Resolver{
		CompanyService:      companyService,
		EntrepreneurService: entrepreneurService,
		StatisticsService:   statisticsService,
		SearchService:       searchService,
		Cache:               cache,
		Logger:              logger,
	}
}

