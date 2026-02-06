package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
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
			id, user_id, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE id = $1
	`, r.schema)

	var sub model.EntitySubscription
	var entityType string
	var lastNotifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&entityType,
		&sub.EntityID,
		&sub.EntityName,
		&sub.ChangeFilters,
		&sub.NotificationChannels,
		&sub.IsActive,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&lastNotifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Конвертируем entityType обратно в uppercase для соответствия GraphQL enum
	sub.EntityType = model.EntityType(strings.ToUpper(entityType))

	if lastNotifiedAt.Valid {
		sub.LastNotifiedAt = &lastNotifiedAt.Time
	}

	return &sub, nil
}

// GetByUserID получает все подписки пользователя
func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID string) ([]*model.EntitySubscription, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_id, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.EntitySubscription

	for rows.Next() {
		var sub model.EntitySubscription
		var entityType string
		var lastNotifiedAt sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&entityType,
			&sub.EntityID,
			&sub.EntityName,
			&sub.ChangeFilters,
			&sub.NotificationChannels,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&lastNotifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		// Конвертируем entityType обратно в uppercase для соответствия GraphQL enum
		sub.EntityType = model.EntityType(strings.ToUpper(entityType))

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

// HasSubscription проверяет наличие подписки
func (r *SubscriptionRepository) HasSubscription(ctx context.Context, userID, entityType, entityID string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s.entity_subscriptions
			WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3
		)
	`, r.schema)

	// Конвертируем в lowercase для соответствия значениям в БД
	entityType = strings.ToLower(entityType)

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, entityType, entityID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check subscription existence: %w", err)
	}

	return exists, nil
}

// Create создает новую подписку
func (r *SubscriptionRepository) Create(ctx context.Context, subscription *model.EntitySubscription) error {
	// Генерируем ID если не задан
	if subscription.ID == "" {
		subscription.ID = uuid.New().String()
	}

	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	query := fmt.Sprintf(`
		INSERT INTO %s.entity_subscriptions (
			id, user_id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, r.schema)

	// Конвертируем EntityType в lowercase для соответствия check constraint
	entityType := strings.ToLower(string(subscription.EntityType))

	_, err := r.db.ExecContext(ctx, query,
		subscription.ID,
		subscription.UserID,
		subscription.UserEmail,
		entityType,
		subscription.EntityID,
		subscription.EntityName,
		subscription.ChangeFilters,
		subscription.NotificationChannels,
		subscription.IsActive,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	r.logger.Info("subscription created",
		zap.String("id", subscription.ID),
		zap.String("user_id", subscription.UserID),
		zap.String("entity_type", string(subscription.EntityType)),
		zap.String("entity_id", subscription.EntityID),
	)

	return nil
}

// Update обновляет подписку
func (r *SubscriptionRepository) Update(ctx context.Context, subscription *model.EntitySubscription) error {
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
		subscription.ChangeFilters,
		subscription.NotificationChannels,
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

// GetNotificationHistory получает историю уведомлений для подписки (для GraphQL)
func (r *SubscriptionRepository) GetNotificationHistory(ctx context.Context, subscriptionID string, limit, offset int) ([]*model.NotificationLogEntry, error) {
	query := fmt.Sprintf(`
		SELECT
			id, subscription_id, change_event_id, user_id, channel,
			status, sent_at, error_message, created_at
		FROM %s.notification_log
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, subscriptionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification log: %w", err)
	}
	defer rows.Close()

	var entries []*model.NotificationLogEntry

	for rows.Next() {
		var entry model.NotificationLogEntry
		var status string
		var sentAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.SubscriptionID,
			&entry.ChangeEventID,
			&entry.UserID,
			&entry.Channel,
			&status,
			&sentAt,
			&errorMessage,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification log entry: %w", err)
		}

		entry.Status = model.NotificationStatus(status)

		if sentAt.Valid {
			entry.SentAt = &sentAt.Time
		}

		if errorMessage.Valid {
			entry.ErrorMessage = &errorMessage.String
		}

		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification log: %w", err)
	}

	return entries, nil
}

// EntitySubscription - структура для работы с Hub (без GraphQL model dependencies)
type EntitySubscription struct {
	ID                   string
	UserEmail            string
	EntityType           string
	EntityID             string
	EntityName           string
	ChangeFilters        map[string]bool
	NotificationChannels map[string]bool
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastNotifiedAt       *time.Time
}

