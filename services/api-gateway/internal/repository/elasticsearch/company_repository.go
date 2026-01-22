package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

// ESCompanyRepository репозиторий для работы с компаниями через Elasticsearch
type ESCompanyRepository struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

// NewESCompanyRepository создает новый Elasticsearch репозиторий компаний
func NewESCompanyRepository(client *elasticsearch.Client, logger *zap.Logger) *ESCompanyRepository {
	return &ESCompanyRepository{
		client: client,
		logger: logger.Named("es_company_repo"),
	}
}

// Search выполняет полнотекстовый поиск компаний в Elasticsearch
func (r *ESCompanyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.Company, error) {
	// Построение multi-match запроса с boost
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					// Exact match по ОГРН (boost 100)
					{
						"term": map[string]interface{}{
							"ogrn": map[string]interface{}{
								"value": query,
								"boost": 100,
							},
						},
					},
					// Exact match по ИНН (boost 100)
					{
						"term": map[string]interface{}{
							"inn": map[string]interface{}{
								"value": query,
								"boost": 100,
							},
						},
					},
					// Морфологический поиск по полному наименованию (boost 10)
					{
						"match": map[string]interface{}{
							"full_name": map[string]interface{}{
								"query": query,
								"boost": 10,
							},
						},
					},
					// Морфологический поиск по краткому наименованию (boost 5)
					{
						"match": map[string]interface{}{
							"short_name": map[string]interface{}{
								"query": query,
								"boost": 5,
							},
						},
					},
					// Поиск по ФИО руководителя (boost 3)
					{
						"match": map[string]interface{}{
							"head_last_name": map[string]interface{}{
								"query": query,
								"boost": 3,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"head_first_name": map[string]interface{}{
								"query": query,
								"boost": 3,
							},
						},
					},
				},
			},
		},
		"from": offset,
		"size": limit,
	}

	// Сериализация запроса
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("encode search query: %w", err)
	}

	r.logger.Debug("Elasticsearch search query",
		zap.String("query", query),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.String("body", buf.String()))

	// Выполнение запроса
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("egrul_companies"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error",
			zap.String("status", res.Status()),
			zap.String("response", string(bodyBytes)))
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	// Парсинг ответа
	companies, err := r.parseSearchResponse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	r.logger.Info("Elasticsearch search completed",
		zap.String("query", query),
		zap.Int("results", len(companies)))

	return companies, nil
}

// SearchWithTotal выполняет поиск и возвращает компании с общим количеством найденных
func (r *ESCompanyRepository) SearchWithTotal(ctx context.Context, query string, limit, offset int) ([]*model.Company, int, error) {
	return r.SearchWithTotalAndFilters(ctx, query, nil, limit, offset)
}

