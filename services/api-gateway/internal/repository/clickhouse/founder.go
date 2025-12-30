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
	r.logger.Info("GetByCompanyOGRN called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	query := `
		SELECT * FROM egrul.founders FINAL
		WHERE company_ogrn = ?
		ORDER BY share_percent DESC NULLS LAST, founder_name
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query founders failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query founders: %w", err)
	}
	defer rows.Close()

	var founders []*model.Founder
	for rows.Next() {
		var row founderRow
		var versionDate sql.NullTime
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&row.ID,
			&row.CompanyOgrn,
			&row.CompanyInn,
			&row.CompanyName,
			&row.FounderType,
			&row.FounderOgrn,
			&row.FounderInn,
			&row.FounderName,
			&row.FounderLastName,
			&row.FounderFirstName,
			&row.FounderMiddleName,
			&row.FounderCountry,
			&row.FounderCitizenship,
			&row.ShareNominalValue,
			&row.SharePercent,
			&versionDate,
			&createdAt,
			&updatedAt,
		); err != nil {
			r.logger.Error("scan founder row failed", zap.Error(err))
			return nil, fmt.Errorf("scan founder row: %w", err)
		}
		founders = append(founders, row.toModel())
	}

	r.logger.Info("GetByCompanyOGRN completed", zap.String("ogrn", ogrn), zap.Int("count", len(founders)))
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

// GetCompaniesWithCommonFounders получает компании с общими учредителями (физическими лицами)
func (r *FounderRepository) GetCompaniesWithCommonFounders(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	r.logger.Info("GetCompaniesWithCommonFounders called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	// Упрощенный запрос без алиасов для ClickHouse
	query := `
		SELECT DISTINCT company_ogrn
		FROM egrul.founders FINAL
		WHERE founder_inn IN (
			SELECT DISTINCT founder_inn
			FROM egrul.founders FINAL
			WHERE company_ogrn = ?
			  AND founder_type = 'person'
			  AND founder_inn != ''
		)
		  AND company_ogrn != ?
		  AND founder_type = 'person'
		  AND founder_inn != ''
		ORDER BY company_ogrn
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query companies with common founders failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query companies with common founders: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var relatedOgrn string
		if err := rows.Scan(&relatedOgrn); err != nil {
			r.logger.Error("scan related ogrn failed", zap.Error(err))
			return nil, fmt.Errorf("scan related ogrn: %w", err)
		}
		ogrns = append(ogrns, relatedOgrn)
	}

	r.logger.Info("GetCompaniesWithCommonFounders completed", zap.String("ogrn", ogrn), zap.Int("count", len(ogrns)))
	return ogrns, nil
}

// GetFounderCompanies получает компании, где учредители основной компании являются учредителями
func (r *FounderRepository) GetFounderCompanies(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	r.logger.Info("GetFounderCompanies called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	query := `
		WITH target_legal_founders AS (
			SELECT DISTINCT founder_ogrn
			FROM egrul.founders FINAL
			WHERE company_ogrn = ?
			  AND founder_type IN ('russian_company', 'foreign_company')
			  AND founder_ogrn != ''
		)
		SELECT DISTINCT tlf.founder_ogrn
		FROM target_legal_founders tlf
		ORDER BY tlf.founder_ogrn
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query founder companies failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query founder companies: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var founderOgrn string
		if err := rows.Scan(&founderOgrn); err != nil {
			r.logger.Error("scan founder ogrn failed", zap.Error(err))
			return nil, fmt.Errorf("scan founder ogrn: %w", err)
		}
		ogrns = append(ogrns, founderOgrn)
	}

	r.logger.Info("GetFounderCompanies completed", zap.String("ogrn", ogrn), zap.Int("count", len(ogrns)))
	return ogrns, nil
}

