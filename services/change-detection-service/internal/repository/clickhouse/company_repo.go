package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/egrul/change-detection-service/internal/model"
	"go.uber.org/zap"
)

// CompanyRepository реализует repository.CompanyRepository для ClickHouse
type CompanyRepository struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// NewCompanyRepository создает новый экземпляр CompanyRepository
func NewCompanyRepository(conn clickhouse.Conn, logger *zap.Logger) *CompanyRepository {
	return &CompanyRepository{
		conn:   conn,
		logger: logger,
	}
}

// GetByOGRN возвращает текущую версию компании по ОГРН
func (r *CompanyRepository) GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error) {
	query := `
		SELECT
			ogrn,
			inn,
			kpp,
			full_name,
			short_name,
			region_code,
			status,
			concat(head_last_name, ' ', head_first_name, ' ', head_middle_name) as director_full_name,
			head_inn,
			head_position,
			full_address,
			postal_code,
			region,
			city,
			street,
			house,
			capital_amount,
			capital_currency,
			okved_main_code,
			registration_date,
			extract_date,
			updated_at
		FROM companies
		WHERE ogrn = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := r.conn.QueryRow(ctx, query, ogrn)

	var company model.Company
	var authorizedCapital sql.NullFloat64
	var capitalCurrency, directorFullName, directorINN, directorPosition, addressFull, postalCode, addressRegion, addressCity, addressStreet, addressHouse sql.NullString

	err := row.Scan(
		&company.OGRN,
		&company.INN,
		&company.KPP,
		&company.FullName,
		&company.ShortName,
		&company.RegionCode,
		&company.Status,
		&directorFullName,
		&directorINN,
		&directorPosition,
		&addressFull,
		&postalCode,
		&addressRegion,
		&addressCity,
		&addressStreet,
		&addressHouse,
		&authorizedCapital,
		&capitalCurrency,
		&company.MainOKVED,
		&company.RegistrationDate,
		&company.ExtractDate,
		&company.LastUpdate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("company with OGRN %s not found", ogrn)
		}
		r.logger.Error("failed to get company by OGRN",
			zap.String("ogrn", ogrn),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Преобразование nullable полей
	if directorFullName.Valid {
		company.DirectorFullName = directorFullName.String
	}
	if directorINN.Valid {
		company.DirectorINN = directorINN.String
	}
	if directorPosition.Valid {
		company.DirectorPosition = directorPosition.String
	}
	if addressFull.Valid {
		company.AddressFull = addressFull.String
	}
	if postalCode.Valid {
		company.AddressPostalCode = postalCode.String
	}
	if addressRegion.Valid {
		company.AddressRegion = addressRegion.String
	}
	if addressCity.Valid {
		company.AddressCity = addressCity.String
	}
	if addressStreet.Valid {
		company.AddressStreet = addressStreet.String
	}
	if addressHouse.Valid {
		company.AddressHouse = addressHouse.String
	}
	if authorizedCapital.Valid {
		company.AuthorizedCapital = authorizedCapital.Float64
	}
	if capitalCurrency.Valid {
		company.CapitalCurrency = capitalCurrency.String
	}

	// Загрузка связанных данных
	founders, err := r.GetFounders(ctx, ogrn)
	if err != nil {
		r.logger.Warn("failed to get founders", zap.String("ogrn", ogrn), zap.Error(err))
		// Не критично, продолжаем
		company.Founders = []model.Founder{}
	} else {
		company.Founders = founders
	}

	_, additionalOKVED, err := r.GetActivities(ctx, ogrn)
	if err != nil {
		r.logger.Warn("failed to get activities", zap.String("ogrn", ogrn), zap.Error(err))
		company.AdditionalOKVED = []string{}
	} else {
		company.AdditionalOKVED = additionalOKVED
	}

	licensesCount, err := r.GetLicensesCount(ctx, ogrn)
	if err != nil {
		r.logger.Warn("failed to get licenses count", zap.String("ogrn", ogrn), zap.Error(err))
		company.LicensesCount = 0
	} else {
		company.LicensesCount = licensesCount
	}

	branchesCount, err := r.GetBranchesCount(ctx, ogrn)
	if err != nil {
		r.logger.Warn("failed to get branches count", zap.String("ogrn", ogrn), zap.Error(err))
		company.BranchesCount = 0
	} else {
		company.BranchesCount = branchesCount
	}

	return &company, nil
}

// GetByOGRNs возвращает несколько компаний по списку ОГРН (для батчинга)
func (r *CompanyRepository) GetByOGRNs(ctx context.Context, ogrns []string) ([]*model.Company, error) {
	if len(ogrns) == 0 {
		return []*model.Company{}, nil
	}

	companies := make([]*model.Company, 0, len(ogrns))
	for _, ogrn := range ogrns {
		company, err := r.GetByOGRN(ctx, ogrn)
		if err != nil {
			r.logger.Warn("failed to get company in batch",
				zap.String("ogrn", ogrn),
				zap.Error(err),
			)
			continue
		}
		companies = append(companies, company)
	}

	return companies, nil
}

// GetFounders возвращает учредителей компании
func (r *CompanyRepository) GetFounders(ctx context.Context, ogrn string) ([]model.Founder, error) {
	query := `
		SELECT
			founder_name,
			founder_inn,
			founder_ogrn,
			share_nominal_value,
			share_percent
		FROM founders
		WHERE company_ogrn = ?
		ORDER BY share_percent DESC
	`

	rows, err := r.conn.Query(ctx, query, ogrn)
	if err != nil {
		return nil, fmt.Errorf("failed to query founders: %w", err)
	}
	defer rows.Close()

	var founders []model.Founder
	for rows.Next() {
		var founder model.Founder
		var shareAmount, sharePercent sql.NullFloat64

		err := rows.Scan(
			&founder.FullName,
			&founder.INN,
			&founder.OGRN,
			&shareAmount,
			&sharePercent,
		)
		if err != nil {
			r.logger.Warn("failed to scan founder", zap.Error(err))
			continue
		}

		if shareAmount.Valid {
			founder.ShareAmount = shareAmount.Float64
		}
		if sharePercent.Valid {
			founder.SharePercent = sharePercent.Float64
		}

		founders = append(founders, founder)
	}

	return founders, nil
}

// GetActivities возвращает виды деятельности компании
func (r *CompanyRepository) GetActivities(ctx context.Context, ogrn string) (string, []string, error) {
	// Основной ОКВЭД уже получен в GetByOGRN, здесь получаем дополнительные
	query := `
		SELECT DISTINCT okved_code
		FROM companies_okved_additional
		WHERE ogrn = ?
		ORDER BY okved_code
	`

	rows, err := r.conn.Query(ctx, query, ogrn)
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

// GetLicensesCount возвращает количество лицензий компании
func (r *CompanyRepository) GetLicensesCount(ctx context.Context, ogrn string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT license_number)
		FROM licenses
		WHERE entity_type = 'company' AND entity_ogrn = ?
	`

	var count uint64
	err := r.conn.QueryRow(ctx, query, ogrn).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get licenses count: %w", err)
	}

	return int(count), nil
}

