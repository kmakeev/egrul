package mapper

import (
	"encoding/json"
	"time"
)

type CompanyRow struct {
	OGRN                   string    `ch:"ogrn"`
	INN                    string    `ch:"inn"`
	KPP                    string    `ch:"kpp"`
	FullName               string    `ch:"full_name"`
	ShortName              string    `ch:"short_name"`
	BrandName              string    `ch:"brand_name"`
	Status                 string    `ch:"status"`
	RegionCode             string    `ch:"region_code"`
	Region                 string    `ch:"region"`
	City                   string    `ch:"city"`
	FullAddress            string    `ch:"full_address"`
	Email                  string    `ch:"email"`
	OKVEDMainCode          string    `ch:"okved_main_code"`
	OKVEDMainName          string    `ch:"okved_main_name"`
	OKVEDAdditional        []string  `ch:"okved_additional"`
	OKVEDAdditionalNames   []string  `ch:"okved_additional_names"`
	HeadLastName           string    `ch:"head_last_name"`
	HeadFirstName          string    `ch:"head_first_name"`
	HeadMiddleName         string    `ch:"head_middle_name"`
	HeadINN                string    `ch:"head_inn"`
	OPFCode                string    `ch:"opf_code"`
	OPFName           string    `ch:"opf_name"`
	RegistrationDate       time.Time `ch:"registration_date"`
	TerminationDate        time.Time `ch:"termination_date"`
	CapitalAmount          *string   `ch:"capital_amount"`
	UpdatedAt              time.Time `ch:"updated_at"`
}

type CompanyDocument struct {
	OGRN                 string    `json:"ogrn"`
	INN                  string    `json:"inn"`
	KPP                  string    `json:"kpp,omitempty"`
	FullName             string    `json:"full_name"`
	ShortName            string    `json:"short_name,omitempty"`
	BrandName            string    `json:"brand_name,omitempty"`
	Status               string    `json:"status"`
	RegionCode           string    `json:"region_code"`
	Region               string    `json:"region,omitempty"`
	City                 string    `json:"city,omitempty"`
	FullAddress          string    `json:"full_address,omitempty"`
	Email                string    `json:"email,omitempty"`
	OKVEDMainCode        string    `json:"okved_main_code,omitempty"`
	OKVEDMainName        string    `json:"okved_main_name,omitempty"`
	OKVEDAdditional      []string  `json:"okved_additional,omitempty"`
	OKVEDAdditionalNames []string  `json:"okved_additional_names,omitempty"`
	HeadLastName         string    `json:"head_last_name,omitempty"`
	HeadFirstName        string    `json:"head_first_name,omitempty"`
	HeadMiddleName       string    `json:"head_middle_name,omitempty"`
	HeadINN              string    `json:"head_inn,omitempty"`
	OPFCode              string    `json:"opf_code,omitempty"`
	OPFName         string    `json:"opf_name,omitempty"`
	RegistrationDate     *time.Time `json:"registration_date,omitempty"`
	TerminationDate      *time.Time `json:"termination_date,omitempty"`
	CapitalAmount        string   `json:"capital_amount,omitempty"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func MapCompanyToES(row CompanyRow) ([]byte, error) {
	doc := CompanyDocument{
		OGRN:                 row.OGRN,
		INN:                  row.INN,
		KPP:                  row.KPP,
		FullName:             row.FullName,
		ShortName:            row.ShortName,
		BrandName:            row.BrandName,
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
		HeadLastName:         row.HeadLastName,
		HeadFirstName:        row.HeadFirstName,
		HeadMiddleName:       row.HeadMiddleName,
		HeadINN:              row.HeadINN,
		OPFCode:              row.OPFCode,
		OPFName:         row.OPFName,
		UpdatedAt:            row.UpdatedAt,
	}

	// Handle nullable dates
	if !row.RegistrationDate.IsZero() {
		doc.RegistrationDate = &row.RegistrationDate
	}
	if !row.TerminationDate.IsZero() {
		doc.TerminationDate = &row.TerminationDate
	}

	// Handle capital amount
	if row.CapitalAmount != nil {
		doc.CapitalAmount = *row.CapitalAmount
	}

	return json.Marshal(doc)
}
