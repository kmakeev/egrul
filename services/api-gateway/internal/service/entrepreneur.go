package service

import (
	"context"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/egrul-system/services/api-gateway/internal/repository/clickhouse"
	"go.uber.org/zap"
)

// EntrepreneurService сервис для работы с ИП
type EntrepreneurService struct {
	entrepreneurRepo *clickhouse.EntrepreneurRepository
	licenseRepo      *clickhouse.LicenseRepository
	historyRepo      *clickhouse.HistoryRepository
	logger           *zap.Logger
}

// NewEntrepreneurService создает новый сервис ИП
func NewEntrepreneurService(
	entrepreneurRepo *clickhouse.EntrepreneurRepository,
	licenseRepo *clickhouse.LicenseRepository,
	historyRepo *clickhouse.HistoryRepository,
	logger *zap.Logger,
) *EntrepreneurService {
	return &EntrepreneurService{
		entrepreneurRepo: entrepreneurRepo,
		licenseRepo:      licenseRepo,
		historyRepo:      historyRepo,
		logger:           logger.Named("entrepreneur_service"),
	}
}

// GetByOGRNIP получает ИП по ОГРНИП
func (s *EntrepreneurService) GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error) {
	return s.entrepreneurRepo.GetByOGRNIP(ctx, ogrnip)
}

// GetByINN получает ИП по ИНН
func (s *EntrepreneurService) GetByINN(ctx context.Context, inn string) (*model.Entrepreneur, error) {
	return s.entrepreneurRepo.GetByINN(ctx, inn)
}

// List возвращает список ИП
func (s *EntrepreneurService) List(ctx context.Context, filter *model.EntrepreneurFilter, pagination *model.Pagination, sort *model.EntrepreneurSort) (*model.EntrepreneurConnection, error) {
	entrepreneurs, totalCount, err := s.entrepreneurRepo.List(ctx, filter, pagination, sort)
	if err != nil {
		return nil, err
	}

	edges := make([]*model.EntrepreneurEdge, len(entrepreneurs))
	for i, entrepreneur := range entrepreneurs {
		edges[i] = &model.EntrepreneurEdge{
			Node:   entrepreneur,
			Cursor: clickhouse.EncodeCursor(entrepreneur.Ogrnip),
		}
	}

	offset := pagination.GetOffset()

	hasNextPage := offset+len(entrepreneurs) < totalCount
	hasPreviousPage := offset > 0

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.EntrepreneurConnection{
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

// Search выполняет поиск ИП
func (s *EntrepreneurService) Search(ctx context.Context, query string, limit, offset int) ([]*model.Entrepreneur, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.entrepreneurRepo.Search(ctx, query, limit, offset)
}

// GetLicenses получает лицензии ИП
func (s *EntrepreneurService) GetLicenses(ctx context.Context, ogrnip string) ([]*model.License, error) {
	return s.licenseRepo.GetByEntityOGRN(ctx, ogrnip)
}

// GetHistory получает историю изменений ИП
func (s *EntrepreneurService) GetHistory(ctx context.Context, ogrnip string, limit, offset int) ([]*model.HistoryRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.historyRepo.GetByEntityID(ctx, "entrepreneur", ogrnip, limit, offset)
}

