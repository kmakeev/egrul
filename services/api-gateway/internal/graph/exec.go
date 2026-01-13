package graph

// This file provides manual schema execution without full gqlgen generation
// For production use, run: go run github.com/99designs/gqlgen generate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
)

// ManualHandler обрабатывает GraphQL запросы без генерации
type ManualHandler struct {
	resolver *Resolver
}

// NewManualHandler создает новый обработчик
func NewManualHandler(resolver *Resolver) *ManualHandler {
	return &ManualHandler{resolver: resolver}
}

// GraphQLRequest структура запроса
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse структура ответа
type GraphQLResponse struct {
	Data   interface{}      `json:"data,omitempty"`
	Errors []GraphQLError   `json:"errors,omitempty"`
}

// GraphQLError ошибка GraphQL
type GraphQLError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// ServeHTTP обрабатывает HTTP запросы
func (h *ManualHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req GraphQLRequest

	if r.Method == "GET" {
		req.Query = r.URL.Query().Get("query")
		req.OperationName = r.URL.Query().Get("operationName")
		if vars := r.URL.Query().Get("variables"); vars != "" {
			json.Unmarshal([]byte(vars), &req.Variables)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	if req.Query == "" {
		h.writeError(w, "query is required", http.StatusBadRequest)
		return
	}

	result, err := h.execute(r.Context(), &req)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (h *ManualHandler) writeError(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(GraphQLResponse{
		Errors: []GraphQLError{{Message: msg}},
	})
}

func (h *ManualHandler) execute(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	query := strings.TrimSpace(req.Query)
	
	// Простой парсер для демонстрации
	// В продакшене используйте gqlgen generate
	
	if strings.Contains(query, "__schema") || strings.Contains(query, "__type") {
		return h.handleIntrospection(ctx, query)
	}
	
	// Определяем операцию
	if strings.Contains(query, "company(") || strings.Contains(query, "company (") {
		return h.handleCompanyQuery(ctx, req)
	}
	if strings.Contains(query, "companyByInn") {
		return h.handleCompanyByInnQuery(ctx, req)
	}
	if strings.Contains(query, "companies(") || strings.Contains(query, "companies {") {
		return h.handleCompaniesQuery(ctx, req)
	}
	if strings.Contains(query, "searchCompanies") {
		return h.handleSearchCompaniesQuery(ctx, req)
	}
	if strings.Contains(query, "entrepreneur(") || strings.Contains(query, "entrepreneur (") {
		return h.handleEntrepreneurQuery(ctx, req)
	}
	if strings.Contains(query, "entrepreneurs(") || strings.Contains(query, "entrepreneurs {") {
		return h.handleEntrepreneursQuery(ctx, req)
	}
	if strings.Contains(query, "search(") || strings.Contains(query, "search {") {
		return h.handleSearchQuery(ctx, req)
	}
	if strings.Contains(query, "statistics") {
		return h.handleStatisticsQuery(ctx, req)
	}

	return &GraphQLResponse{
		Errors: []GraphQLError{{Message: "unsupported query, please use gqlgen generate for full support"}},
	}, nil
}

func (h *ManualHandler) handleIntrospection(ctx context.Context, query string) (*GraphQLResponse, error) {
	// Базовая схема для интроспекции
	schema := map[string]interface{}{
		"__schema": map[string]interface{}{
			"queryType": map[string]string{"name": "Query"},
			"types": []map[string]interface{}{
				{"kind": "OBJECT", "name": "Query"},
				{"kind": "OBJECT", "name": "Company"},
				{"kind": "OBJECT", "name": "Entrepreneur"},
				{"kind": "OBJECT", "name": "Statistics"},
			},
		},
	}
	return &GraphQLResponse{Data: schema}, nil
}

func (h *ManualHandler) handleCompanyQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	// #region agent log
	agentLog("history-debug", "exec.go:handleCompanyQuery", "function called", map[string]interface{}{
		"query": req.Query,
	})
	// #endregion
	
	ogrn, ok := req.Variables["ogrn"].(string)
	if !ok {
		// Try to extract from query
		ogrn = extractArgFromQuery(req.Query, "ogrn")
	}
	// #region agent log
	agentLog("run1", "exec.go:handleCompanyQuery", "parsed ogrn for company", map[string]interface{}{"ogrn": ogrn})
	// #endregion
	if ogrn == "" {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: "ogrn is required"}}}, nil
	}

	company, err := h.resolver.Query().Company(ctx, ogrn)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	// Если в запросе есть поля founders, history, relatedCompanies, licenses или branches, загружаем их
	hasFounders := strings.Contains(req.Query, "founders")
	hasHistory := strings.Contains(req.Query, "history")
	hasRelatedCompanies := strings.Contains(req.Query, "relatedCompanies")
	hasLicenses := strings.Contains(req.Query, "licenses")
	hasBranches := strings.Contains(req.Query, "branches")

	// TEMPORARY DEBUG LOGGING
	fmt.Printf("=== COMPANY QUERY DEBUG ===\n")
	fmt.Printf("OGRN: %s\n", ogrn)
	fmt.Printf("Query: %s\n", req.Query)
	fmt.Printf("hasFounders: %v\n", hasFounders)
	fmt.Printf("hasHistory: %v\n", hasHistory)
	fmt.Printf("hasRelatedCompanies: %v\n", hasRelatedCompanies)
	fmt.Printf("hasLicenses: %v\n", hasLicenses)
	fmt.Printf("hasBranches: %v\n", hasBranches)
	fmt.Printf("company != nil: %v\n", company != nil)
	fmt.Printf("===========================\n")

	// #region agent log
	agentLog("history-debug", "exec.go:handleCompanyQuery", "checking for founders, history, relatedCompanies, licenses and branches fields", map[string]interface{}{
		"hasFounders": hasFounders,
		"hasHistory": hasHistory,
		"hasRelatedCompanies": hasRelatedCompanies,
		"hasLicenses": hasLicenses,
		"hasBranches": hasBranches,
		"query": req.Query,
	})
	// #endregion

	if (hasFounders || hasHistory || hasRelatedCompanies || hasLicenses || hasBranches) && company != nil {
		var founders []*model.Founder
		var history []*model.HistoryRecord
		var relatedCompanies []*model.RelatedCompany
		var licenses []*model.License
		var branches []*model.Branch

		// Загружаем учредителей если запрошены
		if hasFounders {
			agentLog("history-debug", "exec.go:handleCompanyQuery", "loading founders", map[string]interface{}{"ogrn": company.Ogrn})
			foundersResult, err := h.resolver.Company().Founders(ctx, company, nil, nil)
			if err == nil {
				founders = foundersResult
				agentLog("history-debug", "exec.go:handleCompanyQuery", "founders loaded", map[string]interface{}{"count": len(founders)})
			} else {
				agentLog("history-debug", "exec.go:handleCompanyQuery", "founders error", map[string]interface{}{"error": err.Error()})
			}
		}
		
		// Загружаем историю если запрошена
		if hasHistory {
			agentLog("history-debug", "exec.go:handleCompanyQuery", "loading history", map[string]interface{}{"ogrn": company.Ogrn})
			
			// Извлекаем параметры limit и offset для истории из GraphQL запроса
			var historyLimit, historyOffset *int
			
			// Сначала пробуем из variables
			if limitVar, ok := req.Variables["limit"].(float64); ok {
				l := int(limitVar)
				historyLimit = &l
			}
			if offsetVar, ok := req.Variables["offset"].(float64); ok {
				o := int(offsetVar)
				historyOffset = &o
			}
			
			// Если не нашли в variables, пробуем извлечь из строки запроса
			if historyLimit == nil || historyOffset == nil {
				// Ищем паттерн history(limit: X, offset: Y)
				historyIdx := strings.Index(req.Query, "history(")
				if historyIdx != -1 {
					// Находим закрывающую скобку
					start := historyIdx + len("history(")
					end := strings.Index(req.Query[start:], ")")
					if end != -1 {
						argsStr := req.Query[start : start+end]
						
						// Парсим limit
						if limitIdx := strings.Index(argsStr, "limit:"); limitIdx != -1 {
							limitStart := limitIdx + len("limit:")
							for limitStart < len(argsStr) && (argsStr[limitStart] == ' ' || argsStr[limitStart] == '\t') {
								limitStart++
							}
							var limitVal int
							if n, err := fmt.Sscanf(argsStr[limitStart:], "%d", &limitVal); err == nil && n == 1 {
								historyLimit = &limitVal
							}
						}
						
						// Парсим offset
						if offsetIdx := strings.Index(argsStr, "offset:"); offsetIdx != -1 {
							offsetStart := offsetIdx + len("offset:")
							for offsetStart < len(argsStr) && (argsStr[offsetStart] == ' ' || argsStr[offsetStart] == '\t') {
								offsetStart++
							}
							var offsetVal int
							if n, err := fmt.Sscanf(argsStr[offsetStart:], "%d", &offsetVal); err == nil && n == 1 {
								historyOffset = &offsetVal
							}
						}
					}
				}
			}
			
			agentLog("history-debug", "exec.go:handleCompanyQuery", "extracted history parameters", map[string]interface{}{
				"limit": historyLimit,
				"offset": historyOffset,
				"query": req.Query,
				"variables": req.Variables,
			})
			
			historyResult, err := h.resolver.Company().History(ctx, company, historyLimit, historyOffset)
			if err == nil {
				history = historyResult
				agentLog("history-debug", "exec.go:handleCompanyQuery", "history loaded", map[string]interface{}{
					"count": len(history),
					"limit": historyLimit,
					"offset": historyOffset,
				})
			} else {
				agentLog("history-debug", "exec.go:handleCompanyQuery", "history error", map[string]interface{}{"error": err.Error()})
			}
		}
		
		// Загружаем связанные компании если запрошены
		if hasRelatedCompanies {
			agentLog("history-debug", "exec.go:handleCompanyQuery", "loading related companies", map[string]interface{}{"ogrn": company.Ogrn})

			// Для связанных компаний используем стандартные параметры
			relatedResult, err := h.resolver.Company().RelatedCompanies(ctx, company, nil, nil)
			if err == nil {
				relatedCompanies = relatedResult
				agentLog("history-debug", "exec.go:handleCompanyQuery", "related companies loaded", map[string]interface{}{
					"count": len(relatedCompanies),
				})
			} else {
				agentLog("history-debug", "exec.go:handleCompanyQuery", "related companies error", map[string]interface{}{"error": err.Error()})
			}
		}

		// Загружаем лицензии если запрошены
		if hasLicenses {
			agentLog("history-debug", "exec.go:handleCompanyQuery", "loading licenses", map[string]interface{}{"ogrn": company.Ogrn})
			licensesResult, err := h.resolver.Company().Licenses(ctx, company)
			if err == nil {
				licenses = licensesResult
				agentLog("history-debug", "exec.go:handleCompanyQuery", "licenses loaded", map[string]interface{}{
					"count": len(licenses),
				})
			} else {
				agentLog("history-debug", "exec.go:handleCompanyQuery", "licenses error", map[string]interface{}{"error": err.Error()})
			}
		}

		// Загружаем филиалы если запрошены
		if hasBranches {
			agentLog("history-debug", "exec.go:handleCompanyQuery", "loading branches", map[string]interface{}{"ogrn": company.Ogrn})
			branchesResult, err := h.resolver.Company().Branches(ctx, company)
			if err == nil {
				branches = branchesResult
				agentLog("history-debug", "exec.go:handleCompanyQuery", "branches loaded", map[string]interface{}{
					"count": len(branches),
				})
			} else {
				agentLog("history-debug", "exec.go:handleCompanyQuery", "branches error", map[string]interface{}{"error": err.Error()})
			}
		}

		// Создаем map с данными компании
		companyData := map[string]interface{}{
			"ogrn":          company.Ogrn,
			"ogrnDate":      company.OgrnDate,
			"inn":           company.Inn,
			"kpp":           company.Kpp,
			"fullName":      company.FullName,
			"shortName":     company.ShortName,
			"brandName":     company.BrandName,
			"legalForm":     company.LegalForm,
			"status":        company.Status,
			"statusCode":    company.StatusCode,
			"terminationMethod": company.TerminationMethod,
			"registrationDate": company.RegistrationDate,
			"terminationDate": company.TerminationDate,
			"extractDate":   company.ExtractDate,
			"address":       company.Address,
			"email":         company.Email,
			"capital":       company.Capital,
			"director":      company.Director,
			"mainActivity":  company.MainActivity,
			"activities":    company.Activities,
			"regAuthority":  company.RegAuthority,
			"taxAuthority":  company.TaxAuthority,
			"pfrRegNumber":  company.PfrRegNumber,
			"fssRegNumber":  company.FssRegNumber,
			"foundersCount": company.FoundersCount,
			"licensesCount": company.LicensesCount,
			"branchesCount": company.BranchesCount,
			"isBankrupt":    company.IsBankrupt,
			"bankruptcyStage": company.BankruptcyStage,
			"isLiquidating": company.IsLiquidating,
			"isReorganizing": company.IsReorganizing,
			"lastGrn":       company.LastGrn,
			"lastGrnDate":   company.LastGrnDate,
			"sourceFile":    company.SourceFile,
			"versionDate":   company.VersionDate,
			"createdAt":     company.CreatedAt,
			"updatedAt":     company.UpdatedAt,
		}
		
		// Добавляем founders если они были загружены
		if founders != nil {
			companyData["founders"] = founders
		}
		
		// Добавляем history если она была загружена
		if history != nil {
			companyData["history"] = history
			
			// Также добавляем historyCount если запрошен
			if strings.Contains(req.Query, "historyCount") {
				historyCount, err := h.resolver.Company().HistoryCount(ctx, company)
				if err == nil {
					companyData["historyCount"] = historyCount
					agentLog("history-debug", "exec.go:handleCompanyQuery", "historyCount loaded", map[string]interface{}{"count": historyCount})
				}
			}
		}
		
		// Добавляем relatedCompanies если они были запрошены (даже если пустые)
		if hasRelatedCompanies {
			if relatedCompanies == nil {
				relatedCompanies = []*model.RelatedCompany{}
			}
			companyData["relatedCompanies"] = relatedCompanies
		}

		// Добавляем licenses если они были запрошены
		if hasLicenses {
			if licenses == nil {
				licenses = []*model.License{}
			}
			companyData["licenses"] = licenses
		}

		// Добавляем branches если они были запрошены
		if hasBranches {
			if branches == nil {
				branches = []*model.Branch{}
			}
			companyData["branches"] = branches
		}

		result := map[string]interface{}{
			"company": companyData,
		}
		return &GraphQLResponse{Data: result}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"company": company}}, nil
}

