package model

import "time"

// ChangeType представляет тип изменения в данных компании/ИП
type ChangeType string

const (
	// Изменения компаний
	ChangeTypeStatus         ChangeType = "status"           // Изменение статуса (ликвидация, реорганизация)
	ChangeTypeDirector       ChangeType = "director"         // Смена руководителя
	ChangeTypeFounderAdded   ChangeType = "founder_added"    // Добавление учредителя
	ChangeTypeFounderRemoved ChangeType = "founder_removed"  // Удаление учредителя
	ChangeTypeFounderShare   ChangeType = "founder_share"    // Изменение доли учредителя
	ChangeTypeAddress        ChangeType = "address"          // Изменение адреса
	ChangeTypeCapital        ChangeType = "capital"          // Изменение уставного капитала
	ChangeTypeActivityAdded  ChangeType = "activity_added"   // Добавление вида деятельности (ОКВЭД)
	ChangeTypeActivityRemoved ChangeType = "activity_removed" // Удаление вида деятельности
	ChangeTypeLicenseAdded   ChangeType = "license_added"    // Добавление лицензии
	ChangeTypeLicenseRevoked ChangeType = "license_revoked"  // Отзыв лицензии
	ChangeTypeBranchAdded    ChangeType = "branch_added"     // Добавление филиала
	ChangeTypeBranchClosed   ChangeType = "branch_closed"    // Закрытие филиала

	// Изменения ИП
	ChangeTypeIPStatus       ChangeType = "ip_status"        // Изменение статуса ИП
	ChangeTypeIPAddress      ChangeType = "ip_address"       // Изменение адреса ИП
	ChangeTypeIPActivity     ChangeType = "ip_activity"      // Изменение вида деятельности ИП
)

// ChangeEvent представляет событие изменения в данных
type ChangeEvent struct {
	// Идентификация
	ChangeID   string     `json:"change_id"`   // UUID изменения
	EntityType string     `json:"entity_type"` // "company" или "entrepreneur"
	EntityID   string     `json:"entity_id"`   // ОГРН или ОГРНИП
	EntityName string     `json:"entity_name"` // Наименование организации

	// Тип изменения
	ChangeType ChangeType `json:"change_type"` // Тип изменения из констант выше
	FieldName  string     `json:"field_name"`  // Название поля (для логирования)

	// Значения
	OldValue string `json:"old_value"` // Старое значение (JSON)
	NewValue string `json:"new_value"` // Новое значение (JSON)

	// Метаданные
	IsSignificant bool      `json:"is_significant"` // Флаг значимости изменения
	DetectedAt    time.Time `json:"detected_at"`    // Время детектирования
	Description   string    `json:"description"`    // Человекочитаемое описание изменения

	// Дополнительная информация
	RegionCode string `json:"region_code,omitempty"` // Код региона
	INN        string `json:"inn,omitempty"`         // ИНН организации
}

// IsCompany проверяет, относится ли событие к компании
func (e *ChangeEvent) IsCompany() bool {
	return e.EntityType == "company"
}

// IsEntrepreneur проверяет, относится ли событие к ИП
func (e *ChangeEvent) IsEntrepreneur() bool {
	return e.EntityType == "entrepreneur"
}

// Validate проверяет корректность события
func (e *ChangeEvent) Validate() error {
	if e.EntityType != "company" && e.EntityType != "entrepreneur" {
		return ErrInvalidEntityType
	}
	if e.EntityID == "" {
		return ErrEmptyEntityID
	}
	if e.ChangeType == "" {
		return ErrEmptyChangeType
	}
	return nil
}

// Errors
var (
	ErrInvalidEntityType = &ValidationError{Field: "entity_type", Message: "must be 'company' or 'entrepreneur'"}
	ErrEmptyEntityID     = &ValidationError{Field: "entity_id", Message: "cannot be empty"}
	ErrEmptyChangeType   = &ValidationError{Field: "change_type", Message: "cannot be empty"}
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
