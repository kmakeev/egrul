package clickhouse

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// CompanyRepository репозиторий для работы с компаниями
type CompanyRepository struct {
	client *Client
	logger *zap.Logger
}

// NewCompanyRepository создает новый репозиторий компаний
func NewCompanyRepository(client *Client, logger *zap.Logger) *CompanyRepository {
	return &CompanyRepository{
		client: client,
		logger: logger.Named("company_repo"),
	}
}

// companyRow структура для сканирования результатов запроса
type companyRow struct {
	Ogrn                 string          `ch:"ogrn"`
	OgrnDate             sql.NullTime    `ch:"ogrn_date"`
	Inn                  string          `ch:"inn"`
	Kpp                  sql.NullString  `ch:"kpp"`
	FullName             string          `ch:"full_name"`
	ShortName            sql.NullString  `ch:"short_name"`
	BrandName            sql.NullString  `ch:"brand_name"`
	OpfCode              sql.NullString  `ch:"opf_code"`
	OpfName              sql.NullString  `ch:"opf_name"`
	Status               string          `ch:"status"`
	StatusCode           sql.NullString  `ch:"status_code"`
	TerminationMethod    sql.NullString  `ch:"termination_method"`
	RegistrationDate     sql.NullTime    `ch:"registration_date"`
	TerminationDate      sql.NullTime    `ch:"termination_date"`
	ExtractDate          sql.NullTime    `ch:"extract_date"`
	PostalCode           sql.NullString  `ch:"postal_code"`
	RegionCode           sql.NullString  `ch:"region_code"`
	Region               sql.NullString  `ch:"region"`
	District             sql.NullString  `ch:"district"`
	City                 sql.NullString  `ch:"city"`
	Locality             sql.NullString  `ch:"locality"`
	Street               sql.NullString  `ch:"street"`
	House                sql.NullString  `ch:"house"`
	Building             sql.NullString  `ch:"building"`
	Flat                 sql.NullString  `ch:"flat"`
	FullAddress          sql.NullString  `ch:"full_address"`
	FiasID               sql.NullString  `ch:"fias_id"`
	Email                sql.NullString  `ch:"email"`
	CapitalAmount        sql.NullFloat64 `ch:"capital_amount"`
	CapitalCurrency      sql.NullString  `ch:"capital_currency"`
	HeadLastName         sql.NullString  `ch:"head_last_name"`
	HeadFirstName        sql.NullString  `ch:"head_first_name"`
	HeadMiddleName       sql.NullString  `ch:"head_middle_name"`
	HeadInn              sql.NullString  `ch:"head_inn"`
	HeadPosition         sql.NullString  `ch:"head_position"`
	HeadPositionCode     sql.NullString  `ch:"head_position_code"`
	OkvedMainCode        sql.NullString  `ch:"okved_main_code"`
	OkvedMainName        sql.NullString  `ch:"okved_main_name"`
	OkvedAdditional      []string        `ch:"okved_additional"`
	OkvedAdditionalNames []string        `ch:"okved_additional_names"`
	AdditionalActivities sql.NullString  `ch:"additional_activities"`
	RegAuthorityCode     sql.NullString  `ch:"reg_authority_code"`
	RegAuthorityName     sql.NullString  `ch:"reg_authority_name"`
	TaxAuthorityCode     sql.NullString  `ch:"tax_authority_code"`
	TaxAuthorityName     sql.NullString  `ch:"tax_authority_name"`
	PfrRegNumber         sql.NullString  `ch:"pfr_reg_number"`
	FssRegNumber         sql.NullString  `ch:"fss_reg_number"`
	FoundersCount        uint16          `ch:"founders_count"`
	LicensesCount        uint16          `ch:"licenses_count"`
	BranchesCount        uint16          `ch:"branches_count"`
	IsBankrupt           uint8           `ch:"is_bankrupt"`
	BankruptcyStage      sql.NullString  `ch:"bankruptcy_stage"`
	IsLiquidating        uint8           `ch:"is_liquidating"`
	IsReorganizing       uint8           `ch:"is_reorganizing"`
	LastGrn              sql.NullString  `ch:"last_grn"`
	LastGrnDate          sql.NullTime    `ch:"last_grn_date"`
	DocumentID           sql.NullString  `ch:"document_id"`
	SourceFile           sql.NullString  `ch:"source_file"`
	VersionDate          time.Time       `ch:"version_date"`
	CreatedAt            time.Time       `ch:"created_at"`
	UpdatedAt            time.Time       `ch:"updated_at"`
}