func (h *ManualHandler) handleCompanyByInnQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	inn, ok := req.Variables["inn"].(string)
	if !ok {
		inn = extractArgFromQuery(req.Query, "inn")
	}
	if inn == "" {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: "inn is required"}}}, nil
	}

	company, err := h.resolver.Query().CompanyByInn(ctx, inn)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"companyByInn": company}}, nil
}

func (h *ManualHandler) handleCompaniesQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	var filter *model.CompanyFilter
	var pagination *model.Pagination
	var sort *model.CompanySort

	if filterVar, ok := req.Variables["filter"].(map[string]interface{}); ok {
		agentLog("run-filters", "exec.go:handleCompaniesQuery", "received filter variables", map[string]interface{}{
			"filterVar": filterVar,
			"hasRegionCode": filterVar["regionCode"] != nil,
			"regionCode": filterVar["regionCode"],
		})
		filter = parseCompanyFilter(filterVar)
		if filter != nil {
			agentLog("run-filters", "exec.go:handleCompaniesQuery", "parsed company filter", map[string]interface{}{
				"hasRegionCode": filter.RegionCode != nil,
				"regionCode": func() string {
					if filter.RegionCode != nil {
						return *filter.RegionCode
					}
					return ""
				}(),
			})
		}
	}
	if paginationVar, ok := req.Variables["pagination"].(map[string]interface{}); ok {
		pagination = parsePagination(paginationVar)
	}
	// Если пагинация не пришла в variables, пытаемся извлечь из строки запроса
	if pagination == nil {
		pagination = extractPaginationFromQuery(req.Query)
	}
	if pagination == nil {
		pagination = &model.Pagination{}
	}
	if sortVar, ok := req.Variables["sort"].(map[string]interface{}); ok {
		sort = parseCompanySort(sortVar)
	}

	companies, err := h.resolver.Query().Companies(ctx, filter, pagination, sort)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	agentLog("run-filters", "exec.go:handleCompaniesQuery", "companies query result", map[string]interface{}{
		"totalCount": companies.TotalCount,
		"edgesCount": len(companies.Edges),
		"hasRegionCode": filter != nil && filter.RegionCode != nil,
		"regionCode": func() string {
			if filter != nil && filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
	})

	return &GraphQLResponse{Data: map[string]interface{}{"companies": companies}}, nil
}