// SearchWithTotalAndFilters выполняет поиск с фильтрами и возвращает компании с общим количеством
func (r *ESCompanyRepository) SearchWithTotalAndFilters(ctx context.Context, query string, filter *model.CompanyFilter, limit, offset int) ([]*model.Company, int, error) {
	// Build bool query with should (for text search) and must (for filters)
	boolQuery := map[string]interface{}{}

	// Если query задан, добавляем текстовый поиск
	if query != "" {
		boolQuery["should"] = []map[string]interface{}{
			// Exact match по ОГРН (boost 100)
			{
				"term": map[string]interface{}{
					"ogrn": map[string]interface{}{
						"value": query,
						"boost": 100,
					},
				},
			},
			// Exact match по ИНН (boost 100)
			{
				"term": map[string]interface{}{
					"inn": map[string]interface{}{
						"value": query,
						"boost": 100,
					},
				},
			},
			// Морфологический поиск по полному наименованию (boost 10)
			{
				"match": map[string]interface{}{
					"full_name": map[string]interface{}{
						"query": query,
						"boost": 10,
					},
				},
			},
			// Морфологический поиск по краткому наименованию (boost 5)
			{
				"match": map[string]interface{}{
					"short_name": map[string]interface{}{
						"query": query,
						"boost": 5,
					},
				},
			},
			// Поиск по ФИО руководителя (boost 3)
			{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"head_last_name", "head_first_name", "head_middle_name"},
					"boost":  3,
				},
			},
		}
		boolQuery["minimum_should_match"] = 1
	}

	// Add filters if provided
	if filter != nil {
		mustClauses := []map[string]interface{}{}

		if filter.RegionCode != nil && *filter.RegionCode != "" {
			mustClauses = append(mustClauses, map[string]interface{}{
				"term": map[string]interface{}{
					"region_code": *filter.RegionCode,
				},
			})
		}

		if filter.Status != nil {
			mustClauses = append(mustClauses, map[string]interface{}{
				"term": map[string]interface{}{
					// Статусы в ES хранятся в lowercase
					"status": strings.ToLower(string(*filter.Status)),
				},
			})
		}

		if filter.StatusIn != nil && len(filter.StatusIn) > 0 {
			statuses := make([]string, len(filter.StatusIn))
			for i, s := range filter.StatusIn {
				// Статусы в ES хранятся в lowercase
				statuses[i] = strings.ToLower(string(s))
			}
			mustClauses = append(mustClauses, map[string]interface{}{
				"terms": map[string]interface{}{
					"status": statuses,
				},
			})
		}

		if filter.Okved != nil && *filter.Okved != "" {
			mustClauses = append(mustClauses, map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"prefix": map[string]interface{}{
								"okved_main_code": *filter.Okved,
							},
						},
						{
							"prefix": map[string]interface{}{
								"okved_additional": *filter.Okved,
							},
						},
					},
					"minimum_should_match": 1,
				},
			})
		}

		// Date range filters for registration_date
		if filter.RegisteredAfter != nil || filter.RegisteredBefore != nil {
			rangeClause := map[string]interface{}{}
			if filter.RegisteredAfter != nil {
				rangeClause["gte"] = filter.RegisteredAfter.Time.Format("2006-01-02")
			}
			if filter.RegisteredBefore != nil {
				rangeClause["lte"] = filter.RegisteredBefore.Time.Format("2006-01-02")
			}
			mustClauses = append(mustClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"registration_date": rangeClause,
				},
			})
		}

		if len(mustClauses) > 0 {
			boolQuery["must"] = mustClauses
		}
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from": offset,
		"size": limit,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, 0, fmt.Errorf("encode search query: %w", err)
	}

	// Execute search
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("egrul_companies"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("elasticsearch search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		r.logger.Error("Elasticsearch search error",
			zap.String("status", res.Status()),
			zap.String("response", string(bodyBytes)))
		return nil, 0, fmt.Errorf("elasticsearch returned error: %s", res.Status())
	}

	// Parse response with total
	companies, total, err := r.parseSearchResponseWithTotal(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("parse search response: %w", err)
	}

	r.logger.Info("Elasticsearch search completed",
		zap.String("query", query),
		zap.Int("results", len(companies)),
		zap.Int("total", total))

	return companies, total, nil
}

// parseSearchResponseWithTotal парсит ответ Elasticsearch в модели Company с общим количеством
func (r *ESCompanyRepository) parseSearchResponseWithTotal(body io.Reader) ([]*model.Company, int, error) {
	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source esCompanyDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(body).Decode(&esResponse); err != nil {
		return nil, 0, fmt.Errorf("decode elasticsearch response: %w", err)
	}

	companies := make([]*model.Company, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		company := hit.Source.toModel()
		companies = append(companies, company)
	}

	return companies, esResponse.Hits.Total.Value, nil
}

// parseSearchResponse парсит ответ Elasticsearch в модели Company
func (r *ESCompanyRepository) parseSearchResponse(body io.Reader) ([]*model.Company, error) {
	companies, _, err := r.parseSearchResponseWithTotal(body)
	return companies, err
}

