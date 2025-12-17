package service

import (
	"context"
	"sync"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// SearchService сервис для универсального поиска
type SearchService struct {
	companyService     *CompanyService
	entrepreneurService *EntrepreneurService
	logger             *zap.Logger
}

// NewSearchService создает новый сервис поиска
func NewSearchService(
	companyService *CompanyService,
	entrepreneurService *EntrepreneurService,
	logger *zap.Logger,
) *SearchService {
	return &SearchService{
		companyService:     companyService,
		entrepreneurService: entrepreneurService,
		logger:             logger.Named("search_service"),
	}
}

// Search выполняет универсальный поиск
func (s *SearchService) Search(ctx context.Context, query string, limit int) (*model.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	result := &model.SearchResult{
		Companies:     make([]*model.Company, 0),
		Entrepreneurs: make([]*model.Entrepreneur, 0),
	}

	var wg sync.WaitGroup
	var companiesErr, entrepreneursErr error

	// Параллельный поиск компаний и ИП
	wg.Add(2)

	go func() {
		defer wg.Done()
		companies, err := s.companyService.Search(ctx, query, limit, 0)
		if err != nil {
			companiesErr = err
			return
		}
		result.Companies = companies
		result.TotalCompanies = len(companies)
	}()

	go func() {
		defer wg.Done()
		entrepreneurs, err := s.entrepreneurService.Search(ctx, query, limit, 0)
		if err != nil {
			entrepreneursErr = err
			return
		}
		result.Entrepreneurs = entrepreneurs
		result.TotalEntrepreneurs = len(entrepreneurs)
	}()

	wg.Wait()

	// Логируем ошибки, но не прерываем выполнение
	if companiesErr != nil {
		s.logger.Warn("error searching companies", zap.Error(companiesErr))
	}
	if entrepreneursErr != nil {
		s.logger.Warn("error searching entrepreneurs", zap.Error(entrepreneursErr))
	}

	return result, nil
}

