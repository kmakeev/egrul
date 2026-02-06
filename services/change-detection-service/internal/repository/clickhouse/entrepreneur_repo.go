package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/egrul/change-detection-service/internal/model"
	"go.uber.org/zap"
)

// EntrepreneurRepository реализует repository.EntrepreneurRepository для ClickHouse
type EntrepreneurRepository struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// NewEntrepreneurRepository создает новый экземпляр EntrepreneurRepository
func NewEntrepreneurRepository(conn clickhouse.Conn, logger *zap.Logger) *EntrepreneurRepository {
	return &EntrepreneurRepository{
		conn:   conn,
		logger: logger,
	}
}

// GetByOGRNIP возвращает текущую версию ИП по ОГРНИП
func (r *EntrepreneurRepository) GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error) {
	query := `
		SELECT
			ogrnip,
			inn,
			concat(last_name, ' ', first_name, ' ', coalesce(middle_name, '')) AS full_name,
			region_code,
			status,
			termination_date,
			full_address,
			postal_code,
			region,
			city,
			street,
			house,
			okved_main_code,
			registration_date,
			extract_date,
			updated_at,
			citizenship_country_code,
			gender
		FROM entrepreneurs
		WHERE ogrnip = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := r.conn.QueryRow(ctx, query, ogrnip)

	var entrepreneur model.Entrepreneur
	var terminationDate sql.NullTime
	var citizenshipCode, gender sql.NullString

	err := row.Scan(
		&entrepreneur.OGRNIP,
		&entrepreneur.INN,
		&entrepreneur.FullName,
		&entrepreneur.RegionCode,
		&entrepreneur.Status,
		&terminationDate,
		&entrepreneur.AddressFull,
		&entrepreneur.AddressPostalCode,
		&entrepreneur.AddressRegion,
		&entrepreneur.AddressCity,
		&entrepreneur.AddressStreet,
		&entrepreneur.AddressHouse,
		&entrepreneur.MainOKVED,
		&entrepreneur.RegistrationDate,
		&entrepreneur.ExtractDate,
		&entrepreneur.LastUpdate,
		&citizenshipCode,
		&gender,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entrepreneur with OGRNIP %s not found", ogrnip)
		}
		r.logger.Error("failed to get entrepreneur by OGRNIP",
			zap.String("ogrnip", ogrnip),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get entrepreneur: %w", err)
	}

	// Преобразование nullable полей
	if terminationDate.Valid {
		entrepreneur.TerminationDate = &terminationDate.Time
	}
	if citizenshipCode.Valid {
		entrepreneur.CitizenshipCode = citizenshipCode.String
	}
	if gender.Valid {
		entrepreneur.Gender = gender.String
	}

	// Загрузка связанных данных
	_, additionalOKVED, err := r.GetActivities(ctx, ogrnip)
	if err != nil {
		r.logger.Warn("failed to get activities", zap.String("ogrnip", ogrnip), zap.Error(err))
		entrepreneur.AdditionalOKVED = []string{}
	} else {
		entrepreneur.AdditionalOKVED = additionalOKVED
	}

	licensesCount, err := r.GetLicensesCount(ctx, ogrnip)
	if err != nil {
		r.logger.Warn("failed to get licenses count", zap.String("ogrnip", ogrnip), zap.Error(err))
		entrepreneur.LicensesCount = 0
	} else {
		entrepreneur.LicensesCount = licensesCount
	}

	return &entrepreneur, nil
}

// GetByOGRNIPs возвращает несколько ИП по списку ОГРНИП (для батчинга)
func (r *EntrepreneurRepository) GetByOGRNIPs(ctx context.Context, ogrnips []string) ([]*model.Entrepreneur, error) {
	if len(ogrnips) == 0 {
		return []*model.Entrepreneur{}, nil
	}

	entrepreneurs := make([]*model.Entrepreneur, 0, len(ogrnips))
	for _, ogrnip := range ogrnips {
		entrepreneur, err := r.GetByOGRNIP(ctx, ogrnip)
		if err != nil {
			r.logger.Warn("failed to get entrepreneur in batch",
				zap.String("ogrnip", ogrnip),
				zap.Error(err),
			)
			continue
		}
		entrepreneurs = append(entrepreneurs, entrepreneur)
	}

	return entrepreneurs, nil
}

// GetActivities возвращает виды деятельности ИП
func (r *EntrepreneurRepository) GetActivities(ctx context.Context, ogrnip string) (string, []string, error) {
	// Основной ОКВЭД уже получен в GetByOGRNIP, здесь получаем дополнительные
	query := `
		SELECT DISTINCT okved_code
		FROM entrepreneur_okved_extra
		WHERE entrepreneur_ogrnip = ?
		ORDER BY okved_code
	`

	rows, err := r.conn.Query(ctx, query, ogrnip)
	if err != nil {
		return "", nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var additionalOKVED []string
	for rows.Next() {
		var okvedCode string
		if err := rows.Scan(&okvedCode); err != nil {
			r.logger.Warn("failed to scan okved code", zap.Error(err))
			continue
		}
		additionalOKVED = append(additionalOKVED, okvedCode)
	}

	return "", additionalOKVED, nil
}

// GetLicensesCount возвращает количество лицензий ИП
func (r *EntrepreneurRepository) GetLicensesCount(ctx context.Context, ogrnip string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT license_number)
		FROM licenses
		WHERE entrepreneur_ogrnip = ?
	`

	var count uint64
	err := r.conn.QueryRow(ctx, query, ogrnip).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get licenses count: %w", err)
	}

	return int(count), nil
}

