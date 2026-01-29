package repository

import (
	"context"

	"github.com/egrul/notification-service/internal/model"
)

// SubscriptionRepository интерфейс для работы с подписками
type SubscriptionRepository interface {
	// GetByID получает подписку по ID
	GetByID(ctx context.Context, id string) (*model.EntitySubscription, error)

	// GetByEmail получает все подписки пользователя
	GetByEmail(ctx context.Context, email string) ([]*model.EntitySubscription, error)

	// GetByEntity получает подписки на конкретную сущность
	GetByEntity(ctx context.Context, entityType, entityID string) ([]*model.EntitySubscription, error)

	// Create создает новую подписку
	Create(ctx context.Context, subscription *model.EntitySubscription) error

	// Update обновляет подписку
	Update(ctx context.Context, subscription *model.EntitySubscription) error

	// Delete удаляет подписку
	Delete(ctx context.Context, id string) error

	// UpdateLastNotified обновляет время последнего уведомления
	UpdateLastNotified(ctx context.Context, id string) error
}

// NotificationLogRepository интерфейс для работы с логом уведомлений
type NotificationLogRepository interface {
	// Save сохраняет запись об отправленном уведомлении
	Save(ctx context.Context, notification *model.Notification) error

	// GetBySubscription получает историю уведомлений для подписки
	GetBySubscription(ctx context.Context, subscriptionID string, limit, offset int) ([]*model.Notification, error)

	// GetByChangeEvent получает уведомления для конкретного события
	GetByChangeEvent(ctx context.Context, changeEventID string) ([]*model.Notification, error)

	// CheckDuplicate проверяет, было ли уже отправлено уведомление для данного события и подписки
	CheckDuplicate(ctx context.Context, subscriptionID, changeEventID string) (bool, error)
}
