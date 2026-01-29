package model

import "time"

// Entrepreneur представляет упрощенную модель индивидуального предпринимателя
// Содержит только те поля, которые отслеживаются системой
type Entrepreneur struct {
	// Основная идентификация
	OGRNIP     string `json:"ogrnip"`      // Основной государственный регистрационный номер ИП
	INN        string `json:"inn"`         // Идентификационный номер налогоплательщика
	FullName   string `json:"full_name"`   // Полное ФИО
	RegionCode string `json:"region_code"` // Код региона

	// Статус
	Status         string     `json:"status"`           // Статус (ДЕЙСТВУЮЩИЙ, ПРЕКРАТИЛ ДЕЯТЕЛЬНОСТЬ)
	StatusDate     *time.Time `json:"status_date"`      // Дата изменения статуса
	TerminationDate *time.Time `json:"termination_date"` // Дата прекращения деятельности

	// Адрес
	AddressFull       string `json:"address_full"`        // Полный адрес
	AddressPostalCode string `json:"address_postal_code"` // Почтовый индекс
	AddressRegion     string `json:"address_region"`      // Регион
	AddressCity       string `json:"address_city"`        // Город
	AddressStreet     string `json:"address_street"`      // Улица
	AddressHouse      string `json:"address_house"`       // Дом

	// Виды деятельности (ОКВЭД)
	MainOKVED       string   `json:"main_okved"`        // Основной ОКВЭД
	AdditionalOKVED []string `json:"additional_okved"`  // Дополнительные ОКВЭД

	// Лицензии (кол-во, для детектирования изменений)
	LicensesCount int `json:"licenses_count"`

	// Метаданные
	RegistrationDate time.Time `json:"registration_date"` // Дата регистрации
	LastUpdate       time.Time `json:"last_update"`       // Дата последнего обновления

	// Дополнительная информация
	CitizenshipCode string `json:"citizenship_code"` // Код гражданства
	Gender          string `json:"gender"`           // Пол (М/Ж)
}

// IsActive проверяет, действующий ли ИП
func (e *Entrepreneur) IsActive() bool {
	return e.Status == "ДЕЙСТВУЮЩИЙ"
}

// IsTerminated проверяет, прекратил ли ИП деятельность
func (e *Entrepreneur) IsTerminated() bool {
	return e.Status == "ПРЕКРАТИЛ ДЕЯТЕЛЬНОСТЬ"
}

// HasOKVED проверяет, есть ли у ИП указанный ОКВЭД
func (e *Entrepreneur) HasOKVED(okved string) bool {
	if e.MainOKVED == okved {
		return true
	}
	for _, kod := range e.AdditionalOKVED {
		if kod == okved {
			return true
		}
	}
	return false
}

// GetFullAddress возвращает полный адрес ИП
func (e *Entrepreneur) GetFullAddress() string {
	if e.AddressFull != "" {
		return e.AddressFull
	}
	// Собираем адрес из компонентов
	addr := ""
	if e.AddressPostalCode != "" {
		addr += e.AddressPostalCode + ", "
	}
	if e.AddressRegion != "" {
		addr += e.AddressRegion + ", "
	}
	if e.AddressCity != "" {
		addr += e.AddressCity + ", "
	}
	if e.AddressStreet != "" {
		addr += e.AddressStreet + ", "
	}
	if e.AddressHouse != "" {
		addr += "д. " + e.AddressHouse
	}
	return addr
}