func (h *ManualHandler) handleSearchCompaniesQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	query, _ := req.Variables["query"].(string)
	if query == "" {
		query = extractArgFromQuery(req.Query, "query")
	}
	if query == "" {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: "query is required"}}}, nil
	}

	limit := 20
	if l, ok := req.Variables["limit"].(float64); ok {
		limit = int(l)
	}

	companies, err := h.resolver.Query().SearchCompanies(ctx, query, &limit, nil)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"searchCompanies": companies}}, nil
}

func (h *ManualHandler) handleEntrepreneurQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	ogrnip, ok := req.Variables["ogrnip"].(string)
	if !ok {
		ogrnip = extractArgFromQuery(req.Query, "ogrnip")
	}
	// #region agent log
	agentLog("run1", "exec.go:handleEntrepreneurQuery", "parsed ogrnip for entrepreneur", map[string]interface{}{"ogrnip": ogrnip})
	// #endregion
	if ogrnip == "" {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: "ogrnip is required"}}}, nil
	}

	entrepreneur, err := h.resolver.Query().Entrepreneur(ctx, ogrnip)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	// Если в запросе есть поля history, historyCount или licenses, загружаем их
	hasHistory := strings.Contains(req.Query, "history")
	hasHistoryCount := strings.Contains(req.Query, "historyCount")
	hasLicenses := strings.Contains(req.Query, "licenses")

	if (hasHistory || hasHistoryCount || hasLicenses) && entrepreneur != nil {
		var history []*model.HistoryRecord
		var historyCount int
		var licenses []*model.License

		// Загружаем историю если запрошена
		if hasHistory {
			// Извлекаем параметры limit и offset для истории из GraphQL запроса
			var historyLimit, historyOffset *int
			
			// Сначала пробуем из variables
			if limitVar, ok := req.Variables["limit"].(float64); ok {
				l := int(limitVar)
				historyLimit = &l
			}
			if offsetVar, ok := req.Variables["offset"].(float64); ok {
				o := int(offsetVar)
				historyOffset = &o
			}
			
			// Если не нашли в variables, пробуем извлечь из строки запроса
			if historyLimit == nil || historyOffset == nil {
				// Ищем паттерн history(limit: X, offset: Y)
				historyIdx := strings.Index(req.Query, "history(")
				if historyIdx != -1 {
					// Находим закрывающую скобку
					start := historyIdx + len("history(")
					end := strings.Index(req.Query[start:], ")")
					if end != -1 {
						argsStr := req.Query[start : start+end]
						
						// Парсим limit
						if limitIdx := strings.Index(argsStr, "limit:"); limitIdx != -1 {
							limitStart := limitIdx + len("limit:")
							for limitStart < len(argsStr) && (argsStr[limitStart] == ' ' || argsStr[limitStart] == '\t') {
								limitStart++
							}
							var limitVal int
							if n, err := fmt.Sscanf(argsStr[limitStart:], "%d", &limitVal); err == nil && n == 1 {
								historyLimit = &limitVal
							}
						}
						
						// Парсим offset
						if offsetIdx := strings.Index(argsStr, "offset:"); offsetIdx != -1 {
							offsetStart := offsetIdx + len("offset:")
							for offsetStart < len(argsStr) && (argsStr[offsetStart] == ' ' || argsStr[offsetStart] == '\t') {
								offsetStart++
							}
							var offsetVal int
							if n, err := fmt.Sscanf(argsStr[offsetStart:], "%d", &offsetVal); err == nil && n == 1 {
								historyOffset = &offsetVal
							}
						}
					}
				}
			}
			
			historyResult, err := h.resolver.Entrepreneur().History(ctx, entrepreneur, historyLimit, historyOffset)
			if err == nil {
				history = historyResult
			}
		}
		
		// Загружаем количество записей истории если запрошено
		if hasHistoryCount {
			historyCountResult, err := h.resolver.Entrepreneur().HistoryCount(ctx, entrepreneur)
			if err == nil {
				historyCount = historyCountResult
			}
		}

		// Загружаем лицензии если запрошены
		if hasLicenses {
			licensesResult, err := h.resolver.Entrepreneur().Licenses(ctx, entrepreneur)
			if err == nil {
				licenses = licensesResult
			}
		}

		// Создаем map с данными ИП
		entrepreneurData := map[string]interface{}{
			"ogrnip":                 entrepreneur.Ogrnip,
			"ogrnipDate":             entrepreneur.OgrnipDate,
			"inn":                    entrepreneur.Inn,
			"lastName":               entrepreneur.LastName,
			"firstName":              entrepreneur.FirstName,
			"middleName":             entrepreneur.MiddleName,
			"gender":                 entrepreneur.Gender,
			"citizenshipType":        entrepreneur.CitizenshipType,
			"citizenshipCountryCode": entrepreneur.CitizenshipCountryCode,
			"citizenshipCountryName": entrepreneur.CitizenshipCountryName,
			"status":                 entrepreneur.Status,
			"statusCode":             entrepreneur.StatusCode,
			"terminationMethod":      entrepreneur.TerminationMethod,
			"registrationDate":       entrepreneur.RegistrationDate,
			"terminationDate":        entrepreneur.TerminationDate,
			"extractDate":            entrepreneur.ExtractDate,
			"address":                entrepreneur.Address,
			"email":                  entrepreneur.Email,
			"mainActivity":           entrepreneur.MainActivity,
			"activities":             entrepreneur.Activities,
			"regAuthority":           entrepreneur.RegAuthority,
			"taxAuthority":           entrepreneur.TaxAuthority,
			"pfrRegNumber":           entrepreneur.PfrRegNumber,
			"fssRegNumber":           entrepreneur.FssRegNumber,
			"licensesCount":          entrepreneur.LicensesCount,
			"isBankrupt":             entrepreneur.IsBankrupt,
			"bankruptcyDate":         entrepreneur.BankruptcyDate,
			"bankruptcyCaseNumber":   entrepreneur.BankruptcyCaseNumber,
			"lastGrn":                entrepreneur.LastGrn,
			"lastGrnDate":            entrepreneur.LastGrnDate,
			"sourceFile":             entrepreneur.SourceFile,
			"versionDate":            entrepreneur.VersionDate,
			"createdAt":              entrepreneur.CreatedAt,
			"updatedAt":              entrepreneur.UpdatedAt,
		}
		
		// Добавляем history если она была загружена
		if history != nil {
			entrepreneurData["history"] = history
		}
		
		// Добавляем historyCount если он был загружен
		if hasHistoryCount {
			entrepreneurData["historyCount"] = historyCount
		}

		// Добавляем licenses если они были запрошены
		if hasLicenses {
			if licenses == nil {
				licenses = []*model.License{}
			}
			entrepreneurData["licenses"] = licenses
		}

		result := map[string]interface{}{
			"entrepreneur": entrepreneurData,
		}
		return &GraphQLResponse{Data: result}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"entrepreneur": entrepreneur}}, nil
}

