// Package model содержит модели данных для GraphQL
package model

import (
	"time"
)

// EntityStatus статус юридического лица или ИП
type EntityStatus string

const (
	EntityStatusActive      EntityStatus = "ACTIVE"
	EntityStatusLiquidated  EntityStatus = "LIQUIDATED"
	EntityStatusLiquidating EntityStatus = "LIQUIDATING"
	EntityStatusReorganizing EntityStatus = "REORGANIZING"
	EntityStatusBankrupt    EntityStatus = "BANKRUPT"
	EntityStatusUnknown     EntityStatus = "UNKNOWN"
)

func (e EntityStatus) IsValid() bool {
	switch e {
	case EntityStatusActive, EntityStatusLiquidated, EntityStatusLiquidating,
		EntityStatusReorganizing, EntityStatusBankrupt, EntityStatusUnknown:
		return true
	}
	return false
}

func (e EntityStatus) String() string {
	return string(e)
}

// ParseEntityStatus парсит статус из строки БД
func ParseEntityStatus(s string) EntityStatus {
	switch s {
	case "active", "действующий", "действует":
		return EntityStatusActive
	case "liquidated", "ликвидирован", "прекращено":
		return EntityStatusLiquidated
	case "liquidating", "в процессе ликвидации":
		return EntityStatusLiquidating
	case "reorganizing", "в процессе реорганизации":
		return EntityStatusReorganizing
	case "bankrupt", "банкрот":
		return EntityStatusBankrupt
	default:
		return EntityStatusUnknown
	}
}

// FounderType тип учредителя
type FounderType string

const (
	FounderTypePerson        FounderType = "PERSON"
	FounderTypeRussianCompany FounderType = "RUSSIAN_COMPANY"
	FounderTypeForeignCompany FounderType = "FOREIGN_COMPANY"
	FounderTypePublicEntity  FounderType = "PUBLIC_ENTITY"
	FounderTypeFund          FounderType = "FUND"
)

func (f FounderType) IsValid() bool {
	switch f {
	case FounderTypePerson, FounderTypeRussianCompany, FounderTypeForeignCompany,
		FounderTypePublicEntity, FounderTypeFund:
		return true
	}
	return false
}

func (f FounderType) String() string {
	return string(f)
}

// ParseFounderType парсит тип учредителя из строки БД
func ParseFounderType(s string) FounderType {
	switch s {
	case "person", "physical_person", "физическое лицо":
		return FounderTypePerson
	case "russian_company", "юридическое лицо":
		return FounderTypeRussianCompany
	case "foreign_company", "иностранное юридическое лицо":
		return FounderTypeForeignCompany
	case "public_entity", "публичное образование":
		return FounderTypePublicEntity
	case "fund", "фонд":
		return FounderTypeFund
	default:
		return FounderTypePerson
	}
}

// BranchType тип филиала
type BranchType string

const (
	BranchTypeBranch        BranchType = "BRANCH"
	BranchTypeRepresentative BranchType = "REPRESENTATIVE"
)

func (b BranchType) IsValid() bool {
	switch b {
	case BranchTypeBranch, BranchTypeRepresentative:
		return true
	}
	return false
}

func (b BranchType) String() string {
	return string(b)
}

// EntityType тип сущности
type EntityType string

const (
	EntityTypeCompany     EntityType = "COMPANY"
	EntityTypeEntrepreneur EntityType = "ENTREPRENEUR"
)

// SortOrder порядок сортировки
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// CompanySortField поле для сортировки
type CompanySortField string

const (
	CompanySortFieldOgrn            CompanySortField = "OGRN"
	CompanySortFieldInn             CompanySortField = "INN"
	CompanySortFieldFullName        CompanySortField = "FULL_NAME"
	CompanySortFieldRegistrationDate CompanySortField = "REGISTRATION_DATE"
	CompanySortFieldCapitalAmount   CompanySortField = "CAPITAL_AMOUNT"
	CompanySortFieldUpdatedAt       CompanySortField = "UPDATED_AT"
)