func (r *companyRow) toModel() *model.Company {
	company := &model.Company{
		Ogrn:            r.Ogrn,
		Inn:             r.Inn,
		FullName:        r.FullName,
		Status:          model.ParseEntityStatus(r.Status),
		FoundersCount:   int(r.FoundersCount),
		LicensesCount:   int(r.LicensesCount),
		BranchesCount:   int(r.BranchesCount),
		IsBankrupt:      r.IsBankrupt == 1,
		IsLiquidating:   r.IsLiquidating == 1,
		IsReorganizing:  r.IsReorganizing == 1,
		VersionDate:     model.Date{Time: r.VersionDate},
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		Activities:      make([]*model.Activity, 0),
	}

	// Nullable fields
	if r.OgrnDate.Valid {
		company.OgrnDate = &model.Date{Time: r.OgrnDate.Time}
	}
	if r.Kpp.Valid {
		company.Kpp = &r.Kpp.String
	}
	if r.ShortName.Valid {
		company.ShortName = &r.ShortName.String
	}
	if r.BrandName.Valid {
		company.BrandName = &r.BrandName.String
	}
	if r.StatusCode.Valid {
		company.StatusCode = &r.StatusCode.String
	}
	if r.TerminationMethod.Valid {
		company.TerminationMethod = &r.TerminationMethod.String
	}
	if r.RegistrationDate.Valid {
		company.RegistrationDate = &model.Date{Time: r.RegistrationDate.Time}
	}
	if r.TerminationDate.Valid {
		company.TerminationDate = &model.Date{Time: r.TerminationDate.Time}
	}
	if r.ExtractDate.Valid {
		company.ExtractDate = &model.Date{Time: r.ExtractDate.Time}
	}
	if r.Email.Valid {
		company.Email = &r.Email.String
	}
	if r.LastGrn.Valid {
		company.LastGrn = &r.LastGrn.String
	}
	if r.LastGrnDate.Valid {
		company.LastGrnDate = &model.Date{Time: r.LastGrnDate.Time}
	}
	if r.SourceFile.Valid {
		company.SourceFile = &r.SourceFile.String
	}
	if r.BankruptcyStage.Valid {
		company.BankruptcyStage = &r.BankruptcyStage.String
	}
	if r.PfrRegNumber.Valid {
		company.PfrRegNumber = &r.PfrRegNumber.String
	}
	if r.FssRegNumber.Valid {
		company.FssRegNumber = &r.FssRegNumber.String
	}

	// Legal form
	if r.OpfCode.Valid || r.OpfName.Valid {
		company.LegalForm = &model.LegalForm{}
		if r.OpfCode.Valid {
			company.LegalForm.Code = &r.OpfCode.String
		}
		if r.OpfName.Valid {
			company.LegalForm.Name = &r.OpfName.String
		}
	}

	// Address
	company.Address = &model.Address{}
	if r.PostalCode.Valid {
		company.Address.PostalCode = &r.PostalCode.String
	}
	if r.RegionCode.Valid {
		company.Address.RegionCode = &r.RegionCode.String
	}
	if r.Region.Valid {
		company.Address.Region = &r.Region.String
	}
	if r.District.Valid {
		company.Address.District = &r.District.String
	}
	if r.City.Valid {
		company.Address.City = &r.City.String
	}
	if r.Locality.Valid {
		company.Address.Locality = &r.Locality.String
	}
	if r.Street.Valid {
		company.Address.Street = &r.Street.String
	}
	if r.House.Valid {
		company.Address.House = &r.House.String
	}
	if r.Building.Valid {
		company.Address.Building = &r.Building.String
	}
	if r.Flat.Valid {
		company.Address.Flat = &r.Flat.String
	}
	if r.FullAddress.Valid {
		company.Address.FullAddress = &r.FullAddress.String
	}
	if r.FiasID.Valid {
		company.Address.FiasID = &r.FiasID.String
	}

	// Capital
	if r.CapitalAmount.Valid && r.CapitalAmount.Float64 > 0 {
		currency := "RUB"
		if r.CapitalCurrency.Valid {
			currency = r.CapitalCurrency.String
		}
		company.Capital = &model.Money{
			Amount:   r.CapitalAmount.Float64,
			Currency: currency,
		}
	}

	// Director
	if r.HeadLastName.Valid && r.HeadFirstName.Valid {
		company.Director = &model.Person{
			LastName:  r.HeadLastName.String,
			FirstName: r.HeadFirstName.String,
		}
		if r.HeadMiddleName.Valid {
			company.Director.MiddleName = &r.HeadMiddleName.String
		}
		if r.HeadInn.Valid {
			company.Director.Inn = &r.HeadInn.String
		}
		if r.HeadPosition.Valid {
			company.Director.Position = &r.HeadPosition.String
		}
		if r.HeadPositionCode.Valid {
			company.Director.PositionCode = &r.HeadPositionCode.String
		}
	}

	// Main activity
	if r.OkvedMainCode.Valid {
		company.MainActivity = &model.Activity{
			Code:   r.OkvedMainCode.String,
			IsMain: true,
		}
		if r.OkvedMainName.Valid {
			company.MainActivity.Name = &r.OkvedMainName.String
		}
		company.Activities = append(company.Activities, company.MainActivity)
	}

	// Additional activities
	for i, code := range r.OkvedAdditional {
		activity := &model.Activity{
			Code:   code,
			IsMain: false,
		}
		if i < len(r.OkvedAdditionalNames) {
			activity.Name = &r.OkvedAdditionalNames[i]
		}
		company.Activities = append(company.Activities, activity)
	}

	// Authorities
	if r.RegAuthorityCode.Valid || r.RegAuthorityName.Valid {
		company.RegAuthority = &model.Authority{}
		if r.RegAuthorityCode.Valid {
			company.RegAuthority.Code = &r.RegAuthorityCode.String
		}
		if r.RegAuthorityName.Valid {
			company.RegAuthority.Name = &r.RegAuthorityName.String
		}
	}
	if r.TaxAuthorityCode.Valid || r.TaxAuthorityName.Valid {
		company.TaxAuthority = &model.Authority{}
		if r.TaxAuthorityCode.Valid {
			company.TaxAuthority.Code = &r.TaxAuthorityCode.String
		}
		if r.TaxAuthorityName.Valid {
			company.TaxAuthority.Name = &r.TaxAuthorityName.String
		}
	}

	return company
}