func (h *ManualHandler) handleEntrepreneursQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	var filter *model.EntrepreneurFilter
	var pagination *model.Pagination
	var sort *model.EntrepreneurSort

	if filterVar, ok := req.Variables["filter"].(map[string]interface{}); ok {
		agentLog("run-filters", "exec.go:handleEntrepreneursQuery", "received filter variables", map[string]interface{}{
			"filterVar": filterVar,
			"hasRegionCode": filterVar["regionCode"] != nil,
			"regionCode": filterVar["regionCode"],
		})
		filter = parseEntrepreneurFilter(filterVar)
		if filter != nil {
			agentLog("run-filters", "exec.go:handleEntrepreneursQuery", "parsed entrepreneur filter", map[string]interface{}{
				"hasRegionCode": filter.RegionCode != nil,
				"regionCode": func() string {
					if filter.RegionCode != nil {
						return *filter.RegionCode
					}
					return ""
				}(),
			})
		}
	}
	if paginationVar, ok := req.Variables["pagination"].(map[string]interface{}); ok {
		pagination = parsePagination(paginationVar)
	}
	if pagination == nil {
		pagination = &model.Pagination{}
	}
	if sortVar, ok := req.Variables["sort"].(map[string]interface{}); ok {
		sort = parseEntrepreneurSort(sortVar)
	}

	entrepreneurs, err := h.resolver.Query().Entrepreneurs(ctx, filter, pagination, sort)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	agentLog("run-filters", "exec.go:handleEntrepreneursQuery", "entrepreneurs query result", map[string]interface{}{
		"totalCount": entrepreneurs.TotalCount,
		"edgesCount": len(entrepreneurs.Edges),
		"hasRegionCode": filter != nil && filter.RegionCode != nil,
		"regionCode": func() string {
			if filter != nil && filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
	})

	return &GraphQLResponse{Data: map[string]interface{}{"entrepreneurs": entrepreneurs}}, nil
}