// GetCommonFoundersDetails получает детальную информацию об общих учредителях-физлицах между двумя компаниями
func (r *FounderRepository) GetCommonFoundersDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Founder, error) {
	r.logger.Info("GetCommonFoundersDetails called", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2))
	
	query := `SELECT DISTINCT f1.founder_inn, f1.founder_name, f1.founder_last_name, f1.founder_first_name, f1.founder_middle_name, f1.founder_citizenship, f1.share_percent as share_percent_1, f2.share_percent as share_percent_2 FROM egrul.founders f1 FINAL INNER JOIN egrul.founders f2 FINAL ON f1.founder_inn = f2.founder_inn WHERE f1.company_ogrn = ? AND f2.company_ogrn = ? AND f1.founder_type = 'person' AND f2.founder_type = 'person' AND f1.founder_inn != '' ORDER BY f1.founder_name`

	rows, err := r.client.conn.Query(ctx, query, ogrn1, ogrn2)
	if err != nil {
		r.logger.Error("query common founders details failed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Error(err))
		return nil, fmt.Errorf("query common founders details: %w", err)
	}
	defer rows.Close()

	var founders []*model.Founder
	for rows.Next() {
		var founderInn, founderName string
		var founderLastName, founderFirstName, founderMiddleName, founderCitizenship sql.NullString
		var sharePercent1, sharePercent2 sql.NullFloat64
		
		if err := rows.Scan(
			&founderInn,
			&founderName,
			&founderLastName,
			&founderFirstName,
			&founderMiddleName,
			&founderCitizenship,
			&sharePercent1,
			&sharePercent2,
		); err != nil {
			r.logger.Error("scan common founder failed", zap.Error(err))
			return nil, fmt.Errorf("scan common founder: %w", err)
		}

		founder := &model.Founder{
			Type: model.FounderTypePerson,
			Inn:  &founderInn,
			Name: founderName,
		}

		if founderLastName.Valid {
			founder.LastName = &founderLastName.String
		}
		if founderFirstName.Valid {
			founder.FirstName = &founderFirstName.String
		}
		if founderMiddleName.Valid {
			founder.MiddleName = &founderMiddleName.String
		}
		if founderCitizenship.Valid {
			founder.Citizenship = &founderCitizenship.String
		}
		// Используем долю из первой компании как основную
		if sharePercent1.Valid {
			founder.SharePercent = &sharePercent1.Float64
		}

		founders = append(founders, founder)
	}

	r.logger.Info("GetCommonFoundersDetails completed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Int("count", len(founders)))
	return founders, nil
}

// GetCompaniesWithCommonDirectors получает компании с общими руководителями (физическими лицами) по ИНН
func (r *FounderRepository) GetCompaniesWithCommonDirectors(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	r.logger.Info("GetCompaniesWithCommonDirectors called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	query := `SELECT DISTINCT c2.ogrn FROM egrul.companies c1 FINAL INNER JOIN egrul.companies c2 FINAL ON c1.head_inn = c2.head_inn WHERE c1.ogrn = ? AND c2.ogrn != ? AND c1.head_inn != '' AND c2.head_inn != '' ORDER BY c2.ogrn LIMIT ? OFFSET ?`

	rows, err := r.client.conn.Query(ctx, query, ogrn, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query companies with common directors failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query companies with common directors: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var relatedOgrn string
		if err := rows.Scan(&relatedOgrn); err != nil {
			r.logger.Error("scan related ogrn failed", zap.Error(err))
			return nil, fmt.Errorf("scan related ogrn: %w", err)
		}
		ogrns = append(ogrns, relatedOgrn)
	}

	r.logger.Info("GetCompaniesWithCommonDirectors completed", zap.String("ogrn", ogrn), zap.Int("count", len(ogrns)))
	return ogrns, nil
}

// GetCommonDirectorsDetails получает детальную информацию об общих руководителях между двумя компаниями (только по ИНН)
func (r *FounderRepository) GetCommonDirectorsDetails(ctx context.Context, ogrn1, ogrn2 string) ([]*model.Person, error) {
	r.logger.Info("GetCommonDirectorsDetails called", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2))
	
	query := `SELECT DISTINCT c1.head_inn, c1.head_last_name, c1.head_first_name, c1.head_middle_name, c1.head_position FROM egrul.companies c1 FINAL INNER JOIN egrul.companies c2 FINAL ON c1.head_inn = c2.head_inn WHERE c1.ogrn = ? AND c2.ogrn = ? AND c1.head_inn != '' AND c2.head_inn != '' ORDER BY c1.head_last_name, c1.head_first_name`

	rows, err := r.client.conn.Query(ctx, query, ogrn1, ogrn2)
	if err != nil {
		r.logger.Error("query common directors details failed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Error(err))
		return nil, fmt.Errorf("query common directors details: %w", err)
	}
	defer rows.Close()

	var directors []*model.Person
	for rows.Next() {
		var headInn sql.NullString
		var headLastName, headFirstName, headMiddleName, headPosition sql.NullString
		
		if err := rows.Scan(
			&headInn,
			&headLastName,
			&headFirstName,
			&headMiddleName,
			&headPosition,
		); err != nil {
			r.logger.Error("scan common director failed", zap.Error(err))
			return nil, fmt.Errorf("scan common director: %w", err)
		}

		director := &model.Person{}

		if headLastName.Valid {
			director.LastName = headLastName.String
		}
		if headFirstName.Valid {
			director.FirstName = headFirstName.String
		}
		if headMiddleName.Valid {
			middleName := headMiddleName.String
			director.MiddleName = &middleName
		}
		if headInn.Valid {
			inn := headInn.String
			director.Inn = &inn
		}
		if headPosition.Valid {
			position := headPosition.String
			director.Position = &position
		}

		directors = append(directors, director)
	}

	r.logger.Info("GetCommonDirectorsDetails completed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Int("count", len(directors)))
	return directors, nil
}

// GetCompaniesWhereFounderIsDirector получает компании, где учредители основной компании являются руководителями
func (r *FounderRepository) GetCompaniesWhereFounderIsDirector(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	r.logger.Info("GetCompaniesWhereFounderIsDirector called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	query := `SELECT DISTINCT c.ogrn FROM egrul.founders f FINAL INNER JOIN egrul.companies c FINAL ON f.founder_inn = c.head_inn WHERE f.company_ogrn = ? AND f.founder_type = 'person' AND f.founder_inn != '' AND c.head_inn != '' AND c.ogrn != ? ORDER BY c.ogrn LIMIT ? OFFSET ?`

	rows, err := r.client.conn.Query(ctx, query, ogrn, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query companies where founder is director failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query companies where founder is director: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var relatedOgrn string
		if err := rows.Scan(&relatedOgrn); err != nil {
			r.logger.Error("scan related ogrn failed", zap.Error(err))
			return nil, fmt.Errorf("scan related ogrn: %w", err)
		}
		ogrns = append(ogrns, relatedOgrn)
	}

	r.logger.Info("GetCompaniesWhereFounderIsDirector completed", zap.String("ogrn", ogrn), zap.Int("count", len(ogrns)))
	return ogrns, nil
}

// GetCompaniesWhereDirectorIsFounder получает компании, где руководитель основной компании является учредителем
func (r *FounderRepository) GetCompaniesWhereDirectorIsFounder(ctx context.Context, ogrn string, limit, offset int) ([]string, error) {
	r.logger.Info("GetCompaniesWhereDirectorIsFounder called", zap.String("ogrn", ogrn), zap.Int("limit", limit), zap.Int("offset", offset))
	
	query := `SELECT DISTINCT f.company_ogrn FROM egrul.companies c FINAL INNER JOIN egrul.founders f FINAL ON c.head_inn = f.founder_inn WHERE c.ogrn = ? AND c.head_inn != '' AND f.founder_type = 'person' AND f.founder_inn != '' AND f.company_ogrn != ? ORDER BY f.company_ogrn LIMIT ? OFFSET ?`

	rows, err := r.client.conn.Query(ctx, query, ogrn, ogrn, limit, offset)
	if err != nil {
		r.logger.Error("query companies where director is founder failed", zap.String("ogrn", ogrn), zap.Error(err))
		return nil, fmt.Errorf("query companies where director is founder: %w", err)
	}
	defer rows.Close()

	var ogrns []string
	for rows.Next() {
		var relatedOgrn string
		if err := rows.Scan(&relatedOgrn); err != nil {
			r.logger.Error("scan related ogrn failed", zap.Error(err))
			return nil, fmt.Errorf("scan related ogrn: %w", err)
		}
		ogrns = append(ogrns, relatedOgrn)
	}

	r.logger.Info("GetCompaniesWhereDirectorIsFounder completed", zap.String("ogrn", ogrn), zap.Int("count", len(ogrns)))
	return ogrns, nil
}

// GetCrossPersonDetails получает детальную информацию о перекрестных связях через физлицо
func (r *FounderRepository) GetCrossPersonDetails(ctx context.Context, ogrn1, ogrn2 string, crossType string) ([]*model.Person, error) {
	r.logger.Info("GetCrossPersonDetails called", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.String("crossType", crossType))
	
	var query string
	if crossType == "founder_to_director" {
		// Учредители ogrn1, которые являются руководителями ogrn2
		query = `SELECT DISTINCT f.founder_inn, f.founder_last_name, f.founder_first_name, f.founder_middle_name, c.head_position FROM egrul.founders f FINAL INNER JOIN egrul.companies c FINAL ON f.founder_inn = c.head_inn WHERE f.company_ogrn = ? AND c.ogrn = ? AND f.founder_type = 'person' AND f.founder_inn != '' AND c.head_inn != ''`
	} else {
		// Руководитель ogrn1, который является учредителем ogrn2
		query = `SELECT DISTINCT c.head_inn, c.head_last_name, c.head_first_name, c.head_middle_name, c.head_position FROM egrul.companies c FINAL INNER JOIN egrul.founders f FINAL ON c.head_inn = f.founder_inn WHERE c.ogrn = ? AND f.company_ogrn = ? AND c.head_inn != '' AND f.founder_type = 'person' AND f.founder_inn != ''`
	}

	rows, err := r.client.conn.Query(ctx, query, ogrn1, ogrn2)
	if err != nil {
		r.logger.Error("query cross person details failed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Error(err))
		return nil, fmt.Errorf("query cross person details: %w", err)
	}
	defer rows.Close()

	var persons []*model.Person
	for rows.Next() {
		var inn sql.NullString
		var lastName, firstName, middleName, position sql.NullString
		
		if err := rows.Scan(&inn, &lastName, &firstName, &middleName, &position); err != nil {
			r.logger.Error("scan cross person failed", zap.Error(err))
			return nil, fmt.Errorf("scan cross person: %w", err)
		}

		person := &model.Person{}

		if lastName.Valid {
			person.LastName = lastName.String
		}
		if firstName.Valid {
			person.FirstName = firstName.String
		}
		if middleName.Valid {
			middleName := middleName.String
			person.MiddleName = &middleName
		}
		if inn.Valid {
			innStr := inn.String
			person.Inn = &innStr
		}
		if position.Valid {
			positionStr := position.String
			person.Position = &positionStr
		}

		persons = append(persons, person)
	}

	r.logger.Info("GetCrossPersonDetails completed", zap.String("ogrn1", ogrn1), zap.String("ogrn2", ogrn2), zap.Int("count", len(persons)))
	return persons, nil
}