// GetByOGRN получает компанию по ОГРН
func (r *CompanyRepository) GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error) {
	query := `
		SELECT * FROM egrul.companies FINAL
		WHERE ogrn = ?
		LIMIT 1
	`

	rows, err := r.client.conn.Query(ctx, query, ogrn)
	if err != nil {
		return nil, fmt.Errorf("query company by ogrn: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var row companyRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan company row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for company", zap.String("ogrn", ogrn), zap.Error(err))
		}
		return row.toModel(), nil
	}

	return nil, nil
}

// GetByINN получает компанию по ИНН
func (r *CompanyRepository) GetByINN(ctx context.Context, inn string) (*model.Company, error) {
	query := `
		SELECT * FROM egrul.companies FINAL
		WHERE inn = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	rows, err := r.client.conn.Query(ctx, query, inn)
	if err != nil {
		return nil, fmt.Errorf("query company by inn: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var row companyRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan company row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for company by inn", zap.String("inn", inn), zap.Error(err))
		}
		return row.toModel(), nil
	}

	return nil, nil
}

// List возвращает список компаний с фильтрацией и пагинацией
func (r *CompanyRepository) List(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) ([]*model.Company, int, error) {
	whereClause, args := r.buildWhereClause(filter)
	orderClause := r.buildOrderClause(sort)

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT count() FROM egrul.companies FINAL
		%s
	`, whereClause)

	var totalCount uint64
	countRow := r.client.conn.QueryRow(ctx, countQuery, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count companies: %w", err)
	}

	// Data query
	// Используем прямые значения для LIMIT/OFFSET, так как ClickHouse может не поддерживать параметризацию для них
	dataQuery := fmt.Sprintf(`
		SELECT * FROM egrul.companies FINAL
		%s
		%s
		LIMIT %d OFFSET %d
	`, whereClause, orderClause, limit, offset)

	rows, err := r.client.conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query companies: %w", err)
	}
	defer rows.Close()

	var companies []*model.Company
	for rows.Next() {
		var row companyRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, 0, fmt.Errorf("scan company row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for company in list", zap.String("ogrn", row.Ogrn), zap.Error(err))
		}
		companies = append(companies, row.toModel())
	}

	return companies, int(totalCount), nil
}

