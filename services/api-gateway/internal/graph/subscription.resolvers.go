package graph

// This file contains subscription resolvers for ManualHandler
// DO NOT import generated package - we use manual routing

import (
	"context"
	"fmt"

	"github.com/egrul-system/services/api-gateway/internal/auth"
	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// Empty is the resolver for the _empty field.
func (r *mutationResolver) Empty(ctx context.Context) (*string, error) {
	// Placeholder для Mutation type
	empty := ""
	return &empty, nil
}

// CreateSubscription is the resolver for the createSubscription field.
func (r *mutationResolver) CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	// Получаем userID из JWT context
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized: login required")
	}

	// Получаем email пользователя из UserRepo (требуется для БД)
	user, err := r.UserRepo.GetByID(ctx, userID)
	if err != nil {
		r.Logger.Error("failed to get user",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user data")
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Проверяем, нет ли уже подписки
	exists, err := r.SubscriptionRepo.HasSubscription(ctx, userID, string(input.EntityType), input.EntityID)
	if err != nil {
		r.Logger.Error("failed to check subscription existence",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("subscription already exists for this entity")
	}

	// Создаем подписку
	subscription := &model.EntitySubscription{
		UserID:               userID,
		UserEmail:            user.Email, // Добавляем email для БД
		EntityType:           input.EntityType,
		EntityID:             input.EntityID,
		EntityName:           input.EntityName,
		ChangeFilters:        input.ChangeFilters.ToChangeFilters(),
		NotificationChannels: input.NotificationChannels.ToNotificationChannels(),
		IsActive:             true,
	}

	if err := r.SubscriptionRepo.Create(ctx, subscription); err != nil {
		r.Logger.Error("failed to create subscription",
			zap.String("user_id", userID),
			zap.String("entity_id", input.EntityID),
			zap.Error(err),
		)
		return nil, err
	}

	r.Logger.Info("subscription created",
		zap.String("id", subscription.ID),
		zap.String("user_id", subscription.UserID),
		zap.String("entity_type", string(subscription.EntityType)),
		zap.String("entity_id", subscription.EntityID),
	)

	return subscription, nil
}

// UpdateSubscriptionFilters is the resolver for the updateSubscriptionFilters field.
func (r *mutationResolver) UpdateSubscriptionFilters(ctx context.Context, input model.UpdateSubscriptionFiltersInput) (*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	subscription, err := r.SubscriptionRepo.GetByID(ctx, input.ID)
	if err != nil {
		r.Logger.Error("failed to get subscription",
			zap.String("id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	if subscription == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	// Обновляем фильтры
	subscription.ChangeFilters = input.ChangeFilters.ToChangeFilters()

	if err := r.SubscriptionRepo.Update(ctx, subscription); err != nil {
		r.Logger.Error("failed to update subscription filters",
			zap.String("id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	r.Logger.Info("subscription filters updated",
		zap.String("id", subscription.ID),
	)

	return subscription, nil
}

// UpdateSubscriptionChannels is the resolver for the updateSubscriptionChannels field.
func (r *mutationResolver) UpdateSubscriptionChannels(ctx context.Context, input model.UpdateSubscriptionChannelsInput) (*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	subscription, err := r.SubscriptionRepo.GetByID(ctx, input.ID)
	if err != nil {
		r.Logger.Error("failed to get subscription",
			zap.String("id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	if subscription == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	// Обновляем каналы
	subscription.NotificationChannels = input.NotificationChannels.ToNotificationChannels()

	if err := r.SubscriptionRepo.Update(ctx, subscription); err != nil {
		r.Logger.Error("failed to update subscription channels",
			zap.String("id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	r.Logger.Info("subscription channels updated",
		zap.String("id", subscription.ID),
	)

	return subscription, nil
}

// DeleteSubscription is the resolver for the deleteSubscription field.
func (r *mutationResolver) DeleteSubscription(ctx context.Context, id string) (bool, error) {
	if r.SubscriptionRepo == nil {
		return false, fmt.Errorf("subscription repository not configured")
	}

	if err := r.SubscriptionRepo.Delete(ctx, id); err != nil {
		r.Logger.Error("failed to delete subscription",
			zap.String("id", id),
			zap.Error(err),
		)
		return false, err
	}

	r.Logger.Info("subscription deleted",
		zap.String("id", id),
	)

	return true, nil
}

// ToggleSubscription is the resolver for the toggleSubscription field.
func (r *mutationResolver) ToggleSubscription(ctx context.Context, input model.ToggleSubscriptionInput) (*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	subscription, err := r.SubscriptionRepo.GetByID(ctx, input.ID)
	if err != nil {
		r.Logger.Error("failed to get subscription",
			zap.String("id", input.ID),
			zap.Error(err),
		)
		return nil, err
	}

	if subscription == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	// Обновляем статус
	subscription.IsActive = input.IsActive

	if err := r.SubscriptionRepo.Update(ctx, subscription); err != nil {
		r.Logger.Error("failed to toggle subscription",
			zap.String("id", input.ID),
			zap.Bool("is_active", input.IsActive),
			zap.Error(err),
		)
		return nil, err
	}

	r.Logger.Info("subscription toggled",
		zap.String("id", subscription.ID),
		zap.Bool("is_active", subscription.IsActive),
	)

	return subscription, nil
}

// MySubscriptions is the resolver for the mySubscriptions field.
func (r *queryResolver) MySubscriptions(ctx context.Context) ([]*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	// Получаем userID из JWT context
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized: login required")
	}

	subscriptions, err := r.SubscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		r.Logger.Error("failed to get subscriptions by user id",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, err
	}

	return subscriptions, nil
}

// Subscription is the resolver for the subscription field.
func (r *queryResolver) Subscription(ctx context.Context, id string) (*model.EntitySubscription, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	subscription, err := r.SubscriptionRepo.GetByID(ctx, id)
	if err != nil {
		r.Logger.Error("failed to get subscription by id",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	return subscription, nil
}

// NotificationHistory is the resolver for the notificationHistory field.
func (r *queryResolver) NotificationHistory(ctx context.Context, subscriptionID string, limit *int, offset *int) ([]*model.NotificationLogEntry, error) {
	if r.SubscriptionRepo == nil {
		return nil, fmt.Errorf("subscription repository not configured")
	}

	l := 20
	if limit != nil {
		l = *limit
	}

	o := 0
	if offset != nil {
		o = *offset
	}

	history, err := r.SubscriptionRepo.GetNotificationHistory(ctx, subscriptionID, l, o)
	if err != nil {
		r.Logger.Error("failed to get notification history",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		return nil, err
	}

	return history, nil
}

// HasSubscription is the resolver for the hasSubscription field.
func (r *queryResolver) HasSubscription(ctx context.Context, entityType model.EntityType, entityID string) (bool, error) {
	if r.SubscriptionRepo == nil {
		return false, fmt.Errorf("subscription repository not configured")
	}

	// Получаем userID из JWT context
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("unauthorized: login required")
	}

	exists, err := r.SubscriptionRepo.HasSubscription(ctx, userID, string(entityType), entityID)
	if err != nil {
		r.Logger.Error("failed to check subscription existence",
			zap.String("user_id", userID),
			zap.String("entity_type", string(entityType)),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)
		return false, err
	}

	return exists, nil
}

// User is the resolver for the user field in EntitySubscription.
func (r *entitySubscriptionResolver) User(ctx context.Context, obj *model.EntitySubscription) (*model.User, error) {
	if r.UserRepo == nil {
		return nil, fmt.Errorf("user repository not configured")
	}

	// Получаем пользователя из БД
	user, err := r.UserRepo.GetByID(ctx, obj.UserID)
	if err != nil {
		r.Logger.Error("failed to get user",
			zap.String("user_id", obj.UserID),
			zap.Error(err),
		)
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Преобразуем в GraphQL модель
	gqlUser := &model.User{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		LastLoginAt:   user.LastLoginAt,
	}

	return gqlUser, nil
}

// Resolver types for ManualHandler - no generated interface registration
type mutationResolver struct{ *Resolver }
type entitySubscriptionResolver struct{ *Resolver }