// GetBranchesCount возвращает количество филиалов компании
func (r *CompanyRepository) GetBranchesCount(ctx context.Context, ogrn string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT branch_name)
		FROM branches
		WHERE company_ogrn = ?
	`

	var count uint64
	err := r.conn.QueryRow(ctx, query, ogrn).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get branches count: %w", err)
	}

	return int(count), nil
}

// GetPreviousByOGRN возвращает предыдущую версию компании (с максимальной updated_at меньше текущей)
func (r *CompanyRepository) GetPreviousByOGRN(ctx context.Context, ogrn string, beforeDate string) (*model.Company, error) {
	query := `
		SELECT
			ogrn,
			inn,
			kpp,
			full_name,
			short_name,
			region_code,
			status,
			concat(head_last_name, ' ', head_first_name, ' ', head_middle_name) as director_full_name,
			head_inn,
			head_position,
			full_address,
			postal_code,
			region,
			city,
			street,
			house,
			capital_amount,
			capital_currency,
			okved_main_code,
			registration_date,
			extract_date,
			updated_at
		FROM companies
		WHERE ogrn = ? AND updated_at < toDateTime(?)
		ORDER BY updated_at DESC
		LIMIT 1
	`

	row := r.conn.QueryRow(ctx, query, ogrn, beforeDate)

	var company model.Company
	var authorizedCapital sql.NullFloat64
	var capitalCurrency, directorFullName, directorINN, directorPosition, addressFull, postalCode, addressRegion, addressCity, addressStreet, addressHouse sql.NullString

	err := row.Scan(
		&company.OGRN,
		&company.INN,
		&company.KPP,
		&company.FullName,
		&company.ShortName,
		&company.RegionCode,
		&company.Status,
		&directorFullName,
		&directorINN,
		&directorPosition,
		&addressFull,
		&postalCode,
		&addressRegion,
		&addressCity,
		&addressStreet,
		&addressHouse,
		&authorizedCapital,
		&capitalCurrency,
		&company.MainOKVED,
		&company.RegistrationDate,
		&company.ExtractDate,
		&company.LastUpdate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет предыдущей версии - это не ошибка
		}
		r.logger.Error("failed to get previous company version by OGRN",
			zap.String("ogrn", ogrn),
			zap.String("before_date", beforeDate),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get previous company version: %w", err)
	}

	// Преобразование nullable полей
	if directorFullName.Valid {
		company.DirectorFullName = directorFullName.String
	}
	if directorINN.Valid {
		company.DirectorINN = directorINN.String
	}
	if directorPosition.Valid {
		company.DirectorPosition = directorPosition.String
	}
	if addressFull.Valid {
		company.AddressFull = addressFull.String
	}
	if postalCode.Valid {
		company.AddressPostalCode = postalCode.String
	}
	if addressRegion.Valid {
		company.AddressRegion = addressRegion.String
	}
	if addressCity.Valid {
		company.AddressCity = addressCity.String
	}
	if addressStreet.Valid {
		company.AddressStreet = addressStreet.String
	}
	if addressHouse.Valid {
		company.AddressHouse = addressHouse.String
	}
	if authorizedCapital.Valid {
		company.AuthorizedCapital = authorizedCapital.Float64
	}
	if capitalCurrency.Valid {
		company.CapitalCurrency = capitalCurrency.String
	}

	// Загрузка связанных данных (опционально, для упрощения можно пропустить для предыдущей версии)
	founders, err := r.GetFounders(ctx, ogrn)
	if err != nil {
		company.Founders = []model.Founder{}
	} else {
		company.Founders = founders
	}

	_, additionalOKVED, err := r.GetActivities(ctx, ogrn)
	if err != nil {
		company.AdditionalOKVED = []string{}
	} else {
		company.AdditionalOKVED = additionalOKVED
	}

	licensesCount, err := r.GetLicensesCount(ctx, ogrn)
	if err != nil {
		company.LicensesCount = 0
	} else {
		company.LicensesCount = licensesCount
	}

	branchesCount, err := r.GetBranchesCount(ctx, ogrn)
	if err != nil {
		company.BranchesCount = 0
	} else {
		company.BranchesCount = branchesCount
	}

	return &company, nil
}
