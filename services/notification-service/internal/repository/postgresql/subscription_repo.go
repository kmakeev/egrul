package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/egrul/notification-service/internal/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// SubscriptionRepository реализация для PostgreSQL
type SubscriptionRepository struct {
	db     *sql.DB
	schema string
	logger *zap.Logger
}

// NewSubscriptionRepository создает новый экземпляр SubscriptionRepository
func NewSubscriptionRepository(db *sql.DB, schema string, logger *zap.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		schema: schema,
		logger: logger,
	}
}

// GetByID получает подписку по ID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*model.EntitySubscription, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE id = $1
	`, r.schema)

	var sub model.EntitySubscription
	var changeFiltersJSON, channelsJSON []byte
	var lastNotifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserEmail,
		&sub.EntityType,
		&sub.EntityID,
		&sub.EntityName,
		&changeFiltersJSON,
		&channelsJSON,
		&sub.IsActive,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&lastNotifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Десериализация JSON полей
	sub.ChangeFilters, err = model.UnmarshalChangeFilters(changeFiltersJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal change filters: %w", err)
	}

	sub.NotificationChannels, err = model.UnmarshalNotificationChannels(channelsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification channels: %w", err)
	}

	if lastNotifiedAt.Valid {
		sub.LastNotifiedAt = &lastNotifiedAt.Time
	}

	return &sub, nil
}

// GetByEmail получает все подписки пользователя
func (r *SubscriptionRepository) GetByEmail(ctx context.Context, email string) ([]*model.EntitySubscription, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE user_email = $1
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.EntitySubscription

	for rows.Next() {
		var sub model.EntitySubscription
		var changeFiltersJSON, channelsJSON []byte
		var lastNotifiedAt sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.UserEmail,
			&sub.EntityType,
			&sub.EntityID,
			&sub.EntityName,
			&changeFiltersJSON,
			&channelsJSON,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&lastNotifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		sub.ChangeFilters, err = model.UnmarshalChangeFilters(changeFiltersJSON)
		if err != nil {
			r.logger.Warn("failed to unmarshal change filters", zap.Error(err))
			sub.ChangeFilters = make(map[string]bool)
		}

		sub.NotificationChannels, err = model.UnmarshalNotificationChannels(channelsJSON)
		if err != nil {
			r.logger.Warn("failed to unmarshal notification channels", zap.Error(err))
			sub.NotificationChannels = make(map[string]bool)
		}

		if lastNotifiedAt.Valid {
			sub.LastNotifiedAt = &lastNotifiedAt.Time
		}

		subscriptions = append(subscriptions, &sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

// GetByEntity получает подписки на конкретную сущность
func (r *SubscriptionRepository) GetByEntity(ctx context.Context, entityType, entityID string) ([]*model.EntitySubscription, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE entity_type = $1 AND entity_id = $2 AND is_active = true
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.EntitySubscription

	for rows.Next() {
		var sub model.EntitySubscription
		var changeFiltersJSON, channelsJSON []byte
		var lastNotifiedAt sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.UserEmail,
			&sub.EntityType,
			&sub.EntityID,
			&sub.EntityName,
			&changeFiltersJSON,
			&channelsJSON,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&lastNotifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		sub.ChangeFilters, err = model.UnmarshalChangeFilters(changeFiltersJSON)
		if err != nil {
			r.logger.Warn("failed to unmarshal change filters", zap.Error(err))
			sub.ChangeFilters = make(map[string]bool)
		}

		sub.NotificationChannels, err = model.UnmarshalNotificationChannels(channelsJSON)
		if err != nil {
			r.logger.Warn("failed to unmarshal notification channels", zap.Error(err))
			sub.NotificationChannels = make(map[string]bool)
		}

		if lastNotifiedAt.Valid {
			sub.LastNotifiedAt = &lastNotifiedAt.Time
		}

		subscriptions = append(subscriptions, &sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

// Create создает новую подписку
func (r *SubscriptionRepository) Create(ctx context.Context, subscription *model.EntitySubscription) error {
	// Генерируем ID если не задан
	if subscription.ID == "" {
		subscription.ID = uuid.New().String()
	}

	// Сериализация JSON полей
	changeFiltersJSON, err := model.MarshalChangeFilters(subscription.ChangeFilters)
	if err != nil {
		return fmt.Errorf("failed to marshal change filters: %w", err)
	}

	channelsJSON, err := model.MarshalNotificationChannels(subscription.NotificationChannels)
	if err != nil {
		return fmt.Errorf("failed to marshal notification channels: %w", err)
	}

	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	query := fmt.Sprintf(`
		INSERT INTO %s.entity_subscriptions (
			id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.schema)

	_, err = r.db.ExecContext(ctx, query,
		subscription.ID,
		subscription.UserEmail,
		subscription.EntityType,
		subscription.EntityID,
		subscription.EntityName,
		changeFiltersJSON,
		channelsJSON,
		subscription.IsActive,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	r.logger.Info("subscription created",
		zap.String("id", subscription.ID),
		zap.String("email", subscription.UserEmail),
		zap.String("entity_type", subscription.EntityType),
		zap.String("entity_id", subscription.EntityID),
	)

	return nil
}

// Update обновляет подписку
func (r *SubscriptionRepository) Update(ctx context.Context, subscription *model.EntitySubscription) error {
	changeFiltersJSON, err := model.MarshalChangeFilters(subscription.ChangeFilters)
	if err != nil {
		return fmt.Errorf("failed to marshal change filters: %w", err)
	}

	channelsJSON, err := model.MarshalNotificationChannels(subscription.NotificationChannels)
	if err != nil {
		return fmt.Errorf("failed to marshal notification channels: %w", err)
	}

	subscription.UpdatedAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s.entity_subscriptions
		SET change_filters = $1,
		    notification_channels = $2,
		    is_active = $3,
		    updated_at = $4
		WHERE id = $5
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query,
		changeFiltersJSON,
		channelsJSON,
		subscription.IsActive,
		subscription.UpdatedAt,
		subscription.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", subscription.ID)
	}

	r.logger.Info("subscription updated",
		zap.String("id", subscription.ID),
	)

	return nil
}

// Delete удаляет подписку
func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.entity_subscriptions
		WHERE id = $1
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", id)
	}

	r.logger.Info("subscription deleted",
		zap.String("id", id),
	)

	return nil
}

// UpdateLastNotified обновляет время последнего уведомления
func (r *SubscriptionRepository) UpdateLastNotified(ctx context.Context, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s.entity_subscriptions
		SET last_notified_at = $1,
		    updated_at = $2
		WHERE id = $3
	`, r.schema)

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last_notified_at: %w", err)
	}

	return nil
}
