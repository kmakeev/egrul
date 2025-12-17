package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// FounderRepository репозиторий для работы с учредителями
type FounderRepository struct {
	client *Client
	logger *zap.Logger
}

// NewFounderRepository создает новый репозиторий учредителей
func NewFounderRepository(client *Client, logger *zap.Logger) *FounderRepository {
	return &FounderRepository{
		client: client,
		logger: logger.Named("founder_repo"),
	}
}

// founderRow структура для сканирования результатов запроса
type founderRow struct {
	ID                string          `ch:"id"`
	CompanyOgrn       string          `ch:"company_ogrn"`
	CompanyInn        sql.NullString  `ch:"company_inn"`
	CompanyName       sql.NullString  `ch:"company_name"`
	FounderType       string          `ch:"founder_type"`
	FounderOgrn       sql.NullString  `ch:"founder_ogrn"`
	FounderInn        string          `ch:"founder_inn"`
	FounderName       string          `ch:"founder_name"`
	FounderLastName   sql.NullString  `ch:"founder_last_name"`
	FounderFirstName  sql.NullString  `ch:"founder_first_name"`
	FounderMiddleName sql.NullString  `ch:"founder_middle_name"`
	FounderCountry    sql.NullString  `ch:"founder_country"`
	FounderCitizenship sql.NullString `ch:"founder_citizenship"`
	ShareNominalValue sql.NullFloat64 `ch:"share_nominal_value"`
	SharePercent      sql.NullFloat64 `ch:"share_percent"`
}

func (r *founderRow) toModel() *model.Founder {
	founder := &model.Founder{
		Type: model.ParseFounderType(r.FounderType),
		Name: r.FounderName,
	}

	if r.FounderOgrn.Valid && r.FounderOgrn.String != "" {
		founder.Ogrn = &r.FounderOgrn.String
	}
	if r.FounderInn != "" {
		founder.Inn = &r.FounderInn
	}
	if r.FounderLastName.Valid {
		founder.LastName = &r.FounderLastName.String
	}
	if r.FounderFirstName.Valid {
		founder.FirstName = &r.FounderFirstName.String
	}
	if r.FounderMiddleName.Valid {
		founder.MiddleName = &r.FounderMiddleName.String
	}
	if r.FounderCountry.Valid {
		founder.Country = &r.FounderCountry.String
	}
	if r.FounderCitizenship.Valid {
		founder.Citizenship = &r.FounderCitizenship.String
	}
	if r.ShareNominalValue.Valid {
		founder.ShareNominalValue = &r.ShareNominalValue.Float64
	}
	if r.SharePercent.Valid {
		founder.SharePercent = &r.SharePercent.Float64
	}

	return founder
}

// GetByCompanyOGRN получает учредителей компании по ОГРН
func (r *FounderRepository) GetByCompanyOGRN(ctx context.Context, ogrn string, limit, offset int) ([]*model.Founder, error) {
	query := `
		SELECT * FROM egrul.founders FINAL
		WHERE company_ogrn = ?
		ORDER BY share_percent DESC NULLS LAST, founder_name
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query founders: %w", err)
	}
	defer rows.Close()

	var founders []*model.Founder
	for rows.Next() {
		var row founderRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan founder row: %w", err)
		}
		founders = append(founders, row.toModel())
	}

	return founders, nil
}

// GetRelatedCompanies получает компании где лицо является учредителем
func (r *FounderRepository) GetRelatedCompanies(ctx context.Context, inn string, limit, offset int) ([]string, error) {
	query := `
		SELECT DISTINCT company_ogrn FROM egrul.founders FINAL
		WHERE founder_inn = ?
		ORDER BY company_ogrn
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.conn.Query(ctx, query, inn, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query related companies: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var ogrn string
		if err := rows.Scan(&ogrn); err != nil {
			return nil, fmt.Errorf("scan ogrn: %w", err)
		}
		ogrns = append(ogrns, ogrn)
	}

	return ogrns, nil
}