func (h *ManualHandler) handleSearchQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	query, _ := req.Variables["query"].(string)
	if query == "" {
		query = extractArgFromQuery(req.Query, "query")
	}
	if query == "" {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: "query is required"}}}, nil
	}

	limit := 10
	if l, ok := req.Variables["limit"].(float64); ok {
		limit = int(l)
	}

	result, err := h.resolver.Query().Search(ctx, query, &limit)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"search": result}}, nil
}

func (h *ManualHandler) handleStatisticsQuery(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	var filter *model.StatsFilter

	if filterVar, ok := req.Variables["filter"].(map[string]interface{}); ok {
		filter = parseStatsFilter(filterVar)
	}

	stats, err := h.resolver.Query().Statistics(ctx, filter)
	if err != nil {
		return &GraphQLResponse{Errors: []GraphQLError{{Message: err.Error()}}}, nil
	}

	return &GraphQLResponse{Data: map[string]interface{}{"statistics": stats}}, nil
}

// Вспомогательные функции парсинга
func extractArgFromQuery(query, argName string) string {
	// Простой парсер аргументов из строки запроса
	// Ищем паттерн argName: "value" или argName: $variable
	patterns := []string{
		fmt.Sprintf(`%s:\s*"([^"]+)"`, argName),
		fmt.Sprintf(`%s:\s*'([^']+)'`, argName),
	}
	
	// Игнорируем patterns - используем простой поиск
	_ = patterns
	
	// Простой поиск без регулярок, допускаем отсутствие пробела после двоеточия
	searchStr := argName + `:`
	if idx := strings.Index(query, searchStr); idx != -1 {
		// сдвигаем за имя аргумента и двоеточие
		start := idx + len(searchStr)
		// пропускаем пробелы
		for start < len(query) && query[start] == ' ' {
			start++
		}
		// дальше ожидаем кавычку
		if start < len(query) && query[start] == '"' {
			start++
			end := strings.Index(query[start:], `"`)
			if end != -1 {
				return query[start : start+end]
			}
		}
	}
	return ""
}

