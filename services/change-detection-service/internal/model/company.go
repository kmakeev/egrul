package model

import "time"

// Company представляет упрощенную модель компании для сравнения изменений
// Содержит только те поля, которые отслеживаются системой
type Company struct {
	// Основная идентификация
	OGRN       string `json:"ogrn"`        // Основной государственный регистрационный номер
	INN        string `json:"inn"`         // Идентификационный номер налогоплательщика
	KPP        string `json:"kpp"`         // Код причины постановки на учет
	FullName   string `json:"full_name"`   // Полное наименование
	ShortName  string `json:"short_name"`  // Краткое наименование
	RegionCode string `json:"region_code"` // Код региона

	// Статус
	Status         string     `json:"status"`           // Статус (ДЕЙСТВУЮЩАЯ, ЛИКВИДИРОВАНА и т.д.)
	StatusDate     *time.Time `json:"status_date"`      // Дата изменения статуса
	LiquidationDate *time.Time `json:"liquidation_date"` // Дата ликвидации

	// Руководитель
	DirectorFullName string `json:"director_full_name"` // ФИО руководителя
	DirectorINN      string `json:"director_inn"`       // ИНН руководителя
	DirectorPosition string `json:"director_position"`  // Должность руководителя

	// Адрес
	AddressFull       string `json:"address_full"`        // Полный адрес
	AddressPostalCode string `json:"address_postal_code"` // Почтовый индекс
	AddressRegion     string `json:"address_region"`      // Регион
	AddressCity       string `json:"address_city"`        // Город
	AddressStreet     string `json:"address_street"`      // Улица
	AddressHouse      string `json:"address_house"`       // Дом

	// Уставный капитал
	AuthorizedCapital float64 `json:"authorized_capital"` // Уставный капитал (руб)
	CapitalCurrency   string  `json:"capital_currency"`   // Валюта капитала

	// Учредители (упрощенно, для детектирования изменений)
	Founders []Founder `json:"founders"`

	// Виды деятельности (ОКВЭД)
	MainOKVED       string   `json:"main_okved"`        // Основной ОКВЭД
	AdditionalOKVED []string `json:"additional_okved"`  // Дополнительные ОКВЭД

	// Лицензии (кол-во, для детектирования изменений)
	LicensesCount int `json:"licenses_count"`

	// Филиалы (кол-во, для детектирования изменений)
	BranchesCount int `json:"branches_count"`

	// Метаданные
	RegistrationDate time.Time `json:"registration_date"` // Дата регистрации
	ExtractDate      time.Time `json:"extract_date"`      // Дата выписки (версия данных)
	LastUpdate       time.Time `json:"last_update"`       // Дата последнего обновления записи
}

// Founder представляет учредителя компании
type Founder struct {
	FullName    string  `json:"full_name"`    // ФИО/Наименование учредителя
	INN         string  `json:"inn"`          // ИНН учредителя
	OGRN        string  `json:"ogrn"`         // ОГРН учредителя (если юр. лицо)
	ShareAmount float64 `json:"share_amount"` // Размер доли (руб)
	SharePercent float64 `json:"share_percent"` // Доля в процентах
}

// IsActive проверяет, является ли компания действующей
func (c *Company) IsActive() bool {
	return c.Status == "ДЕЙСТВУЮЩАЯ"
}

// IsLiquidated проверяет, ликвидирована ли компания
func (c *Company) IsLiquidated() bool {
	return c.Status == "ЛИКВИДИРОВАНА" || c.Status == "В ПРОЦЕССЕ ЛИКВИДАЦИИ"
}

// HasDirector проверяет, есть ли у компании руководитель
func (c *Company) HasDirector() bool {
	return c.DirectorFullName != ""
}

// GetFounderByINN возвращает учредителя по ИНН
func (c *Company) GetFounderByINN(inn string) *Founder {
	for i := range c.Founders {
		if c.Founders[i].INN == inn {
			return &c.Founders[i]
		}
	}
	return nil
}

// GetFounderByOGRN возвращает учредителя по ОГРН
func (c *Company) GetFounderByOGRN(ogrn string) *Founder {
	for i := range c.Founders {
		if c.Founders[i].OGRN == ogrn {
			return &c.Founders[i]
		}
	}
	return nil
}

// HasOKVED проверяет, есть ли у компании указанный ОКВЭД
func (c *Company) HasOKVED(okved string) bool {
	if c.MainOKVED == okved {
		return true
	}
	for _, kod := range c.AdditionalOKVED {
		if kod == okved {
			return true
		}
	}
	return false
}
