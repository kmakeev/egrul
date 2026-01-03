// Package models содержит общие модели данных
package models

import "time"

// LegalEntity - юридическое лицо
type LegalEntity struct {
	ID               string     `json:"id"`
	OGRN             string     `json:"ogrn"`
	INN              string     `json:"inn"`
	KPP              string     `json:"kpp,omitempty"`
	FullName         string     `json:"full_name"`
	ShortName        string     `json:"short_name,omitempty"`
	RegistrationDate *time.Time `json:"registration_date,omitempty"`
	Address          *Address   `json:"address,omitempty"`
	Status           string     `json:"status"`
	MainActivity     *Activity  `json:"main_activity,omitempty"`
	Capital          *Capital   `json:"capital,omitempty"`
	Head             *Person    `json:"head,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// IndividualEntrepreneur - индивидуальный предприниматель
type IndividualEntrepreneur struct {
	ID               string     `json:"id"`
	OGRNIP           string     `json:"ogrnip"`
	INN              string     `json:"inn"`
	LastName         string     `json:"last_name"`
	FirstName        string     `json:"first_name"`
	MiddleName       string     `json:"middle_name,omitempty"`
	RegistrationDate *time.Time `json:"registration_date,omitempty"`
	Status           string     `json:"status"`
	MainActivity     *Activity  `json:"main_activity,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Address - адрес
type Address struct {
	PostalCode  string `json:"postal_code,omitempty"`
	Region      string `json:"region,omitempty"`
	City        string `json:"city,omitempty"`
	Street      string `json:"street,omitempty"`
	House       string `json:"house,omitempty"`
	FullAddress string `json:"full_address,omitempty"`
}

// Activity - вид деятельности (ОКВЭД)
type Activity struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Capital - уставный капитал
type Capital struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Person - физическое лицо
type Person struct {
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	INN        string `json:"inn,omitempty"`
	Position   string `json:"position,omitempty"`
}

// EntityStatus - статусы организаций
const (
	StatusActive      = "active"
	StatusLiquidated  = "liquidated"
	StatusReorganized = "reorganized"
	StatusBankrupt    = "bankrupt"
	StatusUnknown     = "unknown"
)

