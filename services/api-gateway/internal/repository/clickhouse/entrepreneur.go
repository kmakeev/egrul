package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// EntrepreneurRepository репозиторий для работы с ИП
type EntrepreneurRepository struct {
	client *Client
	logger *zap.Logger
}

// NewEntrepreneurRepository создает новый репозиторий ИП
func NewEntrepreneurRepository(client *Client, logger *zap.Logger) *EntrepreneurRepository {
	return &EntrepreneurRepository{
		client: client,
		logger: logger.Named("entrepreneur_repo"),
	}
}

// entrepreneurRow структура для сканирования результатов запроса
type entrepreneurRow struct {
	Ogrnip                 string          `ch:"ogrnip"`
	OgrnipDate             sql.NullTime    `ch:"ogrnip_date"`
	Inn                    string          `ch:"inn"`
	LastName               string          `ch:"last_name"`
	FirstName              string          `ch:"first_name"`
	MiddleName             sql.NullString  `ch:"middle_name"`
	Gender                 string          `ch:"gender"`
	CitizenshipType        string          `ch:"citizenship_type"`
	CitizenshipCountryCode sql.NullString  `ch:"citizenship_country_code"`
	CitizenshipCountryName sql.NullString  `ch:"citizenship_country_name"`
	Status                 string          `ch:"status"`
	StatusCode             sql.NullString  `ch:"status_code"`
	TerminationMethod      sql.NullString  `ch:"termination_method"`
	RegistrationDate       sql.NullTime    `ch:"registration_date"`
	TerminationDate        sql.NullTime    `ch:"termination_date"`
	ExtractDate            sql.NullTime    `ch:"extract_date"`
	PostalCode             sql.NullString  `ch:"postal_code"`
	RegionCode             sql.NullString  `ch:"region_code"`
	Region                 sql.NullString  `ch:"region"`
	District               sql.NullString  `ch:"district"`
	City                   sql.NullString  `ch:"city"`
	Locality               sql.NullString  `ch:"locality"`
	FullAddress            sql.NullString  `ch:"full_address"`
	FiasID                 sql.NullString  `ch:"fias_id"`
	Email                  sql.NullString  `ch:"email"`
	OkvedMainCode          sql.NullString  `ch:"okved_main_code"`
	OkvedMainName          sql.NullString  `ch:"okved_main_name"`
	OkvedAdditional        []string        `ch:"okved_additional"`
	OkvedAdditionalNames   []string        `ch:"okved_additional_names"`
	AdditionalActivities   sql.NullString  `ch:"additional_activities"`
	RegAuthorityCode       sql.NullString  `ch:"reg_authority_code"`
	RegAuthorityName       sql.NullString  `ch:"reg_authority_name"`
	TaxAuthorityCode       sql.NullString  `ch:"tax_authority_code"`
	TaxAuthorityName       sql.NullString  `ch:"tax_authority_name"`
	PfrRegNumber           sql.NullString  `ch:"pfr_reg_number"`
	FssRegNumber           sql.NullString  `ch:"fss_reg_number"`
	LicensesCount          uint16          `ch:"licenses_count"`
	IsBankrupt             uint8           `ch:"is_bankrupt"`
	BankruptcyDate         sql.NullTime    `ch:"bankruptcy_date"`
	BankruptcyCaseNumber   sql.NullString  `ch:"bankruptcy_case_number"`
	LastGrn                sql.NullString  `ch:"last_grn"`
	LastGrnDate            sql.NullTime    `ch:"last_grn_date"`
	DocumentID             sql.NullString  `ch:"document_id"`
	SourceFile             sql.NullString  `ch:"source_file"`
	VersionDate            time.Time       `ch:"version_date"`
	CreatedAt              time.Time       `ch:"created_at"`
	UpdatedAt              time.Time       `ch:"updated_at"`
}

