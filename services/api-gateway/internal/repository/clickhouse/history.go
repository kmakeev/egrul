package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

// historyRow структура для сканирования результатов запроса (старая таблица)
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
	SnapshotJSON        sql.NullString `ch:"snapshot_json"`
	SourceFile          sql.NullString `ch:"source_file"`
	ExtractDate         sql.NullTime   `ch:"extract_date"`
	CreatedAt           sql.NullTime   `ch:"created_at"`
	UpdatedAt           sql.NullTime   `ch:"updated_at"`
}

// historyRowV2 структура для сканирования результатов запроса (оптимизированная таблица)
type historyRowOptimized struct {
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
	SnapshotJSON        sql.NullString `ch:"snapshot_json"`
	SourceFiles         []string       `ch:"source_files"`
	ExtractDate         sql.NullTime   `ch:"extract_date"`
	FileHash            sql.NullString `ch:"file_hash"`
	CreatedAt           sql.NullTime   `ch:"created_at"`
	UpdatedAt           sql.NullTime   `ch:"updated_at"`
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

func (r *historyRowOptimized) toModel() *model.HistoryRecord {
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

	// Используем оптимизированное представление для получения дедуплицированных данных
	query := `
		SELECT 
			id, entity_type, entity_id, inn, grn, grn_date,
			reason_code, reason_description, authority_code, authority_name,
			certificate_series, certificate_number, certificate_date,
			snapshot_full_name, snapshot_status, snapshot_address,
			snapshot_json, source_files, extract_date, file_hash, created_at, updated_at
		FROM egrul.company_history_view
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
		var row historyRowOptimized
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
			&row.SnapshotJSON,
			&row.SourceFiles,
			&row.ExtractDate,
			&row.FileHash,
			&row.CreatedAt,
			&row.UpdatedAt,
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

// CountByEntityID получает общее количество записей истории для сущности
func (r *HistoryRepository) CountByEntityID(ctx context.Context, entityType, entityID string) (int, error) {
	r.logger.Info("CountByEntityID called",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	// Используем оптимизированное представление для подсчета дедуплицированных записей
	query := `
		SELECT count(*) FROM egrul.company_history_view
		WHERE entity_type = ? AND entity_id = ?
	`

	var count uint64
	err := r.client.conn.QueryRow(ctx, query, entityType, entityID).Scan(&count)
	if err != nil {
		r.logger.Error("count history failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err))
		return 0, fmt.Errorf("count history: %w", err)
	}

	result := int(count)
	r.logger.Info("CountByEntityID completed",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.Int("count", result))
	return result, nil
}

// InsertOrUpdate вставляет или обновляет запись истории с автоматической дедупликацией
func (r *HistoryRepository) InsertOrUpdate(ctx context.Context, record *model.HistoryRecord, entityType, entityID string, extractDate string, sourceFile string, fileHash string) error {
	r.logger.Info("InsertOrUpdate called",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.String("grn", record.Grn),
		zap.String("source_file", sourceFile))

	// Используем простую вставку - ReplacingMergeTree автоматически обработает дедупликацию
	query := `
		INSERT INTO egrul.company_history (
			id,
			entity_type,
			entity_id,
			inn,
			grn,
			grn_date,
			reason_code,
			reason_description,
			authority_code,
			authority_name,
			certificate_series,
			certificate_number,
			certificate_date,
			snapshot_full_name,
			snapshot_status,
			snapshot_address,
			snapshot_json,
			source_files,
			extract_date,
			file_hash,
			created_at,
			updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	// Подготавливаем значения
	var inn *string
	// HistoryRecord doesn't have Inn field, so we skip it
	// if record.Inn != nil {
	// 	inn = record.Inn
	// }

	var grnDate *string
	if !record.Date.IsZero() {
		dateStr := record.Date.Format("2006-01-02")
		grnDate = &dateStr
	}

	var reasonCode *string
	if record.ReasonCode != nil {
		reasonCode = record.ReasonCode
	}

	var reasonDescription *string
	if record.ReasonDescription != nil {
		reasonDescription = record.ReasonDescription
	}

	var authorityCode *string
	var authorityName *string
	if record.Authority != nil {
		if record.Authority.Code != nil {
			authorityCode = record.Authority.Code
		}
		if record.Authority.Name != nil {
			authorityName = record.Authority.Name
		}
	}

	var certificateSeries *string
	if record.CertificateSeries != nil {
		certificateSeries = record.CertificateSeries
	}

	var certificateNumber *string
	if record.CertificateNumber != nil {
		certificateNumber = record.CertificateNumber
	}

	var certificateDate *string
	if record.CertificateDate != nil && !record.CertificateDate.IsZero() {
		dateStr := record.CertificateDate.Format("2006-01-02")
		certificateDate = &dateStr
	}

	var snapshotFullName *string
	if record.SnapshotFullName != nil {
		snapshotFullName = record.SnapshotFullName
	}

	var snapshotStatus *string
	if record.SnapshotStatus != nil {
		snapshotStatus = record.SnapshotStatus
	}

	var snapshotAddress *string
	if record.SnapshotAddress != nil {
		snapshotAddress = record.SnapshotAddress
	}

	now := "now64(3)"
	
	err := r.client.conn.Exec(ctx, query,
		record.ID,
		entityType,
		entityID,
		inn,
		record.Grn,
		grnDate,
		reasonCode,
		reasonDescription,
		authorityCode,
		authorityName,
		certificateSeries,
		certificateNumber,
		certificateDate,
		snapshotFullName,
		snapshotStatus,
		snapshotAddress,
		nil, // snapshot_json
		[]string{sourceFile}, // source_files как массив
		extractDate,
		fileHash,
		now,
		now,
	)

	if err != nil {
		r.logger.Error("insert history failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.String("grn", record.Grn),
			zap.Error(err))
		return fmt.Errorf("insert history: %w", err)
	}

	r.logger.Info("InsertOrUpdate completed",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
		zap.String("grn", record.Grn))
	return nil
}

// BatchInsert выполняет пакетную вставку записей истории для повышения производительности
func (r *HistoryRepository) BatchInsert(ctx context.Context, records []BatchHistoryRecord) error {
	if len(records) == 0 {
		return nil
	}

	r.logger.Info("BatchInsert called", zap.Int("count", len(records)))

	query := `
		INSERT INTO egrul.company_history (
			id, entity_type, entity_id, inn, grn, grn_date,
			reason_code, reason_description, authority_code, authority_name,
			certificate_series, certificate_number, certificate_date,
			snapshot_full_name, snapshot_status, snapshot_address, snapshot_json,
			source_files, extract_date, file_hash, created_at, updated_at
		) VALUES
	`

	// Подготавливаем значения для batch insert
	values := make([]string, len(records))
	args := make([]interface{}, 0, len(records)*22)

	for i, record := range records {
		values[i] = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		
		// Добавляем аргументы для каждой записи
		args = append(args,
			record.ID,
			record.EntityType,
			record.EntityID,
			record.Inn,
			record.Grn,
			record.GrnDate,
			record.ReasonCode,
			record.ReasonDescription,
			record.AuthorityCode,
			record.AuthorityName,
			record.CertificateSeries,
			record.CertificateNumber,
			record.CertificateDate,
			record.SnapshotFullName,
			record.SnapshotStatus,
			record.SnapshotAddress,
			record.SnapshotJSON,
			[]string{record.SourceFile}, // source_files как массив
			record.ExtractDate,
			record.FileHash,
			"now64(3)", // created_at
			"now64(3)", // updated_at
		)
	}

	fullQuery := query + " " + strings.Join(values, ", ")

	err := r.client.conn.Exec(ctx, fullQuery, args...)
	if err != nil {
		r.logger.Error("batch insert history failed", zap.Error(err))
		return fmt.Errorf("batch insert history: %w", err)
	}

	r.logger.Info("BatchInsert completed", zap.Int("count", len(records)))
	return nil
}

// BatchHistoryRecord структура для пакетной вставки
type BatchHistoryRecord struct {
	ID                  string
	EntityType          string
	EntityID            string
	Inn                 *string
	Grn                 string
	GrnDate             *string
	ReasonCode          *string
	ReasonDescription   *string
	AuthorityCode       *string
	AuthorityName       *string
	CertificateSeries   *string
	CertificateNumber   *string
	CertificateDate     *string
	SnapshotFullName    *string
	SnapshotStatus      *string
	SnapshotAddress     *string
	SnapshotJSON        *string
	SourceFile          string
	ExtractDate         string
	FileHash            string
}

// OptimizeTable принудительно запускает дедупликацию для таблицы истории
func (r *HistoryRepository) OptimizeTable(ctx context.Context) error {
	r.logger.Info("OptimizeTable called")

	query := `OPTIMIZE TABLE egrul.company_history FINAL`
	
	err := r.client.conn.Exec(ctx, query)
	if err != nil {
		r.logger.Error("optimize table failed", zap.Error(err))
		return fmt.Errorf("optimize table: %w", err)
	}

	r.logger.Info("OptimizeTable completed")
	return nil
}