// GetActiveSubscriptionsForEntity получает активные подписки для сущности (для Notification Hub)
func (r *SubscriptionRepository) GetActiveSubscriptionsForEntity(ctx context.Context, entityType, entityID string) ([]EntitySubscription, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_email, entity_type, entity_id, entity_name,
			change_filters, notification_channels, is_active,
			created_at, updated_at, last_notified_at
		FROM %s.entity_subscriptions
		WHERE entity_type = $1 AND entity_id = $2 AND is_active = TRUE
	`, r.schema)

	entityType = strings.ToLower(entityType)

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []EntitySubscription

	for rows.Next() {
		var sub EntitySubscription
		var lastNotifiedAt sql.NullTime
		var changeFiltersJSON []byte
		var notificationChannelsJSON []byte

		err := rows.Scan(
			&sub.ID,
			&sub.UserEmail,
			&sub.EntityType,
			&sub.EntityID,
			&sub.EntityName,
			&changeFiltersJSON,
			&notificationChannelsJSON,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&lastNotifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		// Десериализуем JSON
		if err := json.Unmarshal(changeFiltersJSON, &sub.ChangeFilters); err != nil {
			r.logger.Warn("failed to unmarshal change_filters", zap.Error(err))
			sub.ChangeFilters = make(map[string]bool)
		}

		if err := json.Unmarshal(notificationChannelsJSON, &sub.NotificationChannels); err != nil {
			r.logger.Warn("failed to unmarshal notification_channels", zap.Error(err))
			sub.NotificationChannels = make(map[string]bool)
		}

		if lastNotifiedAt.Valid {
			sub.LastNotifiedAt = &lastNotifiedAt.Time
		}

		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

// NotificationLogEntry - структура для истории уведомлений (упрощенная)
type NotificationLogEntry struct {
	ID            string
	EntityType    string
	EntityID      string
	EntityName    string
	ChangeType    string
	FieldName     string
	OldValue      string
	NewValue      string
	DetectedAt    time.Time
	IsRead        bool
	ReadAt        *time.Time
	CreatedAt     time.Time
}

// GetNotificationHistoryByEmail получает историю уведомлений пользователя по email (для Notification Hub)
func (r *SubscriptionRepository) GetNotificationHistoryByEmail(ctx context.Context, email string, limit, offset int) ([]NotificationLogEntry, error) {
	query := fmt.Sprintf(`
		SELECT
			id, entity_type, entity_id, entity_name,
			change_type, field_name, old_value, new_value,
			detected_at, is_read, read_at, created_at
		FROM %s.notification_log
		WHERE recipient = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, email, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification history: %w", err)
	}
	defer rows.Close()

	var entries []NotificationLogEntry

	for rows.Next() {
		var entry NotificationLogEntry
		var readAt sql.NullTime

		err := rows.Scan(
			&entry.ID,
			&entry.EntityType,
			&entry.EntityID,
			&entry.EntityName,
			&entry.ChangeType,
			&entry.FieldName,
			&entry.OldValue,
			&entry.NewValue,
			&entry.DetectedAt,
			&entry.IsRead,
			&readAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification entry: %w", err)
		}

		if readAt.Valid {
			entry.ReadAt = &readAt.Time
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification history: %w", err)
	}

	return entries, nil
}

// MarkNotificationAsRead отмечает уведомление как прочитанное
func (r *SubscriptionRepository) MarkNotificationAsRead(ctx context.Context, notificationID, email string) error {
	query := fmt.Sprintf(`
		UPDATE %s.notification_log
		SET is_read = TRUE, read_at = $1
		WHERE id = $2 AND recipient = $3 AND is_read = FALSE
	`, r.schema)

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, notificationID, email)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.logger.Debug("notification marked as read",
		zap.String("notification_id", notificationID),
		zap.String("email", email),
	)

	return nil
}

// MarkAllNotificationsAsRead отмечает все уведомления пользователя как прочитанные
func (r *SubscriptionRepository) MarkAllNotificationsAsRead(ctx context.Context, email string) (int64, error) {
	query := fmt.Sprintf(`
		UPDATE %s.notification_log
		SET is_read = TRUE, read_at = $1
		WHERE recipient = $2 AND is_read = FALSE
	`, r.schema)

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, email)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		r.logger.Debug("all notifications marked as read",
			zap.String("email", email),
			zap.Int64("count", rowsAffected),
		)
	}

	return rowsAffected, nil
}