func (r *entrepreneurRow) toModel() *model.Entrepreneur {
	entrepreneur := &model.Entrepreneur{
		Ogrnip:        r.Ogrnip,
		Inn:           r.Inn,
		LastName:      r.LastName,
		FirstName:     r.FirstName,
		Status:        model.ParseEntityStatus(r.Status),
		LicensesCount: int(r.LicensesCount),
		IsBankrupt:    r.IsBankrupt == 1,
		VersionDate:   model.Date{Time: r.VersionDate},
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Activities:    make([]*model.Activity, 0),
	}

	// Nullable fields
	if r.OgrnipDate.Valid {
		entrepreneur.OgrnipDate = &model.Date{Time: r.OgrnipDate.Time}
	}
	if r.MiddleName.Valid {
		entrepreneur.MiddleName = &r.MiddleName.String
	}
	if r.Gender != "" {
		entrepreneur.Gender = &r.Gender
	}
	if r.CitizenshipType != "" {
		entrepreneur.CitizenshipType = &r.CitizenshipType
	}
	if r.CitizenshipCountryCode.Valid {
		entrepreneur.CitizenshipCountryCode = &r.CitizenshipCountryCode.String
	}
	if r.CitizenshipCountryName.Valid {
		entrepreneur.CitizenshipCountryName = &r.CitizenshipCountryName.String
	}
	if r.StatusCode.Valid {
		entrepreneur.StatusCode = &r.StatusCode.String
	}
	if r.TerminationMethod.Valid {
		entrepreneur.TerminationMethod = &r.TerminationMethod.String
	}
	if r.RegistrationDate.Valid {
		entrepreneur.RegistrationDate = &model.Date{Time: r.RegistrationDate.Time}
	}
	if r.TerminationDate.Valid {
		entrepreneur.TerminationDate = &model.Date{Time: r.TerminationDate.Time}
	}
	if r.ExtractDate.Valid {
		entrepreneur.ExtractDate = &model.Date{Time: r.ExtractDate.Time}
	}
	if r.Email.Valid {
		entrepreneur.Email = &r.Email.String
	}
	if r.PfrRegNumber.Valid {
		entrepreneur.PfrRegNumber = &r.PfrRegNumber.String
	}
	if r.FssRegNumber.Valid {
		entrepreneur.FssRegNumber = &r.FssRegNumber.String
	}
	if r.BankruptcyDate.Valid {
		entrepreneur.BankruptcyDate = &model.Date{Time: r.BankruptcyDate.Time}
	}
	if r.BankruptcyCaseNumber.Valid {
		entrepreneur.BankruptcyCaseNumber = &r.BankruptcyCaseNumber.String
	}
	if r.LastGrn.Valid {
		entrepreneur.LastGrn = &r.LastGrn.String
	}
	if r.LastGrnDate.Valid {
		entrepreneur.LastGrnDate = &model.Date{Time: r.LastGrnDate.Time}
	}
	if r.SourceFile.Valid {
		entrepreneur.SourceFile = &r.SourceFile.String
	}

	// Address
	entrepreneur.Address = &model.Address{}
	if r.PostalCode.Valid {
		entrepreneur.Address.PostalCode = &r.PostalCode.String
	}
	if r.RegionCode.Valid {
		entrepreneur.Address.RegionCode = &r.RegionCode.String
	}
	if r.Region.Valid {
		entrepreneur.Address.Region = &r.Region.String
	}
	if r.District.Valid {
		entrepreneur.Address.District = &r.District.String
	}
	if r.City.Valid {
		entrepreneur.Address.City = &r.City.String
	}
	if r.Locality.Valid {
		entrepreneur.Address.Locality = &r.Locality.String
	}
	if r.FullAddress.Valid {
		entrepreneur.Address.FullAddress = &r.FullAddress.String
	}
	if r.FiasID.Valid {
		entrepreneur.Address.FiasID = &r.FiasID.String
	}

	// Main activity
	if r.OkvedMainCode.Valid {
		entrepreneur.MainActivity = &model.Activity{
			Code:   r.OkvedMainCode.String,
			IsMain: true,
		}
		if r.OkvedMainName.Valid {
			entrepreneur.MainActivity.Name = &r.OkvedMainName.String
		}
		entrepreneur.Activities = append(entrepreneur.Activities, entrepreneur.MainActivity)
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
		entrepreneur.Activities = append(entrepreneur.Activities, activity)
	}

	// Authorities
	if r.RegAuthorityCode.Valid || r.RegAuthorityName.Valid {
		entrepreneur.RegAuthority = &model.Authority{}
		if r.RegAuthorityCode.Valid {
			entrepreneur.RegAuthority.Code = &r.RegAuthorityCode.String
		}
		if r.RegAuthorityName.Valid {
			entrepreneur.RegAuthority.Name = &r.RegAuthorityName.String
		}
	}
	if r.TaxAuthorityCode.Valid || r.TaxAuthorityName.Valid {
		entrepreneur.TaxAuthority = &model.Authority{}
		if r.TaxAuthorityCode.Valid {
			entrepreneur.TaxAuthority.Code = &r.TaxAuthorityCode.String
		}
		if r.TaxAuthorityName.Valid {
			entrepreneur.TaxAuthority.Name = &r.TaxAuthorityName.String
		}
	}

	return entrepreneur
}

// GetByOGRNIP получает ИП по ОГРНИП
func (r *EntrepreneurRepository) GetByOGRNIP(ctx context.Context, ogrnip string) (*model.Entrepreneur, error) {
	query := `
		SELECT * FROM egrul.entrepreneurs FINAL
		WHERE ogrnip = ?
		LIMIT 1
	`

	rows, err := r.client.conn.Query(ctx, query, ogrnip)
	if err != nil {
		return nil, fmt.Errorf("query entrepreneur by ogrnip: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var row entrepreneurRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan entrepreneur row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for entrepreneur", zap.String("ogrnip", ogrnip), zap.Error(err))
		}
		return row.toModel(), nil
	}

	return nil, nil
}

