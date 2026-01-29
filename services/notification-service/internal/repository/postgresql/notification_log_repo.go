package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/egrul/notification-service/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationLogRepository реализация для PostgreSQL
type NotificationLogRepository struct {
	db     *sql.DB
	schema string
	logger *zap.Logger
}

// NewNotificationLogRepository создает новый экземпляр NotificationLogRepository
func NewNotificationLogRepository(db *sql.DB, schema string, logger *zap.Logger) *NotificationLogRepository {
	return &NotificationLogRepository{
		db:     db,
		schema: schema,
		logger: logger,
	}
}

// Save сохраняет запись об отправленном уведомлении
func (r *NotificationLogRepository) Save(ctx context.Context, notification *model.Notification) error {
	// Генерируем ID если не задан
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.notification_log (
			id, subscription_id, change_event_id, user_email, channel,
			status, sent_at, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, r.schema)

	var sentAt sql.NullTime
	if notification.SentAt != nil {
		sentAt = sql.NullTime{Time: *notification.SentAt, Valid: true}
	}

	var errorMessage sql.NullString
	if notification.ErrorMessage != "" {
		errorMessage = sql.NullString{String: notification.ErrorMessage, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.SubscriptionID,
		notification.ChangeEvent.ChangeID,
		notification.UserEmail,
		notification.Channel,
		string(notification.Status),
		sentAt,
		errorMessage,
		notification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save notification log: %w", err)
	}

	r.logger.Debug("notification log saved",
		zap.String("id", notification.ID),
		zap.String("subscription_id", notification.SubscriptionID),
		zap.String("change_event_id", notification.ChangeEvent.ChangeID),
		zap.String("status", string(notification.Status)),
	)

	return nil
}

// GetBySubscription получает историю уведомлений для подписки
func (r *NotificationLogRepository) GetBySubscription(ctx context.Context, subscriptionID string, limit, offset int) ([]*model.Notification, error) {
	query := fmt.Sprintf(`
		SELECT
			id, subscription_id, change_event_id, user_email, channel,
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

	var notifications []*model.Notification

	for rows.Next() {
		var n model.Notification
		var changeEventID string
		var sentAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&n.ID,
			&n.SubscriptionID,
			&changeEventID,
			&n.UserEmail,
			&n.Channel,
			&n.Status,
			&sentAt,
			&errorMessage,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification log: %w", err)
		}

		if sentAt.Valid {
			n.SentAt = &sentAt.Time
		}

		if errorMessage.Valid {
			n.ErrorMessage = errorMessage.String
		}

		// Создаем минимальный ChangeEvent с ID
		n.ChangeEvent = &model.ChangeEvent{
			ChangeID: changeEventID,
		}

		notifications = append(notifications, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification log: %w", err)
	}

	return notifications, nil
}

// GetByChangeEvent получает уведомления для конкретного события
func (r *NotificationLogRepository) GetByChangeEvent(ctx context.Context, changeEventID string) ([]*model.Notification, error) {
	query := fmt.Sprintf(`
		SELECT
			id, subscription_id, change_event_id, user_email, channel,
			status, sent_at, error_message, created_at
		FROM %s.notification_log
		WHERE change_event_id = $1
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, changeEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification log: %w", err)
	}
	defer rows.Close()

	var notifications []*model.Notification

	for rows.Next() {
		var n model.Notification
		var changeEventID string
		var sentAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&n.ID,
			&n.SubscriptionID,
			&changeEventID,
			&n.UserEmail,
			&n.Channel,
			&n.Status,
			&sentAt,
			&errorMessage,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification log: %w", err)
		}

		if sentAt.Valid {
			n.SentAt = &sentAt.Time
		}

		if errorMessage.Valid {
			n.ErrorMessage = errorMessage.String
		}

		n.ChangeEvent = &model.ChangeEvent{
			ChangeID: changeEventID,
		}

		notifications = append(notifications, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notification log: %w", err)
	}

	return notifications, nil
}

// CheckDuplicate проверяет, было ли уже отправлено уведомление для данного события и подписки
func (r *NotificationLogRepository) CheckDuplicate(ctx context.Context, subscriptionID, changeEventID string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s.notification_log
			WHERE subscription_id = $1 AND change_event_id = $2 AND status = 'sent'
		)
	`, r.schema)

	var exists bool
	err := r.db.QueryRowContext(ctx, query, subscriptionID, changeEventID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return exists, nil
}
