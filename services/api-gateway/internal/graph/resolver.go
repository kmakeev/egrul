// Package graph содержит GraphQL резолверы
package graph

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/auth"
	"github.com/egrul-system/services/api-gateway/internal/cache"
	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/egrul-system/services/api-gateway/internal/repository/postgresql"
	"github.com/egrul-system/services/api-gateway/internal/service"
	"go.uber.org/zap"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// SubscriptionRepository интерфейс для работы с подписками
type SubscriptionRepository interface {
	GetByID(ctx context.Context, id string) (*model.EntitySubscription, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.EntitySubscription, error)
	HasSubscription(ctx context.Context, userID, entityType, entityID string) (bool, error)
	Create(ctx context.Context, subscription *model.EntitySubscription) error
	Update(ctx context.Context, subscription *model.EntitySubscription) error
	Delete(ctx context.Context, id string) error
	GetNotificationHistory(ctx context.Context, subscriptionID string, limit, offset int) ([]*model.NotificationLogEntry, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*postgresql.User, error)
	GetByEmail(ctx context.Context, email string) (*postgresql.User, error)
	Create(ctx context.Context, user *postgresql.User) error
	Update(ctx context.Context, user *postgresql.User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

// FavoriteRepository интерфейс для работы с избранным
type FavoriteRepository interface {
	GetByID(ctx context.Context, id string) (*model.Favorite, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Favorite, error)
	HasFavorite(ctx context.Context, userID, entityType, entityID string) (bool, error)
	Create(ctx context.Context, favorite *model.Favorite) error
	Update(ctx context.Context, favorite *model.Favorite) error
	Delete(ctx context.Context, id string) error
}

// JWTManager интерфейс для работы с JWT токенами
type JWTManager interface {
	Generate(userID, email string) (string, error)
	Verify(tokenString string) (*auth.Claims, error)
}

// Resolver содержит зависимости для GraphQL резолверов
type Resolver struct {
	CompanyService      *service.CompanyService
	EntrepreneurService *service.EntrepreneurService
	StatisticsService   *service.StatisticsService
	SearchService       *service.SearchService
	SubscriptionRepo    SubscriptionRepository
	FavoriteRepo        FavoriteRepository
	UserRepo            UserRepository
	JWTManager          JWTManager
	Cache               cache.Cache
	Logger              *zap.Logger
}

// NewResolver создает новый резолвер
func NewResolver(
	companyService *service.CompanyService,
	entrepreneurService *service.EntrepreneurService,
	statisticsService *service.StatisticsService,
	searchService *service.SearchService,
	subscriptionRepo SubscriptionRepository,
	favoriteRepo FavoriteRepository,
	userRepo UserRepository,
	jwtManager JWTManager,
	cache cache.Cache,
	logger *zap.Logger,
) *Resolver {
	return &Resolver{
		CompanyService:      companyService,
		EntrepreneurService: entrepreneurService,
		StatisticsService:   statisticsService,
		SearchService:       searchService,
		SubscriptionRepo:    subscriptionRepo,
		FavoriteRepo:        favoriteRepo,
		UserRepo:            userRepo,
		JWTManager:          jwtManager,
		Cache:               cache,
		Logger:              logger,
	}
}

