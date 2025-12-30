// Package service содержит бизнес-логику приложения
package service

import (
	"context"
	"fmt"

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

// GetHistoryCount получает общее количество записей истории компании
func (s *CompanyService) GetHistoryCount(ctx context.Context, ogrn string) (int, error) {
	return s.historyRepo.CountByEntityID(ctx, "company", ogrn)
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

// GetCompaniesWithCommonFounders получает компании с общими учредителями-физлицами
func (s *CompanyService) GetCompaniesWithCommonFounders(ctx context.Context, ogrn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetCompaniesWithCommonFounders(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, relatedOgrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, relatedOgrn)
		if err != nil {
			s.logger.Warn("failed to get company by ogrn", zap.String("ogrn", relatedOgrn), zap.Error(err))
			continue
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// GetFounderCompanies получает компании-учредители
func (s *CompanyService) GetFounderCompanies(ctx context.Context, ogrn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetFounderCompanies(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}

	s.logger.Info("GetFounderCompanies: processing founder OGRNs", 
		zap.String("main_ogrn", ogrn), 
		zap.Int("founder_ogrns_count", len(ogrns)),
		zap.Strings("founder_ogrns", ogrns))

	companies := make([]*model.Company, 0, len(ogrns))
	for _, founderOgrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, founderOgrn)
		if err != nil {
			s.logger.Warn("failed to get founder company by ogrn", zap.String("ogrn", founderOgrn), zap.Error(err))
			// Создаем заглушку для компании-учредителя которая не найдена в базе
			company = &model.Company{
				Ogrn:     founderOgrn,
				Inn:      "неизвестно",
				FullName: fmt.Sprintf("Компания-учредитель (ОГРН: %s)", founderOgrn),
				Status:   model.EntityStatusUnknown,
			}
		}
		if company != nil {
			companies = append(companies, company)
			if err == nil {
				s.logger.Info("found founder company", zap.String("founder_ogrn", founderOgrn), zap.String("founder_name", company.FullName))
			} else {
				s.logger.Info("created placeholder for founder company", zap.String("founder_ogrn", founderOgrn))
			}
		}
	}

	s.logger.Info("GetFounderCompanies completed", 
		zap.String("main_ogrn", ogrn), 
		zap.Int("requested_ogrns", len(ogrns)),
		zap.Int("found_companies", len(companies)))

	return companies, nil
}

// GetCommonFoundersDetails получает детальную информацию об общих учредителях между двумя компаниями
func (s *CompanyService) GetCommonFoundersDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Founder, error) {
	return s.founderRepo.GetCommonFoundersDetails(ctx, ogrn1, ogrn2)
}

// GetCompaniesWithCommonDirectors получает компании с общими руководителями-физлицами
func (s *CompanyService) GetCompaniesWithCommonDirectors(ctx context.Context, ogrn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetCompaniesWithCommonDirectors(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, relatedOgrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, relatedOgrn)
		if err != nil {
			s.logger.Warn("failed to get company by ogrn", zap.String("ogrn", relatedOgrn), zap.Error(err))
			continue
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// GetCommonDirectorsDetails получает детальную информацию об общих руководителях между двумя компаниями
func (s *CompanyService) GetCommonDirectorsDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Person, error) {
	return s.founderRepo.GetCommonDirectorsDetails(ctx, ogrn1, ogrn2)
}

// GetCompaniesWhereFounderIsDirector получает компании, где учредители основной компании являются руководителями
func (s *CompanyService) GetCompaniesWhereFounderIsDirector(ctx context.Context, ogrn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetCompaniesWhereFounderIsDirector(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, relatedOgrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, relatedOgrn)
		if err != nil {
			s.logger.Warn("failed to get company by ogrn", zap.String("ogrn", relatedOgrn), zap.Error(err))
			continue
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// GetCompaniesWhereDirectorIsFounder получает компании, где руководитель основной компании является учредителем
func (s *CompanyService) GetCompaniesWhereDirectorIsFounder(ctx context.Context, ogrn string, limit, offset int) ([]*model.Company, error) {
	if limit <= 0 {
		limit = 50
	}

	ogrns, err := s.founderRepo.GetCompaniesWhereDirectorIsFounder(ctx, ogrn, limit, offset)
	if err != nil {
		return nil, err
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, relatedOgrn := range ogrns {
		company, err := s.companyRepo.GetByOGRN(ctx, relatedOgrn)
		if err != nil {
			s.logger.Warn("failed to get company by ogrn", zap.String("ogrn", relatedOgrn), zap.Error(err))
			continue
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// GetCrossPersonDetails получает детальную информацию о перекрестных связях через физлицо
func (s *CompanyService) GetCrossPersonDetails(ctx context.Context, ogrn1, ogrn2 string, crossType string) ([]*model.Person, error) {
	return s.founderRepo.GetCrossPersonDetails(ctx, ogrn1, ogrn2, crossType)
}