// agentLog пишет отладочную информацию в NDJSON-файл для debug-сессии
func agentLog(runID, location, message string, data map[string]interface{}) {
	entry := map[string]interface{}{
		"sessionId": "debug-session",
		"runId":     runID,
		"hypothesisId": "arguments-parsing",
		"location":  location,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
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

func parseCompanyFilter(data map[string]interface{}) *model.CompanyFilter {
	filter := &model.CompanyFilter{}
	if v, ok := data["inn"].(string); ok {
		filter.Inn = &v
	}
	if v, ok := data["ogrn"].(string); ok {
		filter.Ogrn = &v
	}
	if v, ok := data["name"].(string); ok {
		filter.Name = &v
	}
	if v, ok := data["regionCode"].(string); ok {
		filter.RegionCode = &v
		agentLog("run-filters", "exec.go:parseCompanyFilter", "parsed regionCode for companies", map[string]interface{}{"regionCode": v})
	}
	if v, ok := data["okved"].(string); ok {
		filter.Okved = &v
	}
	// Фильтр по ФИО учредителя
	if v, ok := data["founderName"].(string); ok && strings.TrimSpace(v) != "" {
		trimmed := strings.TrimSpace(v)
		filter.FounderName = &trimmed
		agentLog("run-filters", "exec.go:parseCompanyFilter", "parsed founderName for companies", map[string]interface{}{"founderName": trimmed})
	}
	// Статус может приходить как GraphQL enum (\"ACTIVE\") или как строка в нижнем регистре (\"active\")
	if v, ok := data["status"].(string); ok {
		status := model.EntityStatus(strings.ToUpper(v))
		if status.IsValid() {
			filter.Status = &status
		} else {
			agentLog("run-filters", "exec.go:parseCompanyFilter", "invalid status value in company filter", map[string]interface{}{
				"rawStatus": v,
			})
		}
	}
	// Множественный фильтр по статусу (на будущее)
	if raw, ok := data["statusIn"].([]interface{}); ok {
		for _, item := range raw {
			if sv, ok := item.(string); ok {
				s := model.EntityStatus(strings.ToUpper(sv))
				if s.IsValid() {
					filter.StatusIn = append(filter.StatusIn, s)
				}
			}
		}
	}
	// Фильтрация по коду статуса (status_code) как строке
	if v, ok := data["statusCode"].(string); ok {
		if strings.TrimSpace(v) != "" {
			filter.StatusCode = &v
		}
	}
	if raw, ok := data["statusCodeIn"].([]interface{}); ok {
		for _, item := range raw {
			if sv, ok := item.(string); ok && strings.TrimSpace(sv) != "" {
				filter.StatusCodeIn = append(filter.StatusCodeIn, sv)
			}
		}
	}
	// Фильтрация по дате регистрации (диапазон).
	// Frontend отправляет поля registrationDateFrom / registrationDateTo в формате YYYY-MM-DD.
	if v, ok := data["registrationDateFrom"].(string); ok && strings.TrimSpace(v) != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.RegisteredAfter = model.NewDate(t)
		} else {
			agentLog("run-filters", "exec.go:parseCompanyFilter", "invalid registrationDateFrom value", map[string]interface{}{
				"raw": v,
				"err": err.Error(),
			})
		}
	}
	if v, ok := data["registrationDateTo"].(string); ok && strings.TrimSpace(v) != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.RegisteredBefore = model.NewDate(t)
		} else {
			agentLog("run-filters", "exec.go:parseCompanyFilter", "invalid registrationDateTo value", map[string]interface{}{
				"raw": v,
				"err": err.Error(),
			})
		}
	}
	agentLog("run-filters", "exec.go:parseCompanyFilter", "parsed company filter", map[string]interface{}{
		"hasRegionCode": filter.RegionCode != nil,
		"regionCode": func() string {
			if filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
		"hasOkved":      filter.Okved != nil,
		"hasStatus":     filter.Status != nil,
		"statusInLen":   len(filter.StatusIn),
		"hasFounderName": filter.FounderName != nil,
		"founderName": func() string {
			if filter.FounderName != nil {
				return *filter.FounderName
			}
			return ""
		}(),
	})
	return filter
}

