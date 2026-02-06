package notifications

import (
	"sync"
	"time"
)

// Client представляет SSE клиента, подключенного к Notification Hub
type Client struct {
	Email    string                    // Email пользователя
	UserID   string                    // ID пользователя из JWT (опционально)
	Messages chan *NotificationEvent   // Канал для отправки уведомлений клиенту
	Done     chan struct{}             // Канал для сигнализации завершения
	LastSeen time.Time                 // Время последнего heartbeat
	closeOnce sync.Once                // Защита от повторного закрытия
}

// NewClient создает нового SSE клиента
func NewClient(email, userID string, bufferSize int) *Client {
	return &Client{
		Email:    email,
		UserID:   userID,
		Messages: make(chan *NotificationEvent, bufferSize),
		Done:     make(chan struct{}),
		LastSeen: time.Now(),
	}
}

// Close закрывает клиента и освобождает ресурсы
// Защищено от повторного вызова через sync.Once
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.Done)
		close(c.Messages)
	})
}

// Send отправляет уведомление клиенту (неблокирующая отправка)
func (c *Client) Send(event *NotificationEvent) bool {
	select {
	case c.Messages <- event:
		return true
	default:
		// Буфер переполнен, пропускаем уведомление
		return false
	}
}

// UpdateLastSeen обновляет время последней активности
func (c *Client) UpdateLastSeen() {
	c.LastSeen = time.Now()
}
