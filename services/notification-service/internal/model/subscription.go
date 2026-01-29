package model

import (
	"encoding/json"
	"time"
)

// EntitySubscription представляет подписку пользователя на изменения сущности
type EntitySubscription struct {
	ID                   string                 `json:"id"`
	UserEmail            string                 `json:"user_email"`
	EntityType           string                 `json:"entity_type"` // "company" или "entrepreneur"
	EntityID             string                 `json:"entity_id"`   // OGRN или OGRNIP
	EntityName           string                 `json:"entity_name"`
	ChangeFilters        map[string]bool        `json:"change_filters"`        // Какие типы изменений отслеживать
	NotificationChannels map[string]bool        `json:"notification_channels"` // Через какие каналы уведомлять
	IsActive             bool                   `json:"is_active"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	LastNotifiedAt       *time.Time             `json:"last_notified_at,omitempty"`
}

// ChangeFilter типы изменений для фильтрации
type ChangeFilter struct {
	Status     bool `json:"status"`
	Director   bool `json:"director"`
	Founders   bool `json:"founders"`
	Address    bool `json:"address"`
	Capital    bool `json:"capital"`
	Activities bool `json:"activities"`
}

// NotificationChannel каналы уведомлений
type NotificationChannel struct {
	Email bool `json:"email"`
}

// ShouldNotify проверяет, нужно ли отправлять уведомление для данного типа изменения
func (s *EntitySubscription) ShouldNotify(changeType string) bool {
	if !s.IsActive {
		return false
	}

	// Проверяем фильтры
	if s.ChangeFilters == nil {
		return true // Если фильтры не заданы, уведомляем обо всех изменениях
	}

	enabled, ok := s.ChangeFilters[changeType]
	if !ok {
		return true // Если тип не указан в фильтрах, уведомляем
	}

	return enabled
}

// HasEmailChannel проверяет, включен ли канал Email
func (s *EntitySubscription) HasEmailChannel() bool {
	if s.NotificationChannels == nil {
		return false
	}

	return s.NotificationChannels["email"]
}

// MarshalChangeFilters сериализует фильтры изменений в JSON
func MarshalChangeFilters(filters map[string]bool) ([]byte, error) {
	return json.Marshal(filters)
}

// UnmarshalChangeFilters десериализует фильтры изменений из JSON
func UnmarshalChangeFilters(data []byte) (map[string]bool, error) {
	var filters map[string]bool
	err := json.Unmarshal(data, &filters)
	return filters, err
}

// MarshalNotificationChannels сериализует каналы уведомлений в JSON
func MarshalNotificationChannels(channels map[string]bool) ([]byte, error) {
	return json.Marshal(channels)
}

// UnmarshalNotificationChannels десериализует каналы уведомлений из JSON
func UnmarshalNotificationChannels(data []byte) (map[string]bool, error) {
	var channels map[string]bool
	err := json.Unmarshal(data, &channels)
	return channels, err
}