// Address адрес
type Address struct {
	PostalCode  *string `json:"postalCode"`
	RegionCode  *string `json:"regionCode"`
	Region      *string `json:"region"`
	District    *string `json:"district"`
	City        *string `json:"city"`
	Locality    *string `json:"locality"`
	Street      *string `json:"street"`
	House       *string `json:"house"`
	Building    *string `json:"building"`
	Flat        *string `json:"flat"`
	FullAddress *string `json:"fullAddress"`
	FiasID      *string `json:"fiasId"`
	KladrCode   *string `json:"kladrCode"`
}

// LegalForm организационно-правовая форма
type LegalForm struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
}

// Money денежная сумма
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Person физическое лицо
type Person struct {
	LastName     string  `json:"lastName"`
	FirstName    string  `json:"firstName"`
	MiddleName   *string `json:"middleName"`
	Inn          *string `json:"inn"`
	Position     *string `json:"position"`
	PositionCode *string `json:"positionCode"`
}

// Activity вид деятельности
type Activity struct {
	Code   string  `json:"code"`
	Name   *string `json:"name"`
	IsMain bool    `json:"isMain"`
}

// Authority регистрирующий орган
type Authority struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
}

// Share доля в уставном капитале
type Share struct {
	Percent      *float64 `json:"percent"`
	NominalValue *float64 `json:"nominalValue"`
}

// OldRegistration сведения о регистрации до 01.07.2002
type OldRegistration struct {
	RegNumber *string `json:"regNumber"`
	RegDate   *Date   `json:"regDate"`
	Authority *string `json:"authority"`
}

// Founder учредитель
type Founder struct {
	Type             FounderType `json:"type"`
	Ogrn             *string     `json:"ogrn"`
	Inn              *string     `json:"inn"`
	Name             string      `json:"name"`
	LastName         *string     `json:"lastName"`
	FirstName        *string     `json:"firstName"`
	MiddleName       *string     `json:"middleName"`
	Country          *string     `json:"country"`
	Citizenship      *string     `json:"citizenship"`
	ShareNominalValue *float64   `json:"shareNominalValue"`
	SharePercent     *float64    `json:"sharePercent"`
}

// License лицензия
type License struct {
	ID        string  `json:"id"`
	Number    string  `json:"number"`
	Series    *string `json:"series"`
	Activity  *string `json:"activity"`
	StartDate *Date   `json:"startDate"`
	EndDate   *Date   `json:"endDate"`
	Authority *string `json:"authority"`
	Status    *string `json:"status"`
}

// Branch филиал
type Branch struct {
	ID      string     `json:"id"`
	Type    BranchType `json:"type"`
	Name    *string    `json:"name"`
	Kpp     *string    `json:"kpp"`
	Address *Address   `json:"address"`
}

// HistoryRecord запись истории
type HistoryRecord struct {
	ID                 string     `json:"id"`
	Grn                string     `json:"grn"`
	Date               Date       `json:"date"`
	ReasonCode         *string    `json:"reasonCode"`
	ReasonDescription  *string    `json:"reasonDescription"`
	Authority          *Authority `json:"authority"`
	CertificateSeries  *string    `json:"certificateSeries"`
	CertificateNumber  *string    `json:"certificateNumber"`
	CertificateDate    *Date      `json:"certificateDate"`
	SnapshotFullName   *string    `json:"snapshotFullName"`
	SnapshotStatus     *string    `json:"snapshotStatus"`
	SnapshotAddress    *string    `json:"snapshotAddress"`
}

