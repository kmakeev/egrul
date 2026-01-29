package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// EntitySubscription представляет подписку пользователя на изменения сущности
type EntitySubscription struct {
	ID                   string                 `json:"id"`
	UserID               string                 `json:"userId"`
	UserEmail            string                 `json:"-"` // Внутреннее поле для совместимости с БД (не возвращается в API)
	User                 *User                  `json:"user"`
	EntityType           EntityType             `json:"entityType"`
	EntityID             string                 `json:"entityId"`
	EntityName           string                 `json:"entityName"`
	ChangeFilters        *ChangeFilters         `json:"changeFilters"`
	NotificationChannels *NotificationChannels  `json:"notificationChannels"`
	IsActive             bool                   `json:"isActive"`
	CreatedAt            time.Time              `json:"createdAt"`
	UpdatedAt            time.Time              `json:"updatedAt"`
	LastNotifiedAt       *time.Time             `json:"lastNotifiedAt,omitempty"`
}

// ChangeFilters фильтры типов изменений
type ChangeFilters struct {
	Status     bool `json:"status"`
	Director   bool `json:"director"`
	Founders   bool `json:"founders"`
	Address    bool `json:"address"`
	Capital    bool `json:"capital"`
	Activities bool `json:"activities"`
}

// NotificationChannels каналы уведомлений
type NotificationChannels struct {
	Email bool `json:"email"`
}

// NotificationStatus статус отправки уведомления
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSent    NotificationStatus = "SENT"
	NotificationStatusFailed  NotificationStatus = "FAILED"
)

func (n NotificationStatus) IsValid() bool {
	switch n {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusFailed:
		return true
	}
	return false
}

func (n NotificationStatus) String() string {
	return string(n)
}

// NotificationLogEntry запись лога уведомления
type NotificationLogEntry struct {
	ID             string             `json:"id"`
	SubscriptionID string             `json:"subscriptionId"`
	ChangeEventID  string             `json:"changeEventId"`
	UserID         string             `json:"userId"`
	Channel        string             `json:"channel"`
	Status         NotificationStatus `json:"status"`
	SentAt         *time.Time         `json:"sentAt,omitempty"`
	ErrorMessage   *string            `json:"errorMessage,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
}

// Input типы

// CreateSubscriptionInput входные данные для создания подписки
type CreateSubscriptionInput struct {
	EntityType           EntityType                `json:"entityType"`
	EntityID             string                    `json:"entityId"`
	EntityName           string                    `json:"entityName"`
	ChangeFilters        *ChangeFiltersInput       `json:"changeFilters,omitempty"`
	NotificationChannels *NotificationChannelsInput `json:"notificationChannels,omitempty"`
}

// ChangeFiltersInput входные данные для фильтров изменений
type ChangeFiltersInput struct {
	Status     *bool `json:"status,omitempty"`
	Director   *bool `json:"director,omitempty"`
	Founders   *bool `json:"founders,omitempty"`
	Address    *bool `json:"address,omitempty"`
	Capital    *bool `json:"capital,omitempty"`
	Activities *bool `json:"activities,omitempty"`
}

// NotificationChannelsInput входные данные для каналов уведомлений
type NotificationChannelsInput struct {
	Email *bool `json:"email,omitempty"`
}

// UpdateSubscriptionFiltersInput входные данные для обновления фильтров
type UpdateSubscriptionFiltersInput struct {
	ID            string              `json:"id"`
	ChangeFilters *ChangeFiltersInput `json:"changeFilters"`
}

// UpdateSubscriptionChannelsInput входные данные для обновления каналов
type UpdateSubscriptionChannelsInput struct {
	ID                   string                     `json:"id"`
	NotificationChannels *NotificationChannelsInput `json:"notificationChannels"`
}

// ToggleSubscriptionInput входные данные для переключения статуса
type ToggleSubscriptionInput struct {
	ID       string `json:"id"`
	IsActive bool   `json:"isActive"`
}

// Helper методы для работы с JSON в PostgreSQL

// Value implements driver.Valuer для ChangeFilters
func (c ChangeFilters) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner для ChangeFilters
func (c *ChangeFilters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal ChangeFilters value: %v", value)
	}

	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer для NotificationChannels
func (n NotificationChannels) Value() (driver.Value, error) {
	return json.Marshal(n)
}

// Scan implements sql.Scanner для NotificationChannels
func (n *NotificationChannels) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal NotificationChannels value: %v", value)
	}

	return json.Unmarshal(bytes, n)
}

// ToChangeFilters конвертирует input в ChangeFilters
func (i *ChangeFiltersInput) ToChangeFilters() *ChangeFilters {
	if i == nil {
		// Значения по умолчанию - все включено
		return &ChangeFilters{
			Status:     true,
			Director:   true,
			Founders:   true,
			Address:    true,
			Capital:    true,
			Activities: true,
		}
	}

	filters := &ChangeFilters{
		Status:     true,
		Director:   true,
		Founders:   true,
		Address:    true,
		Capital:    true,
		Activities: true,
	}

	if i.Status != nil {
		filters.Status = *i.Status
	}
	if i.Director != nil {
		filters.Director = *i.Director
	}
	if i.Founders != nil {
		filters.Founders = *i.Founders
	}
	if i.Address != nil {
		filters.Address = *i.Address
	}
	if i.Capital != nil {
		filters.Capital = *i.Capital
	}
	if i.Activities != nil {
		filters.Activities = *i.Activities
	}

	return filters
}

// ToNotificationChannels конвертирует input в NotificationChannels
func (i *NotificationChannelsInput) ToNotificationChannels() *NotificationChannels {
	if i == nil {
		// По умолчанию - только email
		return &NotificationChannels{
			Email: true,
		}
	}

	channels := &NotificationChannels{
		Email: true,
	}

	if i.Email != nil {
		channels.Email = *i.Email
	}

	return channels
}
