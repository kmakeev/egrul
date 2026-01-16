package repository

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
)

// CompanyRepository интерфейс для работы с компаниями
type CompanyRepository interface {
	GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error)
	GetByINN(ctx context.Context, inn string) (*model.Company, error)
	List(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) ([]*model.Company, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Company, error)
}

// EntrepreneurRepository интерфейс для работы с ИП
type EntrepreneurRepository interface {
	GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error)
	GetByINN(ctx context.Context, inn string) (*model.Entrepreneur, error)
	List(ctx context.Context, filter *model.EntrepreneurFilter, pagination *model.Pagination, sort *model.EntrepreneurSort) ([]*model.Entrepreneur, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Entrepreneur, error)
}

// FounderRepository интерфейс для работы с учредителями
type FounderRepository interface {
	GetByCompanyOGRN(ctx context.Context, ogrn string, limit, offset int) ([]*model.Founder, error)
	GetRelatedCompanies(ctx context.Context, inn string, limit, offset int) ([]string, error)
	GetCompaniesWithCommonFounders(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetFounderCompanies(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetCommonFoundersDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Founder, error)
	GetCompaniesWithCommonDirectors(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetCommonDirectorsDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Person, error)
	GetCompaniesWhereFounderIsDirector(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetCompaniesWhereDirectorIsFounder(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetCrossPersonDetails(ctx context.Context, ogrn1, ogrn2 string, crossType string) ([]*model.Person, error)
	GetCompaniesWithCommonAddress(ctx context.Context, ogrn string, limit, offset int) ([]string, error)
	GetCommonAddressDetails(ctx context.Context, ogrn1, ogrn2 string) (*model.Address, error)
}

// LicenseRepository интерфейс для работы с лицензиями
type LicenseRepository interface {
	GetByEntityOGRN(ctx context.Context, ogrn string) ([]*model.License, error)
}

// BranchRepository интерфейс для работы с филиалами
type BranchRepository interface {
	GetByCompanyOGRN(ctx context.Context, ogrn string) ([]*model.Branch, error)
}

// StatisticsRepository интерфейс для работы со статистикой
type StatisticsRepository interface {
	GetStatistics(ctx context.Context, filter *model.StatsFilter) (*model.Statistics, error)
	GetActivityStats(ctx context.Context, limit int) ([]*model.ActivityStatistics, error)
}

// HistoryRepository интерфейс для работы с историей изменений
type HistoryRepository interface {
	GetByEntityID(ctx context.Context, entityType, entityID string, limit, offset int) ([]*model.HistoryRecord, error)
	CountByEntityID(ctx context.Context, entityType, entityID string) (int, error)
	InsertOrUpdate(ctx context.Context, record *model.HistoryRecord, entityType, entityID string, extractDate string, sourceFile string, fileHash string) error
}
