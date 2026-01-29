package channels

import (
	"context"

	"github.com/egrul/notification-service/internal/model"
)

// NotificationChannel интерфейс для каналов отправки уведомлений
type NotificationChannel interface {
	// Send отправляет уведомление
	Send(ctx context.Context, notification *model.Notification) error

	// Name возвращает название канала
	Name() string

	// Close закрывает соединения и освобождает ресурсы
	Close() error
}
