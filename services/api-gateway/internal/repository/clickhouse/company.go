package clickhouse

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"go.uber.org/zap"
)

// agentLog пишет отладочную информацию в NDJSON-файл для debug-сессии
func agentLog(runID, location, message string, data map[string]interface{}) {
	entry := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        runID,
		"hypothesisId": "region-filter",
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}

	// Используем переменную окружения или путь по умолчанию
	logPath := os.Getenv("DEBUG_LOG_PATH")
	if logPath == "" {
		logPath = "/Users/konstantin/cursor/egrul/.cursor/debug.log"
	}

	// Создаем директорию, если её нет
	dir := logPath[:strings.LastIndex(logPath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Если не удалось создать директорию, логируем в stderr
		enc := json.NewEncoder(os.Stderr)
		_ = enc.Encode(entry)
		return
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Если не удалось записать в файл, логируем в stderr
		enc := json.NewEncoder(os.Stderr)
		_ = enc.Encode(entry)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	_ = enc.Encode(entry)
}

// getRegionNameByCode возвращает название региона по его коду
func getRegionNameByCode(code string) string {
	regionMap := map[string]string{
		"01": "Республика Адыгея", "02": "Республика Башкортостан", "03": "Республика Бурятия",
		"04": "Республика Алтай", "05": "Республика Дагестан", "06": "Республика Ингушетия",
		"07": "Кабардино-Балкарская Республика", "08": "Республика Калмыкия", "09": "Карачаево-Черкесская Республика",
		"10": "Республика Карелия", "11": "Республика Коми", "12": "Республика Марий Эл",
		"13": "Республика Мордовия", "14": "Республика Саха (Якутия)", "15": "Республика Северная Осетия - Алания",
		"16": "Республика Татарстан", "17": "Республика Тыва", "18": "Удмуртская Республика",
		"19": "Республика Хакасия", "20": "Чеченская Республика", "21": "Чувашская Республика",
		"22": "Алтайский край", "23": "Краснодарский край", "24": "Красноярский край",
		"25": "Приморский край", "26": "Ставропольский край", "27": "Хабаровский край",
		"28": "Амурская область", "29": "Архангельская область", "30": "Астраханская область",
		"31": "Белгородская область", "32": "Брянская область", "33": "Владимирская область",
		"34": "Волгоградская область", "35": "Вологодская область", "36": "Воронежская область",
		"37": "Ивановская область", "38": "Иркутская область", "39": "Калининградская область",
		"40": "Калужская область", "41": "Камчатский край", "42": "Кемеровская область",
		"43": "Кировская область", "44": "Костромская область", "45": "Курганская область",
		"46": "Курская область", "47": "Ленинградская область", "48": "Липецкая область",
		"49": "Магаданская область", "50": "Московская область", "51": "Мурманская область",
		"52": "Нижегородская область", "53": "Новгородская область", "54": "Новосибирская область",
		"55": "Омская область", "56": "Оренбургская область", "57": "Орловская область",
		"58": "Пензенская область", "59": "Пермский край", "60": "Псковская область",
		"61": "Ростовская область", "62": "Рязанская область", "63": "Самарская область",
		"64": "Саратовская область", "65": "Сахалинская область", "66": "Свердловская область",
		"67": "Смоленская область", "68": "Тамбовская область", "69": "Тверская область",
		"70": "Томская область", "71": "Тульская область", "72": "Тюменская область",
		"73": "Ульяновская область", "74": "Челябинская область", "75": "Забайкальский край",
		"76": "Ярославская область", "77": "Москва", "78": "Санкт-Петербург",
		"79": "Еврейская автономная область", "83": "Ненецкий автономный округ",
		"86": "Ханты-Мансийский автономный округ - Югра", "87": "Чукотский автономный округ",
		"89": "Ямало-Ненецкий автономный округ", "92": "Севастополь", "95": "Чеченская Республика",
	}
	if name, ok := regionMap[code]; ok {
		return name
	}
	return ""
}

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
	KladrCode            sql.NullString  `ch:"kladr_code"`
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
		Status:          r.determineEntityStatus(),
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
	if r.KladrCode.Valid {
		company.Address.KladrCode = &r.KladrCode.String
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

// determineEntityStatus определяет статус компании на основе statusCode и terminationDate
func (r *companyRow) determineEntityStatus() model.EntityStatus {
	// Если есть дата прекращения деятельности, значит компания закрыта
	if r.TerminationDate.Valid {
		return model.EntityStatusLiquidated
	}

	// Если есть код статуса, используем его
	if r.StatusCode.Valid && r.StatusCode.String != "" {
		code := r.StatusCode.String
		
		// Коды ликвидации
		if code == "101" {
			return model.EntityStatusLiquidating
		}
		
		// Коды исключения из реестра (недействующие)
		if code == "105" || code == "106" || code == "107" {
			return model.EntityStatusLiquidated
		}
		
		// Коды банкротства (113-117)
		if code == "113" || code == "114" || code == "115" || code == "116" || code == "117" {
			return model.EntityStatusBankrupt
		}
		
		// Коды реорганизации (121-139)
		if len(code) == 3 && (code[0:2] == "12" || code[0:2] == "13") {
			return model.EntityStatusReorganizing
		}
		
		// Коды недействительности регистрации (701, 702, 801, 802)
		if len(code) == 3 && (code[0:2] == "70" || code[0:2] == "80") {
			return model.EntityStatusLiquidated
		}
		
		// Остальные коды (например, 111 - уменьшение капитала, 112 - изменение места нахождения)
		// считаем действующими
		return model.EntityStatusActive
	}

	// По умолчанию считаем действующей
	return model.EntityStatusActive
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
	// #region agent log - проверка существующих region_code в базе
	if filter != nil && filter.RegionCode != nil && *filter.RegionCode != "" {
		// Проверяем общее количество записей
		totalCountQuery := `SELECT count() FROM egrul.companies FINAL`
		var totalCount uint64
		if err := r.client.conn.QueryRow(ctx, totalCountQuery).Scan(&totalCount); err == nil {
			agentLog("run-filters", "company.go:List:totalCount", "total companies in DB", map[string]interface{}{
				"totalCount": totalCount,
			})
		}

		// Проверяем количество записей с заполненным region_code
		regionCountQuery := `SELECT count() FROM egrul.companies FINAL WHERE region_code IS NOT NULL AND region_code != ''`
		var regionCount uint64
		if err := r.client.conn.QueryRow(ctx, regionCountQuery).Scan(&regionCount); err == nil {
			agentLog("run-filters", "company.go:List:regionCount", "companies with region_code", map[string]interface{}{
				"regionCount": regionCount,
			})
		}

		// Проверяем количество записей с заполненным region (название региона)
		regionNameCountQuery := `SELECT count() FROM egrul.companies FINAL WHERE region IS NOT NULL AND region != ''`
		var regionNameCount uint64
		if err := r.client.conn.QueryRow(ctx, regionNameCountQuery).Scan(&regionNameCount); err == nil {
			agentLog("run-filters", "company.go:List:regionNameCount", "companies with region name", map[string]interface{}{
				"regionNameCount": regionNameCount,
			})
		}

		// Проверяем количество записей с заполненным full_address (может содержать информацию о регионе)
		fullAddressCountQuery := `SELECT count() FROM egrul.companies FINAL WHERE full_address IS NOT NULL AND full_address != ''`
		var fullAddressCount uint64
		if err := r.client.conn.QueryRow(ctx, fullAddressCountQuery).Scan(&fullAddressCount); err == nil {
			agentLog("run-filters", "company.go:List:fullAddressCount", "companies with full_address", map[string]interface{}{
				"fullAddressCount": fullAddressCount,
			})
		}

		// Проверяем, есть ли в full_address упоминания "Московская" или "Москва"
		checkAddressQuery := `SELECT count() FROM egrul.companies FINAL WHERE full_address ILIKE ?`
		var moscowAddressCount uint64
		if err := r.client.conn.QueryRow(ctx, checkAddressQuery, "%Московск%").Scan(&moscowAddressCount); err == nil {
			agentLog("run-filters", "company.go:List:moscowAddressCount", "companies with Moscow in full_address", map[string]interface{}{
				"moscowAddressCount": moscowAddressCount,
			})
		}

		// Получаем примеры названий регионов
		sampleRegionNamesQuery := `
			SELECT DISTINCT region, count() as cnt
			FROM egrul.companies FINAL
			WHERE region IS NOT NULL AND region != ''
			GROUP BY region
			ORDER BY cnt DESC
			LIMIT 10
		`
		rows2, err2 := r.client.conn.Query(ctx, sampleRegionNamesQuery)
		if err2 != nil {
			agentLog("run-filters", "company.go:List:sampleRegionNames:error", "error querying sample region names", map[string]interface{}{
				"error": err2.Error(),
			})
		} else {
			var sampleRegionNames []string
			for rows2.Next() {
				var rn string
				var cnt uint64
				if err := rows2.Scan(&rn, &cnt); err == nil {
					sampleRegionNames = append(sampleRegionNames, fmt.Sprintf("%s:%d", rn, cnt))
				}
			}
			rows2.Close()
			agentLog("run-filters", "company.go:List:sampleRegionNames", "sample region names from DB", map[string]interface{}{
				"sampleRegionNames": sampleRegionNames,
				"requestedRegionName": getRegionNameByCode(*filter.RegionCode),
			})
		}

		// Получаем примеры region_code
		sampleQuery := `
			SELECT DISTINCT region_code, count() as cnt
			FROM egrul.companies FINAL
			WHERE region_code IS NOT NULL AND region_code != ''
			GROUP BY region_code
			ORDER BY cnt DESC
			LIMIT 10
		`
		rows, err := r.client.conn.Query(ctx, sampleQuery)
		if err != nil {
			agentLog("run-filters", "company.go:List:sampleRegionCodes:error", "error querying sample region codes", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			var sampleRegionCodes []string
			for rows.Next() {
				var rc string
				var cnt uint64
				if err := rows.Scan(&rc, &cnt); err == nil {
					sampleRegionCodes = append(sampleRegionCodes, fmt.Sprintf("%s:%d", rc, cnt))
				}
			}
			rows.Close()
			agentLog("run-filters", "company.go:List:sampleRegionCodes", "sample region codes from DB", map[string]interface{}{
				"sampleRegionCodes": sampleRegionCodes,
				"requestedRegionCode": *filter.RegionCode,
			})
		}

		// Проверяем, есть ли записи с запрошенным region_code (без FINAL для быстрой проверки)
		checkQuery := `SELECT count() FROM egrul.companies WHERE region_code = ?`
		var checkCount uint64
		if err := r.client.conn.QueryRow(ctx, checkQuery, *filter.RegionCode).Scan(&checkCount); err == nil {
			agentLog("run-filters", "company.go:List:checkRegionCode", "check region_code without FINAL", map[string]interface{}{
				"requestedRegionCode": *filter.RegionCode,
				"checkCount": checkCount,
			})
		}
	}
	// #endregion

	// Если есть фильтр по ФИО учредителя, сначала получаем список ОГРН компаний
	var founderOgrns []string
	var filterForWhereClause *model.CompanyFilter = filter
	if filter != nil && filter.FounderName != nil && *filter.FounderName != "" {
		searchTerm := strings.TrimSpace(*filter.FounderName)
		if searchTerm != "" {
			searchTerm = strings.ReplaceAll(searchTerm, "%", "\\%")
			searchTerm = strings.ReplaceAll(searchTerm, "_", "\\_")
			pattern := "%" + searchTerm + "%"
			
			// Выполняем подзапрос отдельно для получения списка ОГРН
			founderQuery := `
				SELECT DISTINCT company_ogrn FROM egrul.founders FINAL 
				WHERE (founder_name ILIKE ?) OR 
				      (founder_last_name IS NOT NULL AND founder_last_name ILIKE ?) OR 
				      (founder_first_name IS NOT NULL AND founder_first_name ILIKE ?) OR 
				      (founder_middle_name IS NOT NULL AND founder_middle_name ILIKE ?)
			`
			
			rows, err := r.client.conn.Query(ctx, founderQuery, pattern, pattern, pattern, pattern)
			if err != nil {
				r.logger.Warn("failed to query founders by name", zap.String("founderName", *filter.FounderName), zap.Error(err))
			} else {
				defer rows.Close()
				for rows.Next() {
					var ogrn string
					if err := rows.Scan(&ogrn); err == nil {
						founderOgrns = append(founderOgrns, ogrn)
					}
				}
			}
			
			// Если не найдено компаний по учредителям, возвращаем пустой результат
			if len(founderOgrns) == 0 {
				return []*model.Company{}, 0, nil
			}
			
			// Создаем копию фильтра без founderName для buildWhereClause
			filterCopy := *filter
			filterCopy.FounderName = nil
			filterForWhereClause = &filterCopy
		}
	}
	
	whereClause, args := r.buildWhereClause(filterForWhereClause)
	
	// Если есть список ОГРН из поиска по учредителям, добавляем условие
	if len(founderOgrns) > 0 {
		if whereClause != "" {
			whereClause += " AND "
		} else {
			whereClause = "WHERE "
		}
		// Формируем условие IN с конкретными значениями
		placeholders := make([]string, len(founderOgrns))
		for i := range founderOgrns {
			placeholders[i] = "?"
			args = append(args, founderOgrns[i])
		}
		whereClause += fmt.Sprintf("ogrn IN (%s)", strings.Join(placeholders, ","))
	}
	
	orderClause := r.buildOrderClause(sort)

	limit := pagination.GetLimit()
	offset := pagination.GetOffset()

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT count() FROM egrul.companies FINAL
		%s
	`, whereClause)

	// #region agent log
	agentLog("run-filters", "company.go:List:countQuery", "executing count query", map[string]interface{}{
		"countQuery": countQuery,
		"args":       args,
		"argsCount":  len(args),
		"hasRegionCode": filter != nil && filter.RegionCode != nil && *filter.RegionCode != "",
		"regionCode": func() string {
			if filter != nil && filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
	})
	// #endregion

	var totalCount uint64
	countRow := r.client.conn.QueryRow(ctx, countQuery, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count companies: %w", err)
	}

	// #region agent log
	agentLog("run-filters", "company.go:List:countResult", "count query result", map[string]interface{}{
		"totalCount": totalCount,
		"countQuery": countQuery,
		"args":       args,
	})
	// #endregion

	// Data query
	// Используем прямые значения для LIMIT/OFFSET, так как ClickHouse может не поддерживать параметризацию для них
	dataQuery := fmt.Sprintf(`
		SELECT * FROM egrul.companies FINAL
		%s
		%s
		LIMIT %d OFFSET %d
	`, whereClause, orderClause, limit, offset)

	// #region agent log
	agentLog("run-filters", "company.go:List:dataQuery", "executing data query", map[string]interface{}{
		"dataQuery":   dataQuery,
		"args":        args,
		"argsCount":   len(args),
		"limit":       limit,
		"offset":      offset,
		"hasRegionCode": filter != nil && filter.RegionCode != nil && *filter.RegionCode != "",
		"regionCode": func() string {
			if filter != nil && filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
	})
	// #endregion

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
		// Упрощаем фильтр по региону: используем только region_code.
		// Это гарантирует корректное применение логического И с другими условиями (status_code и т.д.)
		conditions = append(conditions, "region_code = ?")
		args = append(args, *filter.RegionCode)
		// #region agent log
		agentLog("run-filters", "company.go:buildWhereClause", "applying simple regionCode filter", map[string]interface{}{
			"regionCode": *filter.RegionCode,
			"condition":  "region_code = ?",
		})
		// #endregion
	}
	if filter.Region != nil && *filter.Region != "" {
		conditions = append(conditions, "region ILIKE ?")
		args = append(args, "%"+*filter.Region+"%")
	}
	if filter.Okved != nil && *filter.Okved != "" {
		conditions = append(conditions, "(okved_main_code = ? OR has(okved_additional, ?))")
		args = append(args, *filter.Okved, *filter.Okved)
	}
	// Фильтрация по текстовому статусу (старый вариант, оставляем для обратной совместимости)
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
	// Фильтрация по коду статуса (status_code)
	if filter.StatusCode != nil && *filter.StatusCode != "" {
		conditions = append(conditions, "status_code = ?")
		args = append(args, *filter.StatusCode)
	}
	if len(filter.StatusCodeIn) > 0 {
		placeholders := make([]string, len(filter.StatusCodeIn))
		for i, code := range filter.StatusCodeIn {
			placeholders[i] = "?"
			args = append(args, code)
		}
		conditions = append(conditions, fmt.Sprintf("status_code IN (%s)", strings.Join(placeholders, ",")))
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
	// Примечание: фильтр по founderName обрабатывается в методе List отдельно,
	// так как требует выполнения отдельного подзапроса для получения списка ОГРН

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *CompanyRepository) buildOrderClause(sort *model.CompanySort) string {
	if sort == nil {
		r.logger.Info("buildOrderClause: sort is nil, using default ORDER BY updated_at DESC")
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
	if sort.Order != nil && *sort.Order == model.SortOrderDesc {
		order = "DESC"
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s", field, order)
	r.logger.Info("buildOrderClause", zap.String("field", field), zap.String("order", order), zap.String("orderClause", orderClause))
	return orderClause
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