// esCompanyDocument структура документа компании в Elasticsearch
type esCompanyDocument struct {
	OGRN             string    `json:"ogrn"`
	INN              string    `json:"inn"`
	KPP              *string   `json:"kpp"`
	FullName         string    `json:"full_name"`
	ShortName        *string   `json:"short_name"`
	Status           string    `json:"status"`
	RegionCode       *string   `json:"region_code"`
	OkvedMainCode    *string   `json:"okved_main_code"`
	OkvedMainName    *string   `json:"okved_main_name"`
	OkvedAdditional  []string  `json:"okved_additional"`
	HeadLastName     *string   `json:"head_last_name"`
	HeadFirstName    *string   `json:"head_first_name"`
	HeadMiddleName   *string   `json:"head_middle_name"`
	RegistrationDate *string   `json:"registration_date"`
	UpdatedAt        string    `json:"updated_at"`
}

// toModel преобразует ES документ в модель GraphQL
func (doc *esCompanyDocument) toModel() *model.Company {
	company := &model.Company{
		Ogrn:       doc.OGRN,
		Inn:        doc.INN,
		FullName:   doc.FullName,
		Status:     model.ParseEntityStatus(doc.Status),
		Activities: make([]*model.Activity, 0),
	}

	// Nullable fields
	if doc.KPP != nil {
		company.Kpp = doc.KPP
	}
	if doc.ShortName != nil {
		company.ShortName = doc.ShortName
	}

	// Address
	if doc.RegionCode != nil {
		company.Address = &model.Address{
			RegionCode: doc.RegionCode,
		}
	}

	// Director
	if doc.HeadLastName != nil && doc.HeadFirstName != nil {
		company.Director = &model.Person{
			LastName:  *doc.HeadLastName,
			FirstName: *doc.HeadFirstName,
		}
		if doc.HeadMiddleName != nil {
			company.Director.MiddleName = doc.HeadMiddleName
		}
	}

	// Main activity
	if doc.OkvedMainCode != nil {
		company.MainActivity = &model.Activity{
			Code:   *doc.OkvedMainCode,
			IsMain: true,
		}
		if doc.OkvedMainName != nil {
			company.MainActivity.Name = doc.OkvedMainName
		}
		company.Activities = append(company.Activities, company.MainActivity)
	}

	// Additional activities
	for _, code := range doc.OkvedAdditional {
		activity := &model.Activity{
			Code:   code,
			IsMain: false,
		}
		company.Activities = append(company.Activities, activity)
	}

	// Dates
	if doc.RegistrationDate != nil {
		if t, err := time.Parse(time.RFC3339, *doc.RegistrationDate); err == nil {
			company.RegistrationDate = &model.Date{Time: t}
		}
	}
	if t, err := time.Parse(time.RFC3339, doc.UpdatedAt); err == nil {
		company.UpdatedAt = t
	}

	// Placeholder values для обязательных полей
	company.VersionDate = model.Date{Time: time.Now()}
	company.CreatedAt = time.Now()

	return company
}

// GetByOGRN не поддерживается в Elasticsearch репозитории
// Для точных совпадений используется ClickHouse
func (r *ESCompanyRepository) GetByOGRN(ctx context.Context, ogrn string) (*model.Company, error) {
	return nil, fmt.Errorf("GetByOGRN not supported in Elasticsearch repository, use ClickHouse instead")
}

// GetByINN не поддерживается в Elasticsearch репозитории
// Для точных совпадений используется ClickHouse
func (r *ESCompanyRepository) GetByINN(ctx context.Context, inn string) (*model.Company, error) {
	return nil, fmt.Errorf("GetByINN not supported in Elasticsearch repository, use ClickHouse instead")
}

// List не поддерживается в Elasticsearch репозитории
// Для фильтрации и пагинации используется ClickHouse
func (r *ESCompanyRepository) List(ctx context.Context, filter *model.CompanyFilter, pagination *model.Pagination, sort *model.CompanySort) ([]*model.Company, int, error) {
	return nil, 0, fmt.Errorf("List not supported in Elasticsearch repository, use ClickHouse instead")
}
