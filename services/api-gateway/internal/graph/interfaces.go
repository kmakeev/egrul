package graph

// This file contains interfaces for GraphQL resolvers

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
)

// QueryResolver interface for Query resolvers
type QueryResolver interface {
	Company(ctx context.Context, ogrn string) (*model.Company, error)
	CompanyByInn(ctx context.Context, inn string) (*model.Company, error)
	Companies(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) (*model.CompanyConnection, error)
	SearchCompanies(ctx context.Context, query string, limit *int, offset *int) ([]*model.Company, error)
	Entrepreneur(ctx context.Context, ogrnip string) (*model.Entrepreneur, error)
	EntrepreneurByInn(ctx context.Context, inn string) (*model.Entrepreneur, error)
	Entrepreneurs(ctx context.Context, filter *model.EntrepreneurFilter, pagination *model.Pagination, sort *model.EntrepreneurSort) (*model.EntrepreneurConnection, error)
	SearchEntrepreneurs(ctx context.Context, query string, limit *int, offset *int) ([]*model.Entrepreneur, error)
	Search(ctx context.Context, query string, limit *int) (*model.SearchResult, error)
	Statistics(ctx context.Context, filter *model.StatsFilter) (*model.Statistics, error)
	EntityHistory(ctx context.Context, entityType model.EntityType, entityID string, limit *int, offset *int) ([]*model.HistoryRecord, error)
	EntityHistoryCount(ctx context.Context, entityType model.EntityType, entityID string) (int, error)
	CompanyFounders(ctx context.Context, ogrn string, limit *int, offset *int) ([]*model.Founder, error)
	RelatedCompanies(ctx context.Context, inn string, limit *int, offset *int) ([]*model.Company, error)
}

// CompanyResolver interface for Company field resolvers
type CompanyResolver interface {
	Founders(ctx context.Context, obj *model.Company, limit *int, offset *int) ([]*model.Founder, error)
	Licenses(ctx context.Context, obj *model.Company) ([]*model.License, error)
	Branches(ctx context.Context, obj *model.Company) ([]*model.Branch, error)
	History(ctx context.Context, obj *model.Company, limit *int, offset *int) ([]*model.HistoryRecord, error)
	HistoryCount(ctx context.Context, obj *model.Company) (int, error)
	RelatedCompanies(ctx context.Context, obj *model.Company, limit *int, offset *int) ([]*model.RelatedCompany, error)
}

// EntrepreneurResolver interface for Entrepreneur field resolvers
type EntrepreneurResolver interface {
	Licenses(ctx context.Context, obj *model.Entrepreneur) ([]*model.License, error)
	History(ctx context.Context, obj *model.Entrepreneur, limit *int, offset *int) ([]*model.HistoryRecord, error)
	HistoryCount(ctx context.Context, obj *model.Entrepreneur) (int, error)
}

// StatisticsResolver interface for Statistics field resolvers
type StatisticsResolver interface {
	ByActivity(ctx context.Context, obj *model.Statistics, limit *int) ([]*model.ActivityStatistics, error)
}

// MutationResolver interface for Mutation resolvers
type MutationResolver interface {
	CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.EntitySubscription, error)
	UpdateSubscriptionFilters(ctx context.Context, input model.UpdateSubscriptionFiltersInput) (*model.EntitySubscription, error)
	UpdateSubscriptionChannels(ctx context.Context, input model.UpdateSubscriptionChannelsInput) (*model.EntitySubscription, error)
	DeleteSubscription(ctx context.Context, id string) (bool, error)
	ToggleSubscription(ctx context.Context, input model.ToggleSubscriptionInput) (*model.EntitySubscription, error)
	CreateFavorite(ctx context.Context, input model.CreateFavoriteInput) (*model.Favorite, error)
	UpdateFavoriteNotes(ctx context.Context, input model.UpdateFavoriteNotesInput) (*model.Favorite, error)
	DeleteFavorite(ctx context.Context, id string) (bool, error)
}

// ResolverRoot is the main resolver interface
type ResolverRoot interface {
	Query() QueryResolver
	Company() CompanyResolver
	Entrepreneur() EntrepreneurResolver
	Statistics() StatisticsResolver
	Mutation() MutationResolver
}

