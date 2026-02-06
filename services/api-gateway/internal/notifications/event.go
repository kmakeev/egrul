package notifications

import (
	"time"
)

// NotificationEvent представляет событие уведомления от Kafka
type NotificationEvent struct {
	ID            string    `json:"id"`              // Уникальный ID уведомления (из ClickHouse change_id)
	Type          string    `json:"type"`            // Тип события (всегда "change_detected")
	EntityType    string    `json:"entity_type"`     // "company" или "entrepreneur"
	EntityID      string    `json:"entity_id"`       // OGRN или OGRNIP
	EntityName    string    `json:"entity_name"`     // Название организации/ФИО ИП
	ChangeType    string    `json:"change_type"`     // Тип изменения (status, director, founders, address, capital, activities)
	FieldName     string    `json:"field_name"`      // Название поля
	OldValue      string    `json:"old_value"`       // Старое значение
	NewValue      string    `json:"new_value"`       // Новое значение
	IsSignificant bool      `json:"is_significant"`  // Важность изменения
	Timestamp     time.Time `json:"timestamp"`       // Время детектирования
	RegionCode    string    `json:"region_code"`     // Код региона
}

// ChangeEvent - структура события из Kafka (соответствует модели в change-detection-service)
type ChangeEvent struct {
	ChangeID      string    `json:"change_id"`
	EntityType    string    `json:"entity_type"`
	EntityID      string    `json:"entity_id"`
	EntityName    string    `json:"entity_name"`
	ChangeType    string    `json:"change_type"`
	FieldName     string    `json:"field_name"`
	OldValue      string    `json:"old_value"`
	NewValue      string    `json:"new_value"`
	IsSignificant bool      `json:"is_significant"`
	RegionCode    string    `json:"region_code"`
	DetectedAt    time.Time `json:"detected_at"`
}

// ToNotificationEvent преобразует ChangeEvent из Kafka в NotificationEvent для SSE
func (ce *ChangeEvent) ToNotificationEvent() *NotificationEvent {
	return &NotificationEvent{
		ID:            ce.ChangeID,
		Type:          "change_detected",
		EntityType:    ce.EntityType,
		EntityID:      ce.EntityID,
		EntityName:    ce.EntityName,
		ChangeType:    ce.ChangeType,
		FieldName:     ce.FieldName,
		OldValue:      ce.OldValue,
		NewValue:      ce.NewValue,
		IsSignificant: ce.IsSignificant,
		Timestamp:     ce.DetectedAt,
		RegionCode:    ce.RegionCode,
	}
}
