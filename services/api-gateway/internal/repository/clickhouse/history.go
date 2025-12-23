package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

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
	ID                  string         `ch:"id"`
	EntityType          string         `ch:"entity_type"`
	EntityID            string         `ch:"entity_id"`
	Inn                 sql.NullString `ch:"inn"`
	Grn                 string         `ch:"grn"`
	GrnDate             sql.NullTime   `ch:"grn_date"`
	ReasonCode          sql.NullString `ch:"reason_code"`
	ReasonDescription   sql.NullString `ch:"reason_description"`
	AuthorityCode       sql.NullString `ch:"authority_code"`
	AuthorityName       sql.NullString `ch:"authority_name"`
	CertificateSeries   sql.NullString `ch:"certificate_series"`
	CertificateNumber   sql.NullString `ch:"certificate_number"`
	CertificateDate     sql.NullTime   `ch:"certificate_date"`
	SnapshotFullName    sql.NullString `ch:"snapshot_full_name"`
	SnapshotStatus      sql.NullString `ch:"snapshot_status"`
	SnapshotAddress     sql.NullString `ch:"snapshot_address"`
}

func (r *historyRow) toModel() *model.HistoryRecord {
	record := &model.HistoryRecord{
		ID:  r.ID,
		Grn: r.Grn,
	}

	if r.GrnDate.Valid {
		date := model.NewDate(r.GrnDate.Time)
		if date != nil {
			record.Date = *date
		}
	}
	if r.ReasonCode.Valid {
		record.ReasonCode = &r.ReasonCode.String
	}
	if r.ReasonDescription.Valid {
		record.ReasonDescription = &r.ReasonDescription.String
	}
	if r.AuthorityCode.Valid || r.AuthorityName.Valid {
		authority := &model.Authority{}
		if r.AuthorityCode.Valid {
			authority.Code = &r.AuthorityCode.String
		}
		if r.AuthorityName.Valid {
			authority.Name = &r.AuthorityName.String
		}
		record.Authority = authority
	}
	if r.CertificateSeries.Valid {
		record.CertificateSeries = &r.CertificateSeries.String
	}
	if r.CertificateNumber.Valid {
		record.CertificateNumber = &r.CertificateNumber.String
	}
	if r.CertificateDate.Valid {
		record.CertificateDate = model.NewDate(r.CertificateDate.Time)
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

// GetByEntityID получает историю изменений для сущности
func (r *HistoryRepository) GetByEntityID(ctx context.Context, entityType, entityID string, limit, offset int) ([]*model.HistoryRecord, error) {
	r.logger.Info("GetByEntityID called",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	query := `
		SELECT 
			id, entity_type, entity_id, inn,
			grn, grn_date,
			reason_code, reason_description,
			authority_code, authority_name,
			certificate_series, certificate_number, certificate_date,
			snapshot_full_name, snapshot_status, snapshot_address
		FROM egrul.company_history FINAL
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY grn_date DESC, grn DESC
		LIMIT ? OFFSET ?
	`

	r.logger.Info("Executing query", zap.String("query", query))

	rows, err := r.client.conn.Query(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		r.logger.Error("query history failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err))
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var records []*model.HistoryRecord
	for rows.Next() {
		var row historyRow
		if err := rows.Scan(
			&row.ID,
			&row.EntityType,
			&row.EntityID,
			&row.Inn,
			&row.Grn,
			&row.GrnDate,
			&row.ReasonCode,
			&row.ReasonDescription,
			&row.AuthorityCode,
			&row.AuthorityName,
			&row.CertificateSeries,
			&row.CertificateNumber,
			&row.CertificateDate,
			&row.SnapshotFullName,
			&row.SnapshotStatus,
			&row.SnapshotAddress,
		); err != nil {
			r.logger.Error("scan history row failed", zap.Error(err))
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		records = append(records, row.toModel())
	}

	r.logger.Info("GetByEntityID completed",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.Int("count", len(records)))
	return records, nil
}