// Company юридическое лицо
type Company struct {
	Ogrn              string        `json:"ogrn"`
	OgrnDate          *Date         `json:"ogrnDate"`
	Inn               string        `json:"inn"`
	Kpp               *string       `json:"kpp"`
	FullName          string        `json:"fullName"`
	ShortName         *string       `json:"shortName"`
	BrandName         *string       `json:"brandName"`
	LegalForm         *LegalForm    `json:"legalForm"`
	Status            EntityStatus  `json:"status"`
	StatusCode        *string       `json:"statusCode"`
	TerminationMethod *string       `json:"terminationMethod"`
	RegistrationDate  *Date         `json:"registrationDate"`
	TerminationDate   *Date         `json:"terminationDate"`
	ExtractDate       *Date         `json:"extractDate"`
	Address           *Address         `json:"address"`
	Email             *string          `json:"email"`
	Capital           *Money           `json:"capital"`
	CompanyShare      *Share           `json:"companyShare"`
	OldRegistration   *OldRegistration `json:"oldRegistration"`
	Director          *Person          `json:"director"`
	MainActivity      *Activity     `json:"mainActivity"`
	Activities        []*Activity   `json:"activities"`
	RegAuthority      *Authority    `json:"regAuthority"`
	TaxAuthority      *Authority    `json:"taxAuthority"`
	PfrRegNumber      *string       `json:"pfrRegNumber"`
	FssRegNumber      *string       `json:"fssRegNumber"`
	FoundersCount     int           `json:"foundersCount"`
	LicensesCount     int           `json:"licensesCount"`
	BranchesCount     int           `json:"branchesCount"`
	IsBankrupt        bool          `json:"isBankrupt"`
	BankruptcyStage   *string       `json:"bankruptcyStage"`
	IsLiquidating     bool          `json:"isLiquidating"`
	IsReorganizing    bool          `json:"isReorganizing"`
	LastGrn           *string       `json:"lastGrn"`
	LastGrnDate       *Date         `json:"lastGrnDate"`
	SourceFile        *string       `json:"sourceFile"`
	VersionDate       Date          `json:"versionDate"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

// Entrepreneur индивидуальный предприниматель
type Entrepreneur struct {
	Ogrnip                 string       `json:"ogrnip"`
	OgrnipDate             *Date        `json:"ogrnipDate"`
	Inn                    string       `json:"inn"`
	LastName               string       `json:"lastName"`
	FirstName              string       `json:"firstName"`
	MiddleName             *string      `json:"middleName"`
	Gender                 *string      `json:"gender"`
	CitizenshipType        *string      `json:"citizenshipType"`
	CitizenshipCountryCode *string      `json:"citizenshipCountryCode"`
	CitizenshipCountryName *string      `json:"citizenshipCountryName"`
	Status                 EntityStatus `json:"status"`
	StatusCode             *string      `json:"statusCode"`
	TerminationMethod      *string      `json:"terminationMethod"`
	RegistrationDate       *Date        `json:"registrationDate"`
	TerminationDate        *Date        `json:"terminationDate"`
	ExtractDate            *Date        `json:"extractDate"`
	Address                *Address     `json:"address"`
	Email                  *string      `json:"email"`
	MainActivity           *Activity    `json:"mainActivity"`
	Activities             []*Activity  `json:"activities"`
	RegAuthority           *Authority   `json:"regAuthority"`
	TaxAuthority           *Authority   `json:"taxAuthority"`
	PfrRegNumber           *string      `json:"pfrRegNumber"`
	FssRegNumber           *string      `json:"fssRegNumber"`
	LicensesCount          int          `json:"licensesCount"`
	IsBankrupt             bool         `json:"isBankrupt"`
	BankruptcyDate         *Date        `json:"bankruptcyDate"`
	BankruptcyCaseNumber   *string      `json:"bankruptcyCaseNumber"`
	LastGrn                *string      `json:"lastGrn"`
	LastGrnDate            *Date        `json:"lastGrnDate"`
	SourceFile             *string      `json:"sourceFile"`
	VersionDate            Date         `json:"versionDate"`
	CreatedAt              time.Time    `json:"createdAt"`
	UpdatedAt              time.Time    `json:"updatedAt"`
}

// Pagination пагинация
type Pagination struct {
	Limit  *int    `json:"limit"`
	Offset *int    `json:"offset"`
	First  *int    `json:"first"`
	After  *string `json:"after"`
	Last   *int    `json:"last"`
	Before *string `json:"before"`
}

// GetLimit возвращает лимит с дефолтным значением
func (p *Pagination) GetLimit() int {
	if p == nil || p.Limit == nil {
		return 20
	}
	if *p.Limit > 100 {
		return 100
	}
	return *p.Limit
}

// GetOffset возвращает оффсет
func (p *Pagination) GetOffset() int {
	if p == nil || p.Offset == nil {
		return 0
	}
	return *p.Offset
}

// CompanySort сортировка компаний
type CompanySort struct {
	Field CompanySortField `json:"field"`
	Order *SortOrder       `json:"order,omitempty"`
}

// CompanyFilter фильтр компаний
type CompanyFilter struct {
	Inn              *string         `json:"inn"`
	Ogrn             *string         `json:"ogrn"`
	Name             *string         `json:"name"`
	RegionCode       *string         `json:"regionCode"`
	Region           *string         `json:"region"`
	Okved            *string         `json:"okved"`
	Status           *EntityStatus   `json:"status"`
	StatusIn         []EntityStatus  `json:"statusIn"`
	StatusCode       *string         `json:"statusCode"`
	StatusCodeIn     []string        `json:"statusCodeIn"`
	RegisteredAfter  *Date           `json:"registeredAfter"`
	RegisteredBefore *Date           `json:"registeredBefore"`
	TerminatedAfter  *Date           `json:"terminatedAfter"`
	TerminatedBefore *Date           `json:"terminatedBefore"`
	CapitalMin       *float64        `json:"capitalMin"`
	CapitalMax       *float64        `json:"capitalMax"`
	IsBankrupt       *bool           `json:"isBankrupt"`
	IsLiquidating    *bool           `json:"isLiquidating"`
	HasDirector      *bool           `json:"hasDirector"`
	FounderName      *string         `json:"founderName"`
}

// EntrepreneurFilter фильтр ИП
type EntrepreneurFilter struct {
	Inn              *string        `json:"inn"`
	Ogrnip           *string        `json:"ogrnip"`
	Name             *string        `json:"name"`
	LastName         *string        `json:"lastName"`
	FirstName        *string        `json:"firstName"`
	RegionCode       *string        `json:"regionCode"`
	Region           *string        `json:"region"`
	Okved            *string        `json:"okved"`
	Status           *EntityStatus  `json:"status"`
	StatusIn         []EntityStatus `json:"statusIn"`
	StatusCode       *string        `json:"statusCode"`
	StatusCodeIn     []string       `json:"statusCodeIn"`
	RegisteredAfter  *Date          `json:"registeredAfter"`
	RegisteredBefore *Date          `json:"registeredBefore"`
	TerminatedAfter  *Date          `json:"terminatedAfter"`
	TerminatedBefore *Date          `json:"terminatedBefore"`
	IsBankrupt       *bool          `json:"isBankrupt"`
}

// StatsFilter фильтр статистики
type StatsFilter struct {
	RegionCode *string `json:"regionCode"`
	Okved      *string `json:"okved"`
	DateFrom   *Date   `json:"dateFrom"`
	DateTo     *Date   `json:"dateTo"`
}

// PageInfo информация о странице
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
	TotalCount      int     `json:"totalCount"`
}

// CompanyEdge ребро компании
type CompanyEdge struct {
	Node   *Company `json:"node"`
	Cursor string   `json:"cursor"`
}

// CompanyConnection соединение компаний
type CompanyConnection struct {
	Edges      []*CompanyEdge `json:"edges"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

// EntrepreneurEdge ребро ИП
type EntrepreneurEdge struct {
	Node   *Entrepreneur `json:"node"`
	Cursor string        `json:"cursor"`
}

// EntrepreneurConnection соединение ИП
type EntrepreneurConnection struct {
	Edges      []*EntrepreneurEdge `json:"edges"`
	PageInfo   *PageInfo           `json:"pageInfo"`
	TotalCount int                 `json:"totalCount"`
}

// SearchResult результат поиска
type SearchResult struct {
	Companies          []*Company      `json:"companies"`
	Entrepreneurs      []*Entrepreneur `json:"entrepreneurs"`
	TotalCompanies     int             `json:"totalCompanies"`
	TotalEntrepreneurs int             `json:"totalEntrepreneurs"`
}

// RegionStatistics статистика по регионам
type RegionStatistics struct {
	RegionCode        string `json:"regionCode"`
	RegionName        string `json:"regionName"`
	CompaniesCount    int    `json:"companiesCount"`
	EntrepreneursCount int   `json:"entrepreneursCount"`
	ActiveCount       int    `json:"activeCount"`
	LiquidatedCount   int    `json:"liquidatedCount"`
}

// ActivityStatistics статистика по видам деятельности
type ActivityStatistics struct {
	OkvedCode          string `json:"okvedCode"`
	OkvedName          string `json:"okvedName"`
	CompaniesCount     int    `json:"companiesCount"`
	EntrepreneursCount int    `json:"entrepreneursCount"`
}

// Statistics общая статистика
type Statistics struct {
	TotalCompanies      int                   `json:"totalCompanies"`
	TotalEntrepreneurs  int                   `json:"totalEntrepreneurs"`
	ActiveCompanies     int                   `json:"activeCompanies"`
	ActiveEntrepreneurs int                   `json:"activeEntrepreneurs"`
	LiquidatedCompanies int                   `json:"liquidatedCompanies"`
	LiquidatedEntrepreneurs int              `json:"liquidatedEntrepreneurs"`
	RegisteredToday     int                   `json:"registeredToday"`
	RegisteredThisMonth int                   `json:"registeredThisMonth"`
	RegisteredThisYear  int                   `json:"registeredThisYear"`
	ByRegion            []*RegionStatistics   `json:"byRegion"`
	ByActivity          []*ActivityStatistics `json:"byActivity"`
}

// RelationshipType тип связи между компаниями
type RelationshipType string

const (
	RelationshipTypeFounderCompany    RelationshipType = "FOUNDER_COMPANY"    // Компания-учредитель
	RelationshipTypeSubsidiaryCompany RelationshipType = "SUBSIDIARY_COMPANY" // Дочерняя компания
	RelationshipTypeCommonFounders    RelationshipType = "COMMON_FOUNDERS"    // Общие учредители-физлица
	RelationshipTypeCommonDirectors   RelationshipType = "COMMON_DIRECTORS"   // Общие руководители-физлица
	RelationshipTypeCommonAddress     RelationshipType = "COMMON_ADDRESS"     // Общий адрес регистрации
	RelationshipTypeFounderToDirector RelationshipType = "FOUNDER_TO_DIRECTOR" // Учредитель → Руководитель
	RelationshipTypeDirectorToFounder RelationshipType = "DIRECTOR_TO_FOUNDER" // Руководитель → Учредитель
	RelationshipTypeRelatedByPerson   RelationshipType = "RELATED_BY_PERSON"  // Связанная через физлицо
)

func (r RelationshipType) IsValid() bool {
	switch r {
	case RelationshipTypeFounderCompany, RelationshipTypeSubsidiaryCompany,
		RelationshipTypeCommonFounders, RelationshipTypeCommonDirectors, 
		RelationshipTypeFounderToDirector, RelationshipTypeDirectorToFounder, RelationshipTypeRelatedByPerson:
		return true
	}
	return false
}

func (r RelationshipType) String() string {
	return string(r)
}

// RelatedCompany связанная компания
type RelatedCompany struct {
	Company          *Company          `json:"company"`
	RelationshipType RelationshipType  `json:"relationshipType"`
	Description      *string           `json:"description"`
	CommonFounders   []*Founder        `json:"commonFounders"`
	CommonDirectors  []*Person         `json:"commonDirectors"`
	CommonAddress    *Address          `json:"commonAddress"`
}