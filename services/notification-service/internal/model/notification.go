package model

import "time"

// ChangeEvent представляет событие изменения (из Kafka)
type ChangeEvent struct {
	ChangeID      string    `json:"change_id"`
	EntityType    string    `json:"entity_type"` // "company" или "entrepreneur"
	EntityID      string    `json:"entity_id"`   // OGRN или OGRNIP
	EntityName    string    `json:"entity_name"`
	ChangeType    string    `json:"change_type"`
	FieldName     string    `json:"field_name"`
	OldValue      string    `json:"old_value"`
	NewValue      string    `json:"new_value"`
	IsSignificant bool      `json:"is_significant"`
	RegionCode    string    `json:"region_code"`
	DetectedAt    time.Time `json:"detected_at"`
}

// Notification представляет уведомление для отправки
type Notification struct {
	ID             string
	SubscriptionID string
	ChangeEvent    *ChangeEvent
	UserEmail      string
	Channel        string // "email", "websocket", "telegram"
	Status         NotificationStatus
	SentAt         *time.Time
	ErrorMessage   string
	CreatedAt      time.Time
}

// NotificationStatus статус уведомления
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// EmailNotificationData данные для email шаблона
type EmailNotificationData struct {
	// Данные компании/ИП
	EntityType string
	EntityID   string
	EntityName string

	// Данные изменения
	ChangeType    string
	FieldName     string
	OldValue      string
	NewValue      string
	IsSignificant bool
	DetectedAt    time.Time

	// Ссылки
	EntityURL       string
	UnsubscribeURL  string
	SettingsURL     string

	// Локализация типов изменений
	ChangeTypeLabel string
	FieldNameLabel  string
}

// GetChangeTypeLabel возвращает локализованное название типа изменения
func GetChangeTypeLabel(changeType string) string {
	labels := map[string]string{
		"status":             "Статус",
		"director":           "Руководитель",
		"founder_added":      "Добавлен учредитель",
		"founder_removed":    "Удален учредитель",
		"founder_share":      "Изменена доля учредителя",
		"address":            "Юридический адрес",
		"capital":            "Уставный капитал",
		"activity_added":     "Добавлен вид деятельности (ОКВЭД)",
		"activity_removed":   "Удален вид деятельности (ОКВЭД)",
		"activity_main":      "Основной вид деятельности",
		"license_added":      "Добавлена лицензия",
		"license_removed":    "Удалена лицензия",
		"branch_added":       "Добавлен филиал",
		"branch_removed":     "Удален филиал",
	}

	if label, ok := labels[changeType]; ok {
		return label
	}
	return changeType
}

// GetFieldNameLabel возвращает локализованное название поля
func GetFieldNameLabel(fieldName string) string {
	labels := map[string]string{
		"status":             "Статус",
		"director_fio":       "ФИО руководителя",
		"director_inn":       "ИНН руководителя",
		"founder_name":       "Наименование учредителя",
		"founder_share":      "Доля в капитале",
		"address":            "Юридический адрес",
		"capital_value":      "Размер капитала",
		"okved_code":         "Код ОКВЭД",
		"okved_name":         "Наименование ОКВЭД",
		"license_number":     "Номер лицензии",
		"license_date":       "Дата лицензии",
		"branch_name":        "Наименование филиала",
		"branch_address":     "Адрес филиала",
	}

	if label, ok := labels[fieldName]; ok {
		return label
	}
	return fieldName
}

// FormatValue форматирует значение для отображения
func FormatValue(value string, fieldName string) string {
	if value == "" {
		return "не указано"
	}

	// Можно добавить специальную обработку для определенных полей
	// Например, форматирование дат, сумм и т.д.

	return value
}
