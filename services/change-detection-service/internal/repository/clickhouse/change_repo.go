package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/egrul/change-detection-service/internal/model"
	"go.uber.org/zap"
)

// ChangeRepository реализует repository.ChangeRepository для ClickHouse
type ChangeRepository struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// NewChangeRepository создает новый экземпляр ChangeRepository
func NewChangeRepository(conn clickhouse.Conn, logger *zap.Logger) *ChangeRepository {
	return &ChangeRepository{
		conn:   conn,
		logger: logger,
	}
}

// SaveCompanyChange сохраняет событие изменения компании
func (r *ChangeRepository) SaveCompanyChange(ctx context.Context, change *model.ChangeEvent) error {
	if err := change.Validate(); err != nil {
		return fmt.Errorf("invalid change event: %w", err)
	}

	if !change.IsCompany() {
		return fmt.Errorf("change event is not for a company")
	}

	query := `
		INSERT INTO company_changes (
			ogrn,
			change_type,
			change_id,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			region_code,
			inn
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.conn.Exec(ctx, query,
		change.EntityID,
		string(change.ChangeType),
		change.ChangeID,
		change.FieldName,
		change.OldValue,
		change.NewValue,
		boolToUInt8(change.IsSignificant),
		change.DetectedAt,
		change.Description,
		change.RegionCode,
		change.INN,
	)

	if err != nil {
		r.logger.Error("failed to save company change",
			zap.String("ogrn", change.EntityID),
			zap.String("change_type", string(change.ChangeType)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save company change: %w", err)
	}

	r.logger.Debug("saved company change",
		zap.String("ogrn", change.EntityID),
		zap.String("change_type", string(change.ChangeType)),
		zap.String("change_id", change.ChangeID),
	)

	return nil
}

// SaveCompanyChanges сохраняет несколько событий изменений компаний (батчинг)
func (r *ChangeRepository) SaveCompanyChanges(ctx context.Context, changes []*model.ChangeEvent) error {
	if len(changes) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO company_changes (
			ogrn,
			change_type,
			change_id,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			region_code,
			inn
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, change := range changes {
		if err := change.Validate(); err != nil {
			r.logger.Warn("skipping invalid change event", zap.Error(err))
			continue
		}

		if !change.IsCompany() {
			r.logger.Warn("skipping non-company change event", zap.String("entity_type", change.EntityType))
			continue
		}

		err = batch.Append(
			change.EntityID,
			string(change.ChangeType),
			change.ChangeID,
			change.FieldName,
			change.OldValue,
			change.NewValue,
			boolToUInt8(change.IsSignificant),
			change.DetectedAt,
			change.Description,
			change.RegionCode,
			change.INN,
		)
		if err != nil {
			r.logger.Error("failed to append change to batch",
				zap.String("ogrn", change.EntityID),
				zap.Error(err),
			)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.Error("failed to send batch", zap.Error(err))
		return fmt.Errorf("failed to send batch: %w", err)
	}

	r.logger.Info("saved company changes batch",
		zap.Int("count", len(changes)),
	)

	return nil
}

// SaveEntrepreneurChange сохраняет событие изменения ИП
func (r *ChangeRepository) SaveEntrepreneurChange(ctx context.Context, change *model.ChangeEvent) error {
	if err := change.Validate(); err != nil {
		return fmt.Errorf("invalid change event: %w", err)
	}

	if !change.IsEntrepreneur() {
		return fmt.Errorf("change event is not for an entrepreneur")
	}

	query := `
		INSERT INTO entrepreneur_changes (
			ogrnip,
			change_type,
			change_id,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			inn,
			full_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := r.conn.Exec(ctx, query,
		change.EntityID,
		string(change.ChangeType),
		change.ChangeID,
		change.FieldName,
		change.OldValue,
		change.NewValue,
		boolToUInt8(change.IsSignificant),
		change.DetectedAt,
		change.Description,
		change.INN,
		change.EntityName,
	)

	if err != nil {
		r.logger.Error("failed to save entrepreneur change",
			zap.String("ogrnip", change.EntityID),
			zap.String("change_type", string(change.ChangeType)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to save entrepreneur change: %w", err)
	}

	r.logger.Debug("saved entrepreneur change",
		zap.String("ogrnip", change.EntityID),
		zap.String("change_type", string(change.ChangeType)),
		zap.String("change_id", change.ChangeID),
	)

	return nil
}

// SaveEntrepreneurChanges сохраняет несколько событий изменений ИП (батчинг)
func (r *ChangeRepository) SaveEntrepreneurChanges(ctx context.Context, changes []*model.ChangeEvent) error {
	if len(changes) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO entrepreneur_changes (
			ogrnip,
			change_type,
			change_id,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			inn,
			full_name
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, change := range changes {
		if err := change.Validate(); err != nil {
			r.logger.Warn("skipping invalid change event", zap.Error(err))
			continue
		}

		if !change.IsEntrepreneur() {
			r.logger.Warn("skipping non-entrepreneur change event", zap.String("entity_type", change.EntityType))
			continue
		}

		err = batch.Append(
			change.EntityID,
			string(change.ChangeType),
			change.ChangeID,
			change.FieldName,
			change.OldValue,
			change.NewValue,
			boolToUInt8(change.IsSignificant),
			change.DetectedAt,
			change.Description,
			change.INN,
			change.EntityName,
		)
		if err != nil {
			r.logger.Error("failed to append change to batch",
				zap.String("ogrnip", change.EntityID),
				zap.Error(err),
			)
			continue
		}
	}

	if err := batch.Send(); err != nil {
		r.logger.Error("failed to send batch", zap.Error(err))
		return fmt.Errorf("failed to send batch: %w", err)
	}

	r.logger.Info("saved entrepreneur changes batch",
		zap.Int("count", len(changes)),
	)

	return nil
}

// GetCompanyChanges возвращает историю изменений компании
func (r *ChangeRepository) GetCompanyChanges(ctx context.Context, ogrn string, limit int) ([]*model.ChangeEvent, error) {
	query := `
		SELECT
			change_id,
			ogrn,
			change_type,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			region_code,
			inn
		FROM company_changes
		WHERE ogrn = ?
		ORDER BY detected_at DESC
		LIMIT ?
	`

	rows, err := r.conn.Query(ctx, query, ogrn, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query company changes: %w", err)
	}
	defer rows.Close()

	var changes []*model.ChangeEvent
	for rows.Next() {
		var change model.ChangeEvent
		var isSignificant uint8
		var changeType string

		err := rows.Scan(
			&change.ChangeID,
			&change.EntityID,
			&changeType,
			&change.FieldName,
			&change.OldValue,
			&change.NewValue,
			&isSignificant,
			&change.DetectedAt,
			&change.Description,
			&change.RegionCode,
			&change.INN,
		)
		if err != nil {
			r.logger.Warn("failed to scan change event", zap.Error(err))
			continue
		}

		change.EntityType = "company"
		change.ChangeType = model.ChangeType(changeType)
		change.IsSignificant = isSignificant == 1

		changes = append(changes, &change)
	}

	return changes, nil
}

// GetEntrepreneurChanges возвращает историю изменений ИП
func (r *ChangeRepository) GetEntrepreneurChanges(ctx context.Context, ogrnip string, limit int) ([]*model.ChangeEvent, error) {
	query := `
		SELECT
			change_id,
			ogrnip,
			change_type,
			field_name,
			old_value,
			new_value,
			is_significant,
			detected_at,
			change_description,
			region_code,
			inn
		FROM entrepreneur_changes
		WHERE ogrnip = ?
		ORDER BY detected_at DESC
		LIMIT ?
	`

	rows, err := r.conn.Query(ctx, query, ogrnip, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query entrepreneur changes: %w", err)
	}
	defer rows.Close()

	var changes []*model.ChangeEvent
	for rows.Next() {
		var change model.ChangeEvent
		var isSignificant uint8
		var changeType string

		err := rows.Scan(
			&change.ChangeID,
			&change.EntityID,
			&changeType,
			&change.FieldName,
			&change.OldValue,
			&change.NewValue,
			&isSignificant,
			&change.DetectedAt,
			&change.Description,
			&change.RegionCode,
			&change.INN,
		)
		if err != nil {
			r.logger.Warn("failed to scan change event", zap.Error(err))
			continue
		}

		change.EntityType = "entrepreneur"
		change.ChangeType = model.ChangeType(changeType)
		change.IsSignificant = isSignificant == 1

		changes = append(changes, &change)
	}

	return changes, nil
}

// GetRecentChanges возвращает последние изменения за период (для тестирования)
func (r *ChangeRepository) GetRecentChanges(ctx context.Context, entityType string, since int64) ([]*model.ChangeEvent, error) {
	var query string
	sinceTime := time.Unix(since, 0)

	if entityType == "company" {
		query = `
			SELECT
				change_id,
				ogrn,
				change_type,
				field_name,
				old_value,
				new_value,
				is_significant,
				detected_at,
				change_description,
				region_code,
				inn
			FROM company_changes
			WHERE detected_at >= ?
			ORDER BY detected_at DESC
			LIMIT 100
		`
	} else {
		query = `
			SELECT
				change_id,
				ogrnip,
				change_type,
				field_name,
				old_value,
				new_value,
				is_significant,
				detected_at,
				change_description,
				region_code,
				inn
			FROM entrepreneur_changes
			WHERE detected_at >= ?
			ORDER BY detected_at DESC
			LIMIT 100
		`
	}

	rows, err := r.conn.Query(ctx, query, sinceTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent changes: %w", err)
	}
	defer rows.Close()

	var changes []*model.ChangeEvent
	for rows.Next() {
		var change model.ChangeEvent
		var isSignificant uint8
		var changeType string

		err := rows.Scan(
			&change.ChangeID,
			&change.EntityID,
			&changeType,
			&change.FieldName,
			&change.OldValue,
			&change.NewValue,
			&isSignificant,
			&change.DetectedAt,
			&change.Description,
			&change.RegionCode,
			&change.INN,
		)
		if err != nil {
			r.logger.Warn("failed to scan change event", zap.Error(err))
			continue
		}

		change.EntityType = entityType
		change.ChangeType = model.ChangeType(changeType)
		change.IsSignificant = isSignificant == 1

		changes = append(changes, &change)
	}

	return changes, nil
}

// boolToUInt8 конвертирует bool в uint8 для ClickHouse
func boolToUInt8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
