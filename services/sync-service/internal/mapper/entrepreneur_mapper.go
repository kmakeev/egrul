package mapper

import (
	"encoding/json"
	"fmt"
	"time"
)

type EntrepreneurRow struct {
	OGRNIP               string    `ch:"ogrnip"`
	INN                  string    `ch:"inn"`
	LastName             string    `ch:"last_name"`
	FirstName            string    `ch:"first_name"`
	MiddleName           string    `ch:"middle_name"`
	Gender               string    `ch:"gender"`
	CitizenshipType      string    `ch:"citizenship_type"`
	Status               string    `ch:"status"`
	RegionCode           string    `ch:"region_code"`
	Region               string    `ch:"region"`
	City                 string    `ch:"city"`
	FullAddress          string    `ch:"full_address"`
	Email                string    `ch:"email"`
	OKVEDMainCode        string    `ch:"okved_main_code"`
	OKVEDMainName        string    `ch:"okved_main_name"`
	OKVEDAdditional      []string  `ch:"okved_additional"`
	OKVEDAdditionalNames []string  `ch:"okved_additional_names"`
	RegistrationDate     time.Time `ch:"registration_date"`
	TerminationDate      time.Time `ch:"termination_date"`
	IsBankrupt           bool      `ch:"is_bankrupt"`
	UpdatedAt            time.Time `ch:"updated_at"`
}

type EntrepreneurDocument struct {
	OGRNIP               string     `json:"ogrnip"`
	INN                  string     `json:"inn"`
	LastName             string     `json:"last_name"`
	FirstName            string     `json:"first_name"`
	MiddleName           string     `json:"middle_name,omitempty"`
	FullName             string     `json:"full_name"`
	Gender               string     `json:"gender,omitempty"`
	CitizenshipType      string     `json:"citizenship_type,omitempty"`
	Status               string     `json:"status"`
	RegionCode           string     `json:"region_code"`
	Region               string     `json:"region,omitempty"`
	City                 string     `json:"city,omitempty"`
	FullAddress          string     `json:"full_address,omitempty"`
	Email                string     `json:"email,omitempty"`
	OKVEDMainCode        string     `json:"okved_main_code,omitempty"`
	OKVEDMainName        string     `json:"okved_main_name,omitempty"`
	OKVEDAdditional      []string   `json:"okved_additional,omitempty"`
	OKVEDAdditionalNames []string   `json:"okved_additional_names,omitempty"`
	RegistrationDate     *time.Time `json:"registration_date,omitempty"`
	TerminationDate      *time.Time `json:"termination_date,omitempty"`
	IsBankrupt           bool       `json:"is_bankrupt"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func MapEntrepreneurToES(row EntrepreneurRow) ([]byte, error) {
	// Construct full name
	fullName := fmt.Sprintf("%s %s", row.LastName, row.FirstName)
	if row.MiddleName != "" {
		fullName += " " + row.MiddleName
	}

	doc := EntrepreneurDocument{
		OGRNIP:               row.OGRNIP,
		INN:                  row.INN,
		LastName:             row.LastName,
		FirstName:            row.FirstName,
		MiddleName:           row.MiddleName,
		FullName:             fullName,
		Gender:               row.Gender,
		CitizenshipType:      row.CitizenshipType,
		Status:               row.Status,
		RegionCode:           row.RegionCode,
		Region:               row.Region,
		City:                 row.City,
		FullAddress:          row.FullAddress,
		Email:                row.Email,
		OKVEDMainCode:        row.OKVEDMainCode,
		OKVEDMainName:        row.OKVEDMainName,
		OKVEDAdditional:      row.OKVEDAdditional,
		OKVEDAdditionalNames: row.OKVEDAdditionalNames,
		IsBankrupt:           row.IsBankrupt,
		UpdatedAt:            row.UpdatedAt,
	}

	// Handle nullable dates
	if !row.RegistrationDate.IsZero() {
		doc.RegistrationDate = &row.RegistrationDate
	}
	if !row.TerminationDate.IsZero() {
		doc.TerminationDate = &row.TerminationDate
	}

	return json.Marshal(doc)
}