func parseCompanySort(data map[string]interface{}) *model.CompanySort {
	sort := &model.CompanySort{}
	if v, ok := data["field"].(string); ok {
		sort.Field = model.CompanySortField(v)
	}
	if v, ok := data["order"].(string); ok {
		order := model.SortOrder(strings.ToUpper(v))
		sort.Order = &order
	}
	return sort
}

func parseEntrepreneurFilter(data map[string]interface{}) *model.EntrepreneurFilter {
	filter := &model.EntrepreneurFilter{}
	if v, ok := data["inn"].(string); ok {
		filter.Inn = &v
	}
	if v, ok := data["ogrnip"].(string); ok {
		filter.Ogrnip = &v
	}
	if v, ok := data["name"].(string); ok {
		filter.Name = &v
	}
	if v, ok := data["lastName"].(string); ok {
		filter.LastName = &v
	}
	if v, ok := data["regionCode"].(string); ok {
		filter.RegionCode = &v
		agentLog("run-filters", "exec.go:parseEntrepreneurFilter", "parsed regionCode for entrepreneurs", map[string]interface{}{"regionCode": v})
	}
	if v, ok := data["okved"].(string); ok {
		filter.Okved = &v
	}
	if v, ok := data["status"].(string); ok {
		status := model.EntityStatus(strings.ToUpper(v))
		if status.IsValid() {
			filter.Status = &status
		} else {
			agentLog("run-filters", "exec.go:parseEntrepreneurFilter", "invalid status value in entrepreneur filter", map[string]interface{}{
				"rawStatus": v,
			})
		}
	}
	if raw, ok := data["statusIn"].([]interface{}); ok {
		for _, item := range raw {
			if sv, ok := item.(string); ok {
				s := model.EntityStatus(strings.ToUpper(sv))
				if s.IsValid() {
					filter.StatusIn = append(filter.StatusIn, s)
				}
			}
		}
	}
	// Фильтрация по коду статуса (status_code) как строке
	if v, ok := data["statusCode"].(string); ok {
		if strings.TrimSpace(v) != "" {
			filter.StatusCode = &v
		}
	}
	if raw, ok := data["statusCodeIn"].([]interface{}); ok {
		for _, item := range raw {
			if sv, ok := item.(string); ok && strings.TrimSpace(sv) != "" {
				filter.StatusCodeIn = append(filter.StatusCodeIn, sv)
			}
		}
	}
	// Фильтрация по дате регистрации (диапазон) для ИП.
	// Используем те же ключи, что и для компаний: registrationDateFrom / registrationDateTo.
	if v, ok := data["registrationDateFrom"].(string); ok && strings.TrimSpace(v) != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.RegisteredAfter = model.NewDate(t)
		} else {
			agentLog("run-filters", "exec.go:parseEntrepreneurFilter", "invalid registrationDateFrom value", map[string]interface{}{
				"raw": v,
				"err": err.Error(),
			})
		}
	}
	if v, ok := data["registrationDateTo"].(string); ok && strings.TrimSpace(v) != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.RegisteredBefore = model.NewDate(t)
		} else {
			agentLog("run-filters", "exec.go:parseEntrepreneurFilter", "invalid registrationDateTo value", map[string]interface{}{
				"raw": v,
				"err": err.Error(),
			})
		}
	}
	agentLog("run-filters", "exec.go:parseEntrepreneurFilter", "parsed entrepreneur filter", map[string]interface{}{
		"hasRegionCode": filter.RegionCode != nil,
		"regionCode": func() string {
			if filter.RegionCode != nil {
				return *filter.RegionCode
			}
			return ""
		}(),
		"hasOkved":   filter.Okved != nil,
		"hasStatus":  filter.Status != nil,
		"statusInLen": len(filter.StatusIn),
	})
	return filter
}