// GetByINN получает ИП по ИНН
func (r *EntrepreneurRepository) GetByINN(ctx context.Context, inn string) (*model.Entrepreneur, error) {
	query := `
		SELECT * FROM egrul.entrepreneurs FINAL
		WHERE inn = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	rows, err := r.client.conn.Query(ctx, query, inn)
	if err != nil {
		return nil, fmt.Errorf("query entrepreneur by inn: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var row entrepreneurRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan entrepreneur row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for entrepreneur by inn", zap.String("inn", inn), zap.Error(err))
		}
		return row.toModel(), nil
	}

	return nil, nil
}

// List возвращает список ИП с фильтрацией и пагинацией
func (r *EntrepreneurRepository) List(ctx context.Context, filter *model.EntrepreneurFilter, pagination *model.Pagination) ([]*model.Entrepreneur, int, error) {
	whereClause, args := r.buildWhereClause(filter)

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT count() FROM egrul.entrepreneurs FINAL
		%s
	`, whereClause)

	var totalCount uint64
	countRow := r.client.conn.QueryRow(ctx, countQuery, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count entrepreneurs: %w", err)
	}

	// Data query
	// Используем прямые значения для LIMIT/OFFSET, так как ClickHouse может не поддерживать параметризацию для них
	dataQuery := fmt.Sprintf(`
		SELECT * FROM egrul.entrepreneurs FINAL
		%s
		ORDER BY updated_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := r.client.conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []*model.Entrepreneur
	for rows.Next() {
		var row entrepreneurRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, 0, fmt.Errorf("scan entrepreneur row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for entrepreneur in list", zap.String("ogrnip", row.Ogrnip), zap.Error(err))
		}
		entrepreneurs = append(entrepreneurs, row.toModel())
	}

	return entrepreneurs, int(totalCount), nil
}

// Search выполняет текстовый поиск ИП
func (r *EntrepreneurRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Entrepreneur, error) {
	searchQuery := `
		SELECT * FROM egrul.entrepreneurs FINAL
		WHERE 
			concat(last_name, ' ', first_name, ' ', coalesce(middle_name, '')) ILIKE ?
			OR inn LIKE ?
			OR ogrnip LIKE ?
		ORDER BY 
			CASE 
				WHEN ogrnip = ? THEN 1
				WHEN inn = ? THEN 2
				ELSE 3
			END,
			last_name, first_name
		LIMIT ? OFFSET ?
	`

	pattern := "%" + query + "%"
	exactPattern := query + "%"

	rows, err := r.client.conn.Query(ctx, searchQuery,
		pattern, exactPattern, exactPattern,
		query, query,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("search entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []*model.Entrepreneur
	for rows.Next() {
		var row entrepreneurRow
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("scan entrepreneur row: %w", err)
		}
		if err := r.loadAdditionalActivities(ctx, &row); err != nil {
			r.logger.Warn("failed to load additional activities for entrepreneur in search", zap.String("ogrnip", row.Ogrnip), zap.Error(err))
		}
		entrepreneurs = append(entrepreneurs, row.toModel())
	}

	return entrepreneurs, nil
}

func (r *EntrepreneurRepository) buildWhereClause(filter *model.EntrepreneurFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if filter.Ogrnip != nil && *filter.Ogrnip != "" {
		conditions = append(conditions, "ogrnip = ?")
		args = append(args, *filter.Ogrnip)
	}
	if filter.Inn != nil && *filter.Inn != "" {
		conditions = append(conditions, "inn = ?")
		args = append(args, *filter.Inn)
	}
	if filter.Name != nil && *filter.Name != "" {
		conditions = append(conditions, "concat(last_name, ' ', first_name, ' ', coalesce(middle_name, '')) ILIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}
	if filter.LastName != nil && *filter.LastName != "" {
		conditions = append(conditions, "last_name ILIKE ?")
		args = append(args, *filter.LastName+"%")
	}
	if filter.FirstName != nil && *filter.FirstName != "" {
		conditions = append(conditions, "first_name ILIKE ?")
		args = append(args, *filter.FirstName+"%")
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
	if filter.IsBankrupt != nil {
		if *filter.IsBankrupt {
			conditions = append(conditions, "is_bankrupt = 1")
		} else {
			conditions = append(conditions, "is_bankrupt = 0")
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// loadAdditionalActivities загружает дополнительные ОКВЭД ИП из таблицы entrepreneurs_okved_additional
func (r *EntrepreneurRepository) loadAdditionalActivities(ctx context.Context, row *entrepreneurRow) error {
	const q = `
		SELECT okved_code, okved_name
		FROM egrul.entrepreneurs_okved_additional
		WHERE ogrnip = ?
	`

	rows, err := r.client.conn.Query(ctx, q, row.Ogrnip)
	if err != nil {
		return fmt.Errorf("query entrepreneurs_okved_additional: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var name sql.NullString
		if err := rows.Scan(&code, &name); err != nil {
			return fmt.Errorf("scan entrepreneurs_okved_additional row: %w", err)
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

