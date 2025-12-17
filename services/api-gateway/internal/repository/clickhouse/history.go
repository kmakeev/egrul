package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// HistoryRepository репозиторий для работы с историей изменений
type HistoryRepository struct {
	client *Client
	logger *zap.Logger
}

// NewHistoryRepository создает новый репозиторий истории
func NewHistoryRepository(client *Client, logger *zap.Logger) *HistoryRepository {
	return &HistoryRepository{
		client: client,
		logger: logger.Named("history_repo"),
	}
}

// historyRow структура для сканирования результатов запроса
type historyRow struct {
	ID                string         `ch:"id"`
	EntityType        string         `ch:"entity_type"`
	EntityID          string         `ch:"entity_id"`
	Inn               sql.NullString `ch:"inn"`
	Grn               string         `ch:"grn"`
	GrnDate           time.Time      `ch:"grn_date"`
	ReasonCode        sql.NullString `ch:"reason_code"`
	ReasonDescription sql.NullString `ch:"reason_description"`
	AuthorityCode     sql.NullString `ch:"authority_code"`
	AuthorityName     sql.NullString `ch:"authority_name"`
	CertificateSeries sql.NullString `ch:"certificate_series"`
	CertificateNumber sql.NullString `ch:"certificate_number"`
	CertificateDate   sql.NullTime   `ch:"certificate_date"`
	SnapshotFullName  sql.NullString `ch:"snapshot_full_name"`
	SnapshotStatus    sql.NullString `ch:"snapshot_status"`
	SnapshotAddress   sql.NullString `ch:"snapshot_address"`
	SourceFile        sql.NullString `ch:"source_file"`
	CreatedAt         time.Time      `ch:"created_at"`
}

func (r *historyRow) toModel() *model.HistoryRecord {
	record := &model.HistoryRecord{
		ID:   r.ID,
		Grn:  r.Grn,
		Date: model.Date{Time: r.GrnDate},
	}

	if r.ReasonCode.Valid {
		record.ReasonCode = &r.ReasonCode.String
	}
	if r.ReasonDescription.Valid {
		record.ReasonDescription = &r.ReasonDescription.String
	}
	if r.AuthorityCode.Valid || r.AuthorityName.Valid {
		record.Authority = &model.Authority{}
		if r.AuthorityCode.Valid {
			record.Authority.Code = &r.AuthorityCode.String
		}
		if r.AuthorityName.Valid {
			record.Authority.Name = &r.AuthorityName.String
		}
	}
	if r.CertificateSeries.Valid {
		record.CertificateSeries = &r.CertificateSeries.String
	}
	if r.CertificateNumber.Valid {
		record.CertificateNumber = &r.CertificateNumber.String
	}
	if r.CertificateDate.Valid {
		record.CertificateDate = &model.Date{Time: r.CertificateDate.Time}
	}
	if r.SnapshotFullName.Valid {
		record.SnapshotFullName = &r.SnapshotFullName.String
	}
	if r.SnapshotStatus.Valid {
		record.SnapshotStatus = &r.SnapshotStatus.String
	}
	if r.SnapshotAddress.Valid {
		record.SnapshotAddress = &r.SnapshotAddress.String
	}

	return record
}

// GetByEntityID получает историю по ID сущности
func (r *HistoryRepository) GetByEntityID(ctx context.Context, entityType model.EntityType, entityID string, limit, offset int) ([]*model.HistoryRecord, error) {
	query := `
		SELECT * FROM egrul.company_history
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY grn_date DESC
		LIMIT ? OFFSET ?
	`

	dbEntityType := "company"
	if entityType == model.EntityTypeEntrepreneur {
		dbEntityType = "entrepreneur"
	}

	rows, err := r.client.conn.Query(ctx, query, dbEntityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var records []*model.HistoryRecord
	for rows.Next() {
		var row historyRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		records = append(records, row.toModel())
	}

	return records, nil
}