// Search выполняет текстовый поиск компаний
func (r *CompanyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Company, error) {
	searchQuery := `
		SELECT * FROM egrul.companies FINAL
		WHERE 
			full_name ILIKE ?
			OR short_name ILIKE ?
			OR inn LIKE ?
			OR ogrn LIKE ?
		ORDER BY 
			CASE 
				WHEN ogrn = ? THEN 1
				WHEN inn = ? THEN 2
				WHEN full_name ILIKE ? THEN 3
				ELSE 4
			END,
			full_name
		LIMIT ? OFFSET ?
	`

	pattern := "%" + query + "%"
	exactPattern := query + "%"

	rows, err := r.client.conn.Query(ctx, searchQuery,
		pattern, pattern, exactPattern, exactPattern,
		query, query, exactPattern,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("search companies: %w", err)
	}
	defer rows.Close()

	var companies []*model.Company
	for rows.Next() {
		var row companyRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan company row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for company in search", zap.String("ogrn", row.Ogrn), zap.Error(err))
		}
		companies = append(companies, row.toModel())
	}

	return companies, nil
}

func (r *CompanyRepository) buildWhereClause(filter *model.CompanyFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if filter.Ogrn != nil && *filter.Ogrn != "" {
		conditions = append(conditions, "ogrn = ?")
		args = append(args, *filter.Ogrn)
	}
	if filter.Inn != nil && *filter.Inn != "" {
		conditions = append(conditions, "inn = ?")
		args = append(args, *filter.Inn)
	}
	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, "(full_name ILIKE ? OR short_name ILIKE ?)")
		pattern := "%" + *filter.Name + "%"
		args = append(args, pattern, pattern)
	}
	if filter.RegionCode != nil && *filter.RegionCode != "" {
		conditions = append(conditions, "region_code = ?")
		args = append(args, *filter.RegionCode)
	}
	if filter.Region != nil && *filter.Region != "" {
		conditions = append(conditions, "region ILIKE ?")
		args = append(args, "%"+*filter.Region+"%")
	}
	if filter.Okved != nil && *filter.Okved != "" {
		conditions = append(conditions, "(okved_main_code = ? OR has(okved_additional, ?))")
		args = append(args, *filter.Okved, *filter.Okved)
	}
	if filter.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, statusToDBValue(*filter.Status))
	}
	if len(filter.StatusIn) > 0 {
		placeholders := make([]string, len(filter.StatusIn))
		for i, s := range filter.StatusIn {
			placeholders[i] = "?"
			args = append(args, statusToDBValue(s))
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}
	if filter.RegisteredAfter != nil {
		conditions = append(conditions, "registration_date >= ?")
		args = append(args, filter.RegisteredAfter.Time)
	}
	if filter.RegisteredBefore != nil {
		conditions = append(conditions, "registration_date <= ?")
		args = append(args, filter.RegisteredBefore.Time)
	}
	if filter.TerminatedAfter != nil {
		conditions = append(conditions, "termination_date >= ?")
		args = append(args, filter.TerminatedAfter.Time)
	}
	if filter.TerminatedBefore != nil {
		conditions = append(conditions, "termination_date <= ?")
		args = append(args, filter.TerminatedBefore.Time)
	}
	if filter.CapitalMin != nil {
		conditions = append(conditions, "capital_amount >= ?")
		args = append(args, *filter.CapitalMin)
	}
	if filter.CapitalMax != nil {
		conditions = append(conditions, "capital_amount <= ?")
		args = append(args, *filter.CapitalMax)
	}
	if filter.IsBankrupt != nil {
		if *filter.IsBankrupt {
			conditions = append(conditions, "is_bankrupt = 1")
		} else {
			conditions = append(conditions, "is_bankrupt = 0")
		}
	}
	if filter.IsLiquidating != nil {
		if *filter.IsLiquidating {
			conditions = append(conditions, "is_liquidating = 1")
		} else {
			conditions = append(conditions, "is_liquidating = 0")
		}
	}
	if filter.HasDirector != nil {
		if *filter.HasDirector {
			conditions = append(conditions, "head_last_name IS NOT NULL AND head_last_name != ''")
		} else {
			conditions = append(conditions, "(head_last_name IS NULL OR head_last_name = '')")
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *CompanyRepository) buildOrderClause(sort *model.CompanySort) string {
	if sort == nil {
		return "ORDER BY updated_at DESC"
	}

	var field string
	switch sort.Field {
	case model.CompanySortFieldOgrn:
		field = "ogrn"
	case model.CompanySortFieldInn:
		field = "inn"
	case model.CompanySortFieldFullName:
		field = "full_name"
	case model.CompanySortFieldRegistrationDate:
		field = "registration_date"
	case model.CompanySortFieldCapitalAmount:
		field = "capital_amount"
	case model.CompanySortFieldUpdatedAt:
		field = "updated_at"
	default:
		field = "updated_at"
	}

	order := "ASC"
	if sort.Order == model.SortOrderDesc {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", field, order)
}

func statusToDBValue(status model.EntityStatus) string {
	switch status {
	case model.EntityStatusActive:
		return "active"
	case model.EntityStatusLiquidated:
		return "liquidated"
	case model.EntityStatusLiquidating:
		return "liquidating"
	case model.EntityStatusReorganizing:
		return "reorganizing"
	case model.EntityStatusBankrupt:
		return "bankrupt"
	default:
		return "unknown"
	}
}

// loadAdditionalActivities загружает дополнительные ОКВЭД компании из таблицы companies_okved_additional
func (r *CompanyRepository) loadAdditionalActivities(ctx context.Context, row *companyRow) error {
	const q = `
		SELECT okved_code, okved_name
		FROM egrul.companies_okved_additional
		WHERE ogrn = ?
	`

	rows, err := r.client.conn.Query(ctx, q, row.Ogrn)
	if err != nil {
		return fmt.Errorf("query companies_okved_additional: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var name sql.NullString
		if err := rows.Scan(&code, &name); err != nil {
			return fmt.Errorf("scan companies_okved_additional row: %w", err)
		}
		row.OkvedAdditional = append(row.OkvedAdditional, code)
		if name.Valid {
			row.OkvedAdditionalNames = append(row.OkvedAdditionalNames, name.String)
		} else {
			row.OkvedAdditionalNames = append(row.OkvedAdditionalNames, "")
		}
	}
	return nil
}

// EncodeCursor кодирует курсор для пагинации
func EncodeCursor(ogrn string) string {
	return base64.StdEncoding.EncodeToString([]byte(ogrn))
}

// DecodeCursor декодирует курсор
func DecodeCursor(cursor string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

