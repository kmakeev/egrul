package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// LicenseRepository репозиторий для работы с лицензиями
type LicenseRepository struct {
	client *Client
	logger *zap.Logger
}

// NewLicenseRepository создает новый репозиторий лицензий
func NewLicenseRepository(client *Client, logger *zap.Logger) *LicenseRepository {
	return &LicenseRepository{
		client: client,
		logger: logger.Named("license_repo"),
	}
}

// licenseRow структура для сканирования результатов запроса
type licenseRow struct {
	ID            string         `ch:"id"`
	EntityType    string         `ch:"entity_type"`
	EntityOgrn    string         `ch:"entity_ogrn"`
	EntityInn     sql.NullString `ch:"entity_inn"`
	EntityName    sql.NullString `ch:"entity_name"`
	LicenseNumber string         `ch:"license_number"`
	LicenseSeries sql.NullString `ch:"license_series"`
	Activity      string         `ch:"activity"`
	StartDate     sql.NullTime   `ch:"start_date"`
	EndDate       sql.NullTime   `ch:"end_date"`
	Authority     sql.NullString `ch:"authority"`
	Status        string         `ch:"status"`
	VersionDate   time.Time      `ch:"version_date"`
	CreatedAt     time.Time      `ch:"created_at"`
	UpdatedAt     time.Time      `ch:"updated_at"`
}

func (r *licenseRow) toModel() *model.License {
	license := &model.License{
		ID:     r.ID,
		Number: r.LicenseNumber,
	}

	if r.LicenseSeries.Valid {
		license.Series = &r.LicenseSeries.String
	}
	if r.Activity != "" {
		license.Activity = &r.Activity
	}
	if r.StartDate.Valid {
		license.StartDate = &model.Date{Time: r.StartDate.Time}
	}
	if r.EndDate.Valid {
		license.EndDate = &model.Date{Time: r.EndDate.Time}
	}
	if r.Authority.Valid {
		license.Authority = &r.Authority.String
	}
	if r.Status != "" {
		license.Status = &r.Status
	}

	return license
}

// GetByEntityOGRN получает лицензии по ОГРН/ОГРНИП
func (r *LicenseRepository) GetByEntityOGRN(ctx context.Context, ogrn string) ([]*model.License, error) {
	query := `
		SELECT * FROM egrul.licenses FINAL
		WHERE entity_ogrn = ?
		ORDER BY start_date DESC NULLS LAST
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn)
	if err != nil {
		return nil, fmt.Errorf("query licenses: %w", err)
	}
	defer rows.Close()

	var licenses []*model.License
	for rows.Next() {
		var row licenseRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan license row: %w", err)
		}
		licenses = append(licenses, row.toModel())
	}

	return licenses, nil
}

