// Package service содержит бизнес-логику приложения
package service

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/egrul-system/services/api-gateway/internal/repository/clickhouse"
	"go.uber.org/zap"
)

// CompanyService сервис для работы с компаниями
type CompanyService struct {
	companyRepo  *clickhouse.CompanyRepository
	founderRepo  *clickhouse.FounderRepository
	licenseRepo  *clickhouse.LicenseRepository
	branchRepo   *clickhouse.BranchRepository
	historyRepo  *clickhouse.HistoryRepository
	logger       *zap.Logger
}

// NewCompanyService создает новый сервис компаний
func NewCompanyService(
	companyRepo *clickhouse.CompanyRepository,
	founderRepo *clickhouse.FounderRepository,
	licenseRepo *clickhouse.LicenseRepository,
	branchRepo *clickhouse.BranchRepository,
	historyRepo *clickhouse.HistoryRepository,
	logger *zap.Logger,
) *CompanyService {
	return &CompanyService{
		companyRepo:  companyRepo,
		founderRepo:  founderRepo,
		licenseRepo:  licenseRepo,
		branchRepo:   branchRepo,
		historyRepo:  historyRepo,
		logger:       logger.Named("company_service"),
	}
}

// GetByOGRN получает компанию по ОГРН
func (s *CompanyService) GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error) {
	return s.companyRepo.GetByOGRN(ctx, ogrn)
}

// GetByINN получает компанию по ИНН
func (s *CompanyService) GetByINN(ctx context.Context, inn string) (*model.Company, error) {
	return s.companyRepo.GetByINN(ctx, inn)
}

// List возвращает список компаний
func (s *CompanyService) List(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) (*model.CompanyConnection, error) {
	companies, totalCount, err := s.companyRepo.List(ctx, filter, pagination, sort)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.CompanyEdge, len(companies))
	for i, company := range companies {
		edges[i] = &model.CompanyEdge{
			Node:   company,
			Cursor: clickhouse.EncodeCursor(company.Ogrn),
		}
	}

	offset := pagination.GetOffset()

	hasNextPage := offset+len(companies) < totalCount
	hasPreviousPage := offset > 0

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.CompanyConnection{
		Edges:      edges,
		TotalCount: totalCount,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: hasPreviousPage,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			TotalCount:      totalCount,
		},
	}, nil
}

// Search выполняет поиск компаний
func (s *CompanyService) Search(ctx context.Context, query string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.companyRepo.Search(ctx, query, limit, offset)
}

// GetFounders получает учредителей компании
func (s *CompanyService) GetFounders(ctx context.Context, ogrn string, limit, offset int) ([]*model.Founder, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.founderRepo.GetByCompanyOGRN(ctx, ogrn, limit, offset)
}

// GetLicenses получает лицензии компании
func (s *CompanyService) GetLicenses(ctx context.Context, ogrn string) ([]*model.License, error) {
	return s.licenseRepo.GetByEntityOGRN(ctx, ogrn)
}

// GetBranches получает филиалы компании
func (s *CompanyService) GetBranches(ctx context.Context, ogrn string) ([]*model.Branch, error) {
	return s.branchRepo.GetByCompanyOGRN(ctx, ogrn)
}

// GetHistory получает историю изменений компании
func (s *CompanyService) GetHistory(ctx context.Context, ogrn string, limit, offset int) ([]*model.HistoryRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.historyRepo.GetByEntityID(ctx, "company", ogrn, limit, offset)
}

// GetRelatedCompanies получает связанные компании по ИНН учредителя
func (s *CompanyService) GetRelatedCompanies(ctx context.Context, inn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetRelatedCompanies(ctx, inn, limit, offset)
	if err != nil {
		return nil, err
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, ogrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, ogrn)
		if err != nil {
			s.logger.Warn("failed to get company by ogrn", zap.String("ogrn", ogrn), zap.Error(err))
			continue
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