func parseEntrepreneurSort(data map[string]interface{}) *model.EntrepreneurSort {
	sort := &model.EntrepreneurSort{}
	if v, ok := data["field"].(string); ok {
		field := model.EntrepreneurSortField(v)
		if field.IsValid() {
			sort.Field = field
		}
	}
	if v, ok := data["order"].(string); ok {
		order := model.SortOrder(strings.ToUpper(v))
		sort.Order = &order
	}
	return sort
}

func parsePagination(data map[string]interface{}) *model.Pagination {
	pagination := &model.Pagination{}
	if v, ok := data["limit"].(float64); ok {
		i := int(v)
		pagination.Limit = &i
	}
	if v, ok := data["offset"].(float64); ok {
		i := int(v)
		pagination.Offset = &i
	}
	return pagination
}

// extractPaginationFromQuery извлекает пагинацию из строки GraphQL запроса
// Ищет паттерн: pagination: {limit: 3, offset: 0}
func extractPaginationFromQuery(query string) *model.Pagination {
	pagination := &model.Pagination{}
	
	// Ищем "pagination:"
	paginationIdx := strings.Index(query, "pagination:")
	if paginationIdx == -1 {
		return nil
	}
	
	// Находим начало объекта { после pagination:
	start := paginationIdx + len("pagination:")
	for start < len(query) && (query[start] == ' ' || query[start] == '\n' || query[start] == '\t') {
		start++
	}
	if start >= len(query) || query[start] != '{' {
		return nil
	}
	
	// Ищем limit
	limitIdx := strings.Index(query[start:], "limit:")
	if limitIdx != -1 {
		limitStart := start + limitIdx + len("limit:")
		for limitStart < len(query) && (query[limitStart] == ' ' || query[limitStart] == '\n' || query[limitStart] == '\t') {
			limitStart++
		}
		// Парсим число
		var limitVal int
		if n, err := fmt.Sscanf(query[limitStart:], "%d", &limitVal); err == nil && n == 1 {
			pagination.Limit = &limitVal
		}
	}
	
	// Ищем offset
	offsetIdx := strings.Index(query[start:], "offset:")
	if offsetIdx != -1 {
		offsetStart := start + offsetIdx + len("offset:")
		for offsetStart < len(query) && (query[offsetStart] == ' ' || query[offsetStart] == '\n' || query[offsetStart] == '\t') {
			offsetStart++
		}
		// Парсим число
		var offsetVal int
		if n, err := fmt.Sscanf(query[offsetStart:], "%d", &offsetVal); err == nil && n == 1 {
			pagination.Offset = &offsetVal
		}
	}
	
	// Если нашли хотя бы одно значение, возвращаем пагинацию
	if pagination.Limit != nil || pagination.Offset != nil {
		return pagination
	}
	
	return nil
}

func parseStatsFilter(data map[string]interface{}) *model.StatsFilter {
	filter := &model.StatsFilter{}
	if v, ok := data["regionCode"].(string); ok {
		filter.RegionCode = &v
	}
	if v, ok := data["okved"].(string); ok {
		filter.Okved = &v
	}
	return filter
}

// NewDefaultExecutableSchema возвращает handler для совместимости
func NewDefaultExecutableSchema(resolver *Resolver) *ManualHandler {
	return NewManualHandler(resolver)
}
