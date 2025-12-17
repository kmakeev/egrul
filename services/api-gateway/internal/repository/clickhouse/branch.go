package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// BranchRepository репозиторий для работы с филиалами
type BranchRepository struct {
	client *Client
	logger *zap.Logger
}

// NewBranchRepository создает новый репозиторий филиалов
func NewBranchRepository(client *Client, logger *zap.Logger) *BranchRepository {
	return &BranchRepository{
		client: client,
		logger: logger.Named("branch_repo"),
	}
}

// branchRow структура для сканирования результатов запроса
type branchRow struct {
	ID          string         `ch:"id"`
	CompanyOgrn string         `ch:"company_ogrn"`
	CompanyInn  sql.NullString `ch:"company_inn"`
	CompanyName sql.NullString `ch:"company_name"`
	BranchType  string         `ch:"branch_type"`
	BranchName  sql.NullString `ch:"branch_name"`
	BranchKpp   string         `ch:"branch_kpp"`
	PostalCode  sql.NullString `ch:"postal_code"`
	RegionCode  sql.NullString `ch:"region_code"`
	Region      sql.NullString `ch:"region"`
	City        sql.NullString `ch:"city"`
	FullAddress sql.NullString `ch:"full_address"`
}

func (r *branchRow) toModel() *model.Branch {
	branch := &model.Branch{
		ID:   r.ID,
		Type: model.BranchTypeBranch,
	}

	if r.BranchType == "representative" {
		branch.Type = model.BranchTypeRepresentative
	}

	if r.BranchName.Valid {
		branch.Name = &r.BranchName.String
	}
	if r.BranchKpp != "" {
		branch.Kpp = &r.BranchKpp
	}

	branch.Address = &model.Address{}
	if r.PostalCode.Valid {
		branch.Address.PostalCode = &r.PostalCode.String
	}
	if r.RegionCode.Valid {
		branch.Address.RegionCode = &r.RegionCode.String
	}
	if r.Region.Valid {
		branch.Address.Region = &r.Region.String
	}
	if r.City.Valid {
		branch.Address.City = &r.City.String
	}
	if r.FullAddress.Valid {
		branch.Address.FullAddress = &r.FullAddress.String
	}

	return branch
}

// GetByCompanyOGRN получает филиалы компании по ОГРН
func (r *BranchRepository) GetByCompanyOGRN(ctx context.Context, ogrn string) ([]*model.Branch, error) {
	query := `
		SELECT * FROM egrul.branches
		WHERE company_ogrn = ?
		ORDER BY branch_type, branch_name NULLS LAST
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn)
	if err != nil {
		return nil, fmt.Errorf("query branches: %w", err)
	}
	defer rows.Close()

	var branches []*model.Branch
	for rows.Next() {
		var row branchRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan branch row: %w", err)
		}
		branches = append(branches, row.toModel())
	}

	return branches, nil
}