// GetPreviousByOGRNIP возвращает предыдущую версию ИП (с максимальной extract_date меньше текущей)
func (r *EntrepreneurRepository) GetPreviousByOGRNIP(ctx context.Context, ogrnip string, beforeDate string) (*model.Entrepreneur, error) {
	query := `
		SELECT
			ogrnip,
			inn,
			concat(last_name, ' ', first_name, ' ', coalesce(middle_name, '')) AS full_name,
			region_code,
			status,
			termination_date,
			full_address,
			postal_code,
			region,
			city,
			street,
			house,
			okved_main_code,
			registration_date,
			extract_date,
			updated_at,
			citizenship_country_code,
			gender
		FROM entrepreneurs
		WHERE ogrnip = ? AND updated_at < toDateTime(?)
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := r.conn.QueryRow(ctx, query, ogrnip, beforeDate)

	var entrepreneur model.Entrepreneur
	var terminationDate sql.NullTime
	var citizenshipCode, gender sql.NullString

	err := row.Scan(
		&entrepreneur.OGRNIP,
		&entrepreneur.INN,
		&entrepreneur.FullName,
		&entrepreneur.RegionCode,
		&entrepreneur.Status,
		&terminationDate,
		&entrepreneur.AddressFull,
		&entrepreneur.AddressPostalCode,
		&entrepreneur.AddressRegion,
		&entrepreneur.AddressCity,
		&entrepreneur.AddressStreet,
		&entrepreneur.AddressHouse,
		&entrepreneur.MainOKVED,
		&entrepreneur.RegistrationDate,
		&entrepreneur.ExtractDate,
		&entrepreneur.LastUpdate,
		&citizenshipCode,
		&gender,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет предыдущей версии - это не ошибка
		}
		r.logger.Error("failed to get previous entrepreneur version by OGRNIP",
			zap.String("ogrnip", ogrnip),
			zap.String("before_date", beforeDate),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get previous entrepreneur version: %w", err)
	}

	// Преобразование nullable полей
	if terminationDate.Valid {
		entrepreneur.TerminationDate = &terminationDate.Time
	}
	if citizenshipCode.Valid {
		entrepreneur.CitizenshipCode = citizenshipCode.String
	}
	if gender.Valid {
		entrepreneur.Gender = gender.String
	}

	// Загрузка связанных данных (опционально, для упрощения можно пропустить для предыдущей версии)
	_, additionalOKVED, err := r.GetActivities(ctx, ogrnip)
	if err != nil {
		entrepreneur.AdditionalOKVED = []string{}
	} else {
		entrepreneur.AdditionalOKVED = additionalOKVED
	}

	licensesCount, err := r.GetLicensesCount(ctx, ogrnip)
	if err != nil {
		entrepreneur.LicensesCount = 0
	} else {
		entrepreneur.LicensesCount = licensesCount
	}

	return &entrepreneur, nil
}
